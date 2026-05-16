package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AndreeJait/go-utility/v2/logw"
	"github.com/AndreeJait/go-utility/v2/valuew"
	"github.com/spf13/viper"
)

// AppConfig holds all application configuration.
type AppConfig struct {
	App struct {
		Name     string `mapstructure:"name"`
		Env      string `mapstructure:"env"`
		HTTPPort int    `mapstructure:"http_port"`
	} `mapstructure:"app"`

	HTTP struct {
		Engine        string `mapstructure:"engine"`
		EnableSwagger bool   `mapstructure:"enable_swagger"`
		DebugMode     bool   `mapstructure:"debug_mode"`
		APIKey        string `mapstructure:"api_key"`
		SwaggerHost   string `mapstructure:"swagger_host"`
	} `mapstructure:"http"`

	Log struct {
		Level       string         `mapstructure:"level"`
		Format      logw.LogFormat `mapstructure:"format"`
		WriteToFile bool           `mapstructure:"write_to_file"`
		FilePath    string         `mapstructure:"file_path"`
	} `mapstructure:"log"`

	DB struct {
		Driver          string        `mapstructure:"driver"`
		Dialect         string        `mapstructure:"dialect"`
		DSN             string        `mapstructure:"dsn"`
		MaxOpenConns    int           `mapstructure:"max_open_conns"`
		MaxIdleConns    int           `mapstructure:"max_idle_conns"`
		ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
		DebugMode      bool          `mapstructure:"debug_mode"`
	} `mapstructure:"db"`

	Redis struct {
		Address   string `mapstructure:"address"`
		Password  string `mapstructure:"password"`
		DB        int    `mapstructure:"db"`
		PoolSize  int    `mapstructure:"pool_size"`
		DebugMode bool   `mapstructure:"debug_mode"`
	} `mapstructure:"redis"`

	MinIO struct {
		Endpoint      string `mapstructure:"endpoint"`
		AccessKey     string `mapstructure:"access_key"`
		SecretKey     string `mapstructure:"secret_key"`
		UseSSL        bool   `mapstructure:"use_ssl"`
		Region        string `mapstructure:"region"`
		ScriptsBucket string `mapstructure:"scripts_bucket"`
	} `mapstructure:"minio"`

	Embedding struct {
		BaseURL string `mapstructure:"base_url"`
		Model   string `mapstructure:"model"`
	} `mapstructure:"embedding"`

	LLM struct {
		BaseURL string `mapstructure:"base_url"`
		Model   string `mapstructure:"model"`
		APIKey  string `mapstructure:"api_key"`
	} `mapstructure:"llm"`

	Graceful struct {
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	} `mapstructure:"graceful"`
}

// Load reads the base config file at configPath, then merges app.local.yaml
// from the same directory if it exists. Environment variables override both.
func Load(configPath string) (*AppConfig, error) {
	v := viper.New()

	v.SetConfigFile(configPath)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Merge local overrides if app.local.yaml exists alongside the base config
	localPath := strings.Replace(configPath, "app.yaml", "app.local.yaml", 1)
	if _, err := os.Stat(localPath); err == nil {
		v.SetConfigFile(localPath)
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("failed to load local config: %w", err)
		}
	}

	cfg := &AppConfig{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply defaults
	cfg.HTTP.Engine = valuew.Coalesce(cfg.HTTP.Engine, "echo")
	cfg.App.HTTPPort = valuew.Coalesce(cfg.App.HTTPPort, 8080)
	cfg.DB.Driver = valuew.Coalesce(cfg.DB.Driver, "gorm")
	cfg.DB.Dialect = valuew.Coalesce(cfg.DB.Dialect, "postgres")
	cfg.MinIO.ScriptsBucket = valuew.Coalesce(cfg.MinIO.ScriptsBucket, "zora-scripts")
	cfg.Embedding.BaseURL = valuew.Coalesce(cfg.Embedding.BaseURL, "http://localhost:11434")
	cfg.Embedding.Model = valuew.Coalesce(cfg.Embedding.Model, "nomic-embed-text")
	cfg.LLM.BaseURL = valuew.Coalesce(cfg.LLM.BaseURL, "http://localhost:11434/v1")
	cfg.LLM.Model = valuew.Coalesce(cfg.LLM.Model, "qwen3:14b")

	cfg.Graceful.ShutdownTimeout = valuew.Coalesce(cfg.Graceful.ShutdownTimeout, 10*time.Second)

	return cfg, nil
}