package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AndreeJait/go-utility/v2/logw"
	"github.com/AndreeJait/go-utility/v2/valuew"
	"github.com/AndreeJait/zora-mcp-server/config"
	domainEntity "github.com/AndreeJait/zora-mcp-server/domain/entity"
	domainError "github.com/AndreeJait/zora-mcp-server/domain/error"
	"github.com/AndreeJait/zora-mcp-server/port/inbound/tool"
	"github.com/AndreeJait/zora-mcp-server/port/outbound"
	"github.com/pgvector/pgvector-go"
)

type toolUseCase struct {
	cfg           *config.AppConfig
	repo          outbound.ToolRepository
	tagRepo       outbound.TagRepository
	embedProvider outbound.EmbeddingProvider
	llmProvider   outbound.LLMProvider
	storage       outbound.Storage
}

var _ tool.UseCase = (*toolUseCase)(nil)

func NewToolUseCase(cfg *config.AppConfig, repo outbound.ToolRepository, tagRepo outbound.TagRepository, embedProvider outbound.EmbeddingProvider, llmProvider outbound.LLMProvider, storage outbound.Storage) tool.UseCase {
	return &toolUseCase{cfg: cfg, repo: repo, tagRepo: tagRepo, embedProvider: embedProvider, llmProvider: llmProvider, storage: storage}
}

// resolveTagNames takes a list of tag names, generates additional tags via LLM,
// merges them, and persists them via find-or-create. Returns tag IDs and tag names.
func (uc *toolUseCase) resolveTagNames(ctx context.Context, userTags []string, name, description string) (tagIDs []string, tagNames []string, err error) {
	// Generate tags from LLM and merge with user-provided tags
	allCandidates := uc.generateTags(ctx, name, description)

	// Merge user-provided tag names with LLM-generated ones
	seen := make(map[string]bool)
	var merged []tagCandidate

	for _, name := range userTags {
		lower := strings.ToLower(strings.TrimSpace(name))
		if lower != "" && !seen[lower] {
			seen[lower] = true
			merged = append(merged, tagCandidate{Name: lower})
		}
	}
	for _, tc := range allCandidates {
		lower := strings.ToLower(strings.TrimSpace(tc.Name))
		if lower != "" && !seen[lower] {
			seen[lower] = true
			merged = append(merged, tc)
		}
	}

	if len(merged) == 0 {
		return nil, nil, nil
	}

	// Build names list and descriptions map for find-or-create
	names := make([]string, 0, len(merged))
	descriptions := make(map[string]string)
	for _, tc := range merged {
		names = append(names, tc.Name)
		if tc.Description != "" {
			descriptions[tc.Name] = tc.Description
		}
	}

	// Find or create tags in the database
	tags, err := uc.tagRepo.FindOrCreateByNames(ctx, names, descriptions)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve tags: %w", err)
	}

	tagIDs = make([]string, len(tags))
	tagNames = make([]string, len(tags))
	for i, tg := range tags {
		tagIDs[i] = tg.ID
		tagNames[i] = tg.Name
	}
	return tagIDs, tagNames, nil
}

func (uc *toolUseCase) Register(ctx context.Context, input tool.RegisterInput) (*domainEntity.Tool, error) {
	existing, _ := uc.repo.GetByName(ctx, input.Name)
	if existing != nil {
		return nil, domainError.ErrToolNameExists
	}

	if input.Language != "python" && input.Language != "go" && input.Language != "bash" {
		return nil, domainError.ErrInvalidLanguage
	}

	if err := validateEnvKeys(input.Env); err != nil {
		return nil, err
	}

	t := &domainEntity.Tool{
		Name:        input.Name,
		Description: input.Description,
		Language:    input.Language,
		ObjectKey:   input.ObjectKey,
		Bucket:      input.Bucket,
		Parameters:  input.Parameters,
		Env:         input.Env,
		CreatedBy:   input.CreatedBy,
		Metadata:    input.Metadata,
	}
	if t.Bucket == "" {
		t.Bucket = uc.cfg.MinIO.ScriptsBucket
	}
	if t.Parameters == nil {
		t.Parameters = make(map[string]any)
	}
	if t.Env == nil {
		t.Env = make(map[string]string)
	}
	if t.Metadata == nil {
		t.Metadata = make(map[string]any)
	}

	// Resolve tags: find-or-create, merging user-provided with LLM-generated
	tagIDs, tagNames, err := uc.resolveTagNames(ctx, input.Tags, t.Name, t.Description)
	if err != nil {
		return nil, err
	}

	// Generate embedding using LLM-enriched description for better search accuracy
	embedText := uc.generateEmbedText(ctx, t.Name, t.Description, t.Parameters, tagNames)
	if emb, err := uc.embedProvider.Embed(ctx, embedText); err == nil {
		t.Embedding = pgvector.NewVector(float64sTo32(emb))
	} else {
		logw.CtxWarningf(ctx, "toolUseCase: failed to generate embedding for %s: %v", t.Name, err)
	}

	if err := uc.repo.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("register tool: %w", err)
	}

	// Save tag associations
	if len(tagIDs) > 0 {
		if err := uc.repo.ReplaceToolTags(ctx, t.ID, tagIDs); err != nil {
			return nil, fmt.Errorf("save tool tags: %w", err)
		}
	}

	// Reload the tool with tags populated
	t, _ = uc.repo.GetByID(ctx, t.ID)

	logw.CtxInfof(ctx, "toolUseCase: registered tool %s (language=%s)", t.Name, t.Language)
	return t, nil
}

const createWithPromptSystem = `You are a tool generator for an MCP (Model Context Protocol) server. Given a user's description of what they want a tool to do, generate a complete tool definition.

You MUST respond with ONLY valid JSON (no markdown, no explanation) in this exact format:
{
  "name": "short-kebab-case-name",
  "description": "Clear one-sentence description of what the tool does and when to use it.",
  "parameters": {
    "type": "object",
    "required": ["param1"],
    "properties": {
      "param1": {"type": "string", "description": "Description of param1"},
      "param2": {"type": "string", "description": "Description of param2"}
    }
  },
  "script": "complete Python script code",
  "tags": [{"name": "tag1", "description": "What this tag represents"}, {"name": "tag2", "description": "What this tag represents"}]
}

Rules:
- name: short, kebab-case (e.g. "analyze-code", "generate-qr"). No spaces.
- description: one clear sentence explaining what the tool does and when to use it.
- parameters: valid JSON Schema object with "type", "required", and "properties". Each property must have a "description".
- script: a complete, runnable Python script that:
  - Reads arguments from sys.argv[1] as a JSON string
  - Outputs results as JSON to stdout
  - Uses only Python standard library (no pip packages)
  - If the tool needs LLM access, reads MCP_LLM_BASE_URL, MCP_LLM_API_KEY, MCP_LLM_MODEL from environment variables
  - If the tool needs HTTP, uses urllib.request (no requests library)
  - Handles errors gracefully and outputs {"error": "..."} on failure
  - Starts with #!/usr/bin/env python3
- tags: an array of 2-5 tag objects, each with "name" (lowercase, single word or short hyphenated phrase) and "description" (a short sentence explaining what the tag represents)`

const embedTextSystem = `You are a tool search optimizer. Given a tool's name, description, and parameters, generate a rich, detailed description optimized for semantic search and vector embedding.

Your goal is to make this tool easily discoverable by semantic search queries. Write 2-4 sentences that:
1. Explain what the tool does in detail, including its domain and use cases
2. Mention related concepts, synonyms, and alternative phrasings people might search for
3. Distinguish this tool from similar tools by highlighting what makes it unique

Respond with ONLY the description text, no JSON, no markdown, no labels.`

type llmToolResponse struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Parameters  map[string]any   `json:"parameters"`
	Script      string           `json:"script"`
	Tags        []tagCandidate   `json:"tags"`
}

// tagCandidate holds a tag name and its LLM-generated description.
type tagCandidate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// generateEmbedText produces an LLM-enriched description for embedding.
// Falls back to buildEmbedText if the LLM call fails.
func (uc *toolUseCase) generateEmbedText(ctx context.Context, name, description string, parameters map[string]any, tagNames []string) string {
	baseText := buildEmbedText(name, description, parameters, tagNames)

	userPrompt := fmt.Sprintf("Tool name: %s\nDescription: %s", name, description)

	// Add tags info to the prompt
	if len(tagNames) > 0 {
		userPrompt += fmt.Sprintf("\nTags: %s", strings.Join(tagNames, ", "))
	}

	// Add parameter info to the prompt
	if properties, ok := parameters["properties"].(map[string]any); ok && len(properties) > 0 {
		userPrompt += "\nParameters:"
		for paramName, paramDef := range properties {
			if pm, ok := paramDef.(map[string]any); ok {
				if desc, ok := pm["description"].(string); ok {
					userPrompt += fmt.Sprintf("\n- %s: %s", paramName, desc)
				}
			}
		}
	}

	enriched, err := uc.llmProvider.Generate(ctx, embedTextSystem, userPrompt)
	if err != nil {
		logw.CtxWarningf(ctx, "toolUseCase: LLM embed text generation failed for %s, using fallback: %v", name, err)
		return baseText
	}

	enriched = strings.TrimSpace(enriched)
	if enriched == "" {
		return baseText
	}

	// Strip markdown fences if the LLM wrapped the response
	if strings.HasPrefix(enriched, "```") {
		if idx := strings.Index(enriched, "\n"); idx != -1 {
			enriched = enriched[idx+1:]
		}
		if idx := strings.LastIndex(enriched, "```"); idx != -1 {
			enriched = enriched[:idx]
		}
		enriched = strings.TrimSpace(enriched)
	}

	return enriched
}

func (uc *toolUseCase) CreateWithPrompt(ctx context.Context, input tool.CreateWithPromptInput) (*domainEntity.Tool, error) {
	// Generate tool definition from LLM
	raw, err := uc.llmProvider.Generate(ctx, createWithPromptSystem, input.Prompt)
	if err != nil {
		return nil, fmt.Errorf("generate tool from prompt: %w", err)
	}

	// Strip markdown code fences if present
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		// Remove opening fence (e.g. ```json or ```)
		if idx := strings.Index(raw, "\n"); idx != -1 {
			raw = raw[idx+1:]
		}
		// Remove closing fence
		if idx := strings.LastIndex(raw, "```"); idx != -1 {
			raw = raw[:idx]
		}
		raw = strings.TrimSpace(raw)
	}

	var resp llmToolResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, fmt.Errorf("parse LLM response as tool definition: %w\nraw: %s", err, raw)
	}

	// Validate required fields
	if resp.Name == "" {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("LLM did not generate a tool name")
	}
	if resp.Description == "" {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("LLM did not generate a tool description")
	}
	if resp.Script == "" {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("LLM did not generate a script")
	}

	// Check name uniqueness
	existing, _ := uc.repo.GetByName(ctx, resp.Name)
	if existing != nil {
		return nil, domainError.ErrToolNameExists.WithCustomMessage(
			fmt.Sprintf("tool name '%s' already exists", resp.Name),
		)
	}

	// Sanitize name: only lowercase letters, digits, and hyphens
	sanitizedName := sanitizeToolName(resp.Name)

	// Build object key
	objectKey := "tools/" + sanitizedName + ".py"
	bucket := uc.cfg.MinIO.ScriptsBucket

	// Upload script to MinIO
	reader := bytes.NewReader([]byte(resp.Script))
	if err := uc.storage.Upload(ctx, bucket, objectKey, reader, int64(len(resp.Script)), "text/x-python"); err != nil {
		return nil, fmt.Errorf("upload script to storage: %w", err)
	}

	// Default parameters if not provided
	parameters := resp.Parameters
	if parameters == nil {
		parameters = make(map[string]any)
	}

	// Collect tag names from LLM response
	var tagNames []string
	for _, tc := range resp.Tags {
		tagNames = append(tagNames, tc.Name)
	}

	registerInput := tool.RegisterInput{
		Name:        sanitizedName,
		Description: resp.Description,
		Language:    "python",
		ObjectKey:   objectKey,
		Bucket:      bucket,
		Parameters:  parameters,
		Env:         map[string]string{},
		Tags:        tagNames,
		CreatedBy:   valuew.Coalesce(input.CreatedBy, "prompt"),
		Metadata:    map[string]any{"prompt": input.Prompt},
	}

	result, err := uc.Register(ctx, registerInput)
	if err != nil {
		return nil, fmt.Errorf("register generated tool: %w", err)
	}

	logw.CtxInfof(ctx, "toolUseCase: created tool from prompt: %s", result.Name)
	return result, nil
}

// sanitizeToolName converts a name to a valid kebab-case tool name.
func sanitizeToolName(name string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash {
			b.WriteRune('-')
			prevDash = true
		}
	}
	result := b.String()
	result = strings.Trim(result, "-")
	if result == "" {
		result = "tool"
	}
	return result
}

func (uc *toolUseCase) Update(ctx context.Context, id string, input tool.UpdateInput) (*domainEntity.Tool, error) {
	t, err := uc.repo.GetByID(ctx, id)
	if err != nil || t == nil {
		return nil, domainError.ErrToolNotFound
	}

	needReembed := false
	var userTagNames []string

	if input.Description != nil {
		t.Description = *input.Description
		needReembed = true
	}
	if input.ObjectKey != nil {
		t.ObjectKey = *input.ObjectKey
	}
	if input.Bucket != nil {
		t.Bucket = *input.Bucket
	}
	if input.Parameters != nil {
		t.Parameters = input.Parameters
		needReembed = true
	}
	if input.Env != nil {
		if err := validateEnvKeys(input.Env); err != nil {
			return nil, err
		}
		t.Env = input.Env
	}
	if input.IsActive != nil {
		t.IsActive = *input.IsActive
	}
	if input.Metadata != nil {
		t.Metadata = input.Metadata
	}
	if input.Tags != nil {
		userTagNames = *input.Tags
		needReembed = true
	}

	// Resolve tags: find-or-create, merging user-provided with LLM-generated
	tagIDs, tagNames, err := uc.resolveTagNames(ctx, userTagNames, t.Name, t.Description)
	if err != nil {
		return nil, err
	}
	if len(tagIDs) > 0 {
		needReembed = true
	}

	t.Version++

	// Re-generate embedding if description, parameters, or tags changed
	if needReembed {
		embedText := uc.generateEmbedText(ctx, t.Name, t.Description, t.Parameters, tagNames)
		if emb, err := uc.embedProvider.Embed(ctx, embedText); err == nil {
			t.Embedding = pgvector.NewVector(float64sTo32(emb))
		} else {
			logw.CtxWarningf(ctx, "toolUseCase: failed to re-generate embedding for %s: %v", t.Name, err)
		}
	}

	if err := uc.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("update tool: %w", err)
	}

	// Save tag associations
	if len(tagIDs) > 0 {
		if err := uc.repo.ReplaceToolTags(ctx, t.ID, tagIDs); err != nil {
			return nil, fmt.Errorf("save tool tags: %w", err)
		}
	}

	// Reload the tool with tags populated
	t, _ = uc.repo.GetByID(ctx, t.ID)

	logw.CtxInfof(ctx, "toolUseCase: updated tool %s (v%d)", t.Name, t.Version)
	return t, nil
}

func (uc *toolUseCase) Delete(ctx context.Context, id string) error {
	t, err := uc.repo.GetByID(ctx, id)
	if err != nil || t == nil {
		return domainError.ErrToolNotFound
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete tool: %w", err)
	}

	logw.CtxInfof(ctx, "toolUseCase: deleted tool %s", t.Name)
	return nil
}

func (uc *toolUseCase) GetByID(ctx context.Context, id string) (*domainEntity.Tool, error) {
	t, err := uc.repo.GetByID(ctx, id)
	if err != nil || t == nil {
		return nil, domainError.ErrToolNotFound
	}
	return t, nil
}

func (uc *toolUseCase) List(ctx context.Context, page, perPage int) ([]domainEntity.Tool, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	return uc.repo.List(ctx, page, perPage)
}

func (uc *toolUseCase) SearchTools(ctx context.Context, input tool.SearchInput) ([]tool.ToolSearchResult, error) {
	limit := input.Limit
	if limit == 0 {
		limit = 15
	}

	// If a text query is provided instead of an embedding, generate the embedding
	embedding := input.Embedding
	if len(embedding) == 0 && input.Query != "" {
		emb, err := uc.embedProvider.Embed(ctx, input.Query)
		if err != nil {
			return nil, fmt.Errorf("embed search query: %w", err)
		}
		embedding = emb
	}

	if len(embedding) == 0 {
		return nil, fmt.Errorf("search requires either 'query' text or 'embedding' vector")
	}

	tools, err := uc.repo.SearchByEmbedding(ctx, embedding, limit, outbound.ToolSearchFilter{
		Tags:   input.Tags,
		UserID: input.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("search tools: %w", err)
	}

	results := make([]tool.ToolSearchResult, 0, len(tools))
	for _, t := range tools {
		tags := make([]tool.TagDTO, 0, len(t.Tags))
		for _, tg := range t.Tags {
			tags = append(tags, tool.TagDTO{Name: tg.Name, Description: tg.Description})
		}
		results = append(results, tool.ToolSearchResult{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Language:    t.Language,
			Parameters:  t.Parameters,
			Score:       t.Score,
			Tags:        tags,
		})
	}
	return results, nil
}

// validateEnvKeys ensures all environment variable keys start with MCP_.
func validateEnvKeys(env map[string]string) error {
	for k := range env {
		if !strings.HasPrefix(k, "MCP_") {
			return domainError.ErrInvalidEnvKey.WithCustomMessage(
				fmt.Sprintf("environment variable key '%s' must start with MCP_", k),
			)
		}
	}
	return nil
}

func (uc *toolUseCase) SetEnv(ctx context.Context, id string, input tool.SetEnvInput) (*domainEntity.Tool, error) {
	if err := validateEnvKeys(input.Env); err != nil {
		return nil, err
	}

	t, err := uc.repo.GetByID(ctx, id)
	if err != nil || t == nil {
		return nil, domainError.ErrToolNotFound
	}

	if t.Env == nil {
		t.Env = make(map[string]string)
	}
	for k, v := range input.Env {
		t.Env[k] = v
	}
	t.Version++

	if err := uc.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("set env: %w", err)
	}

	logw.CtxInfof(ctx, "toolUseCase: set env on tool %s (v%d, %d keys)", t.Name, t.Version, len(t.Env))
	return t, nil
}

