# Zora MCP Server

Tool registry and execution service for the [Zora](https://github.com/AndreeJait) agent platform. Manages script-based tools (Python, Go, Bash) stored in MinIO, executes them on demand via a sandbox subprocess, and exposes them through the MCP (Model Context Protocol) interface.

Part of the Zora ecosystem:

- **[go-utility](https://github.com/AndreeJait/go-utility)** — shared infrastructure wrappers (logging, HTTP, DB, Redis, auth, storage, etc.)
- **[zora-core](https://github.com/AndreeJait/zora-core)** — agent orchestration service (think-act-observe loop)
- **zora-mcp-server** (this repo) — tool registry and execution
- **[zora-knowledge](https://github.com/AndreeJait/zora-knowledge)** — knowledge ingestion and semantic search

## Architecture

Hexagonal (Ports & Adapters) with strict inward dependency: **adapters → ports → domain**.

```
cmd/
  http/                HTTP server entry point (wiring + DI)
domain/
  entity/              Tool, ToolExecution, Tag models
  error/               Domain errors (ErrToolNotFound, ErrInvalidEnvKey, etc.)
port/
  inbound/
    tool/              Tool CRUD use case interface + input DTOs
    mcp/               MCP use case interface (ListTools, CallTool)
    health/            Health check interface
  outbound/            ToolRepository, ScriptStorage, ScriptExecutor interfaces
usecase/               Business logic implementations
  tool.go              Tool CRUD + env validation + LLM-powered creation
  mcp.go               Tool listing and execution
  health.go            Health check
adapter/
  inbound/echo/        HTTP handlers (Echo v5)
  outbound/
    tool/              GORM + pgvector repository
    executor/          Sandbox subprocess executor
    minio/             MinIO script storage
config/                Configuration loading
files/
  config/              app.yaml + app.local.yaml
  migrations/          PostgreSQL migrations (pgvector)
  scripts/             Built-in tool scripts (Python)
```

## API

### Tool Management

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/tools` | Register a new tool |
| POST | `/api/v1/tools/from-prompt` | Create tool from natural language prompt (LLM) |
| GET | `/api/v1/tools/:id` | Get tool by ID |
| PUT | `/api/v1/tools/:id` | Update tool (partial) |
| DELETE | `/api/v1/tools/:id` | Delete tool |
| GET | `/api/v1/tools` | List tools (paginated) |
| POST | `/api/v1/tools/search` | Semantic search by embedding |

### Tool Tags

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/tags` | List all available tags |

### Tool Environment Variables

Each tool can store environment variables (keys must start with `MCP_`) that are injected into the subprocess at execution time.

| Method | Path | Description |
|--------|------|-------------|
| PUT | `/api/v1/tools/:id/env` | Add/update env vars (merges with existing) |
| DELETE | `/api/v1/tools/:id/env/:key` | Delete a specific env var |

**Example — set env vars:**
```bash
curl -X PUT http://localhost:8081/api/v1/tools/<id>/env \
  -H 'Content-Type: application/json' \
  -d '{"env": {"MCP_API_KEY": "sk-xxx", "MCP_DB_URL": "postgres://..."}}'
```

Scripts access these via `os.Getenv("MCP_API_KEY")` (Go), `os.environ["MCP_API_KEY"]` (Python), or `$MCP_API_KEY` (Bash).

### MCP Protocol

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/mcp/tools` | List active tools in MCP format |
| POST | `/api/v1/mcp/tools/call` | Execute a tool by name |

### Health

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Service health (DB + Redis connectivity) |

## Tool Execution Flow

```
zora-core (agent)
    │
    │  POST /api/v1/mcp/tools/call  {name, arguments}
    ▼
mcpUseCase.CallTool
    ├── repo.GetByName(name)
    ├── storage.GetScript(bucket, objectKey)     ← MinIO
    ├── executor.Execute(language, script, args, env)  ← Sandbox subprocess
    │       └── script runs with MCP_* env vars injected
    └── repo.SaveExecution(result)
```

Supported languages: `python`, `go`, `bash`. Scripts receive arguments as a JSON string via the first positional argument.

## Getting Started

### Prerequisites

- Go 1.26+
- PostgreSQL (with pgvector extension)
- Redis
- MinIO

### Setup

```bash
git clone https://github.com/AndreeJait/zora-mcp-server.git
cd zora-mcp-server

# Copy local config and customize DSN/Redis/MinIO
cp files/config/app.local.yaml.example files/config/app.local.yaml

# Run migrations
make migrate-up

# Run
make run
```

### Run

```bash
make run                      # Default engine (echo)
make run-engine E=gin         # Specific engine (gin|mux)
make build                    # Build binary to bin/server
```

### Run with Docker

```bash
docker compose -f deploy/docker-compose.yaml up --build
```

See `deploy/docker-compose.yaml` for environment variables. Services included:
- **zora-mcp-server** — port 8081
- **PostgreSQL** (pgvector/pg17) — port 5433
- **Redis** — port 6381
- **MinIO** — ports 9002 (API), 9003 (console)

## CI/CD

Pushes to `master` trigger the GitHub Actions workflow (`.github/workflows/deploy.yml`):

1. Build Docker image
2. Push to `ghcr.io/andreejait/zora-mcp-server:latest`
3. SSH into server via cloudflared
4. Run `deploy/deploy.sh` — login, pull, up, migrate, cleanup

## Configuration

Config files in `files/config/`:

| File | Purpose |
|------|---------|
| `app.yaml` | Base config (committed) |
| `app.local.yaml` | Local overrides (gitignored) |

**Override priority** (highest wins): environment variables → `app.local.yaml` → `app.yaml`

```yaml
app:
  name: zora-mcp-server
  env: development
  http_port: 8081

http:
  engine: echo
  enable_swagger: true
  debug_mode: true
  api_key: ""          # Set to protect management endpoints

log:
  level: debug
  format: JSON

db:
  driver: gorm
  dialect: postgres
  dsn: "postgres://zora:zora@localhost:5432/zora_mcp?sslmode=disable"

redis:
  address: "localhost:6379"
  db: 0

minio:
  endpoint: "localhost:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
  use_ssl: false
  scripts_bucket: "zora-scripts"

embedding:
  base_url: "http://localhost:11434"
  model: "nomic-embed-text"

llm:
  base_url: "http://localhost:11434/v1"
  model: "qwen3:14b"
  api_key: ""

graceful:
  shutdown_timeout: 10s
```

## Database Migrations

```bash
make migrate-new name=add_env_column   # Create new migration
make migrate-up                        # Run pending migrations
make migrate-down                     # Roll back last migration
make migrate-fresh                    # Drop all + re-run all
```

## Swagger

Swagger UI available at `http://localhost:8081/swagger/` when `http.enable_swagger: true`.

```bash
make swag    # Regenerate docs from annotations
```

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make run` | Run with default engine (echo) |
| `make run-engine E=gin` | Run with specific engine |
| `make build` | Build binary to `bin/server` |
| `make swag` | Generate Swagger docs |
| `make test` | Run all tests |
| `make vet` | Run static analysis |
| `make tidy` | Clean up dependencies |
| `make migrate-new name=foo` | Create new migration |
| `make migrate-up` | Run pending migrations |
| `make migrate-down` | Roll back last migration |
| `make migrate-fresh` | Drop all + re-run all |

## License

MIT
