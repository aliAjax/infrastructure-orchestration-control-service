package platform

import (
	"context"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the root application configuration.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Executor ExecutorConfig `yaml:"executor"`
	Drift    DriftConfig    `yaml:"drift"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
	HTTPAddr        string        `yaml:"http_addr"`
	GRPCAddr        string        `yaml:"grpc_addr"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseConfig struct {
	URL            string `yaml:"url"`
	MaxConnections int32  `yaml:"max_connections"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type ExecutorConfig struct {
	PollInterval time.Duration `yaml:"poll_interval"`
	TaskTimeout  time.Duration `yaml:"task_timeout"`
	MaxRetries   int           `yaml:"max_retries"`
}

type DriftConfig struct {
	Interval time.Duration `yaml:"interval"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

func LoadConfig(path string) (Config, error) {
	cfg := defaultConfig()
	if path == "" {
		path = os.Getenv("INFRA_CONFIG")
	}
	if path == "" {
		path = "configs/config.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("read config %q: %w", path, err)
		}
	} else if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}
	applyEnv(&cfg)
	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			HTTPAddr:        ":8080",
			GRPCAddr:        ":9090",
			ShutdownTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			URL:            "postgres://infra:infra@localhost:55432/infra_controlplane?sslmode=disable",
			MaxConnections: 20,
		},
		Redis: RedisConfig{
			Addr: "localhost:56379",
		},
		Executor: ExecutorConfig{
			PollInterval: 2 * time.Second,
			TaskTimeout:  30 * time.Second,
			MaxRetries:   3,
		},
		Drift: DriftConfig{
			Interval: 60 * time.Second,
		},
		Logging: LoggingConfig{Level: "info"},
	}
}

func applyEnv(cfg *Config) {
	setString(&cfg.Server.HTTPAddr, "INFRA_HTTP_ADDR")
	setString(&cfg.Server.GRPCAddr, "INFRA_GRPC_ADDR")
	setString(&cfg.Database.URL, "INFRA_DATABASE_URL")
	setString(&cfg.Redis.Addr, "INFRA_REDIS_ADDR")
	setString(&cfg.Redis.Password, "INFRA_REDIS_PASSWORD")
	setString(&cfg.Logging.Level, "INFRA_LOG_LEVEL")
}

func setString(target *string, key string) {
	if value := os.Getenv(key); value != "" {
		*target = value
	}
}

func (c Config) WithContext(ctx context.Context) context.Context {
	return ctx
}