func (uc *toolUseCase) DeleteEnv(ctx context.Context, id, key string) (*domainEntity.Tool, error) {
	t, err := uc.repo.GetByID(ctx, id)
	if err != nil || t == nil {
		return nil, domainError.ErrToolNotFound
	}

	if t.Env == nil {
		t.Env = make(map[string]string)
	}
	delete(t.Env, key)
	t.Version++

	if err := uc.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("delete env: %w", err)
	}

	logw.CtxInfof(ctx, "toolUseCase: deleted env key %s from tool %s (v%d)", key, t.Name, t.Version)
	return t, nil
}

func (uc *toolUseCase) ListTags(ctx context.Context) ([]tool.TagOutput, error) {
	tags, err := uc.tagRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	result := make([]tool.TagOutput, len(tags))
	for i, t := range tags {
		result[i] = tool.TagOutput{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
		}
	}
	return result, nil
}

func float64sTo32(in []float64) []float32 {
	out := make([]float32, len(in))
	for i, v := range in {
		out[i] = float32(v)
	}
	return out
}

// buildEmbedText constructs the text used for embedding generation.
// It includes the tool name, description, and parameter descriptions
// to produce more differentiated vectors for semantic search.
// This is used as a fallback when the LLM is unavailable.
func buildEmbedText(name, description string, parameters map[string]any, tags []string) string {
	var sb strings.Builder
	sb.WriteString(name)
	sb.WriteString(": ")
	sb.WriteString(description)

	if len(tags) > 0 {
		sb.WriteString(". Tags: ")
		for i, tag := range tags {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(tag)
		}
	}

	if properties, ok := parameters["properties"].(map[string]any); ok && len(properties) > 0 {
		sb.WriteString(". Parameters: ")
		first := true
		for paramName, paramDef := range properties {
			if !first {
				sb.WriteString(", ")
			}
			first = false
			sb.WriteString(paramName)
			if pm, ok := paramDef.(map[string]any); ok {
				if desc, ok := pm["description"].(string); ok && desc != "" {
					sb.WriteString(" - ")
					sb.WriteString(desc)
				}
			}
		}
	}

	return sb.String()
}

const generateTagsSystem = `You are a tool tagging assistant. Given a tool's name and description, generate relevant tags that help categorize and discover the tool.

Respond with ONLY a JSON array of objects, no explanation, no markdown. Example: [{"name": "text", "description": "Tools for processing and manipulating text data"}, {"name": "analysis", "description": "Tools that analyze input data to produce insights"}]

Rules:
- Generate 2-5 tags
- Each tag has a "name" (lowercase, single word or short hyphenated phrase) and a "description" (a short sentence explaining what the tag represents)
- Include tags for: domain, purpose, input/output type, and related concepts
- Be specific but not overly narrow`

// generateTags uses the LLM to generate tags for a tool based on its name and description.
// Returns tag candidates with descriptions, or nil if generation fails.
func (uc *toolUseCase) generateTags(ctx context.Context, name, description string) []tagCandidate {
	userPrompt := fmt.Sprintf("Tool name: %s\nDescription: %s", name, description)

	raw, err := uc.llmProvider.Generate(ctx, generateTagsSystem, userPrompt)
	if err != nil {
		logw.CtxWarningf(ctx, "toolUseCase: LLM tag generation failed for %s: %v", name, err)
		return nil
	}

	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		if idx := strings.Index(raw, "\n"); idx != -1 {
			raw = raw[idx+1:]
		}
		if idx := strings.LastIndex(raw, "```"); idx != -1 {
			raw = raw[:idx]
		}
		raw = strings.TrimSpace(raw)
	}

	var candidates []tagCandidate
	if err := json.Unmarshal([]byte(raw), &candidates); err != nil {
		logw.CtxWarningf(ctx, "toolUseCase: failed to parse LLM tags for %s: %v", name, err)
		return nil
	}

	return candidates
}