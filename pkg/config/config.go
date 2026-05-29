package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

// Config holds the full application configuration loaded from config.yaml.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Pipefy   PipefyConfig   `mapstructure:"pipefy"`
	Logger   LoggerConfig   `mapstructure:"logger"`
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name        string      `mapstructure:"name"`
	Port        int         `mapstructure:"port"`
	Environment Environment `mapstructure:"environment"`
}

// DatabaseConfig holds PostgreSQL connection parameters.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
}

// DSN returns the PostgreSQL connection string for GORM.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// PipefyConfig holds Pipefy API settings.
// Behavior per environment:
//   - dev:     token ignored, fake adapter used
//   - staging: token required, simulate=true logs payload without sending
//   - prod:    token required, simulate must be false, mutations are sent
type PipefyConfig struct {
	Token    string `mapstructure:"token"`
	PipeID   string `mapstructure:"pipe_id"`
	Simulate bool   `mapstructure:"simulate"`
}

// LoggerConfig holds logging settings.
type LoggerConfig struct {
	// Level accepts: debug, info, warn, error
	Level string `mapstructure:"level"`
	// Format accepts: json, console
	Format string `mapstructure:"format"`
}

// Load reads the configuration file at the given path and returns a validated Config.
// If path is empty, it looks for "config.yaml" in the current directory.
//
// Environment variables override any config file value.
// Each binding follows the pattern PIPEFY_<SECTION>_<KEY> (uppercase).
// Examples:
//
//	PIPEFY_DATABASE_HOST=postgres   override database.host
//	PIPEFY_DATABASE_PORT=5432       override database.port
//	PIPEFY_APP_PORT=9090            override app.port
//	PIPEFY_TOKEN=mytoken            override pipefy.token
func Load(path string) (*Config, error) {
	v := viper.New()

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Explicit env var bindings — AutomaticEnv alone does not work with
	// Unmarshal for nested keys in Viper; BindEnv is required per key.
	envBindings := map[string]string{
		"app.port":          "PIPEFY_APP_PORT",
		"app.environment":   "PIPEFY_APP_ENVIRONMENT",
		"database.host":     "PIPEFY_DATABASE_HOST",
		"database.port":     "PIPEFY_DATABASE_PORT",
		"database.name":     "PIPEFY_DATABASE_NAME",
		"database.user":     "PIPEFY_DATABASE_USER",
		"database.password": "PIPEFY_DATABASE_PASSWORD",
		"database.sslmode":  "PIPEFY_DATABASE_SSLMODE",
		"pipefy.token":      "PIPEFY_TOKEN",
		"pipefy.pipe_id":    "PIPEFY_PIPE_ID",
		"pipefy.simulate":   "PIPEFY_SIMULATE",
		"logger.level":      "PIPEFY_LOG_LEVEL",
		"logger.format":     "PIPEFY_LOG_FORMAT",
	}
	for key, env := range envBindings {
		if err := v.BindEnv(key, env); err != nil {
			return nil, fmt.Errorf("binding env %s: %w", env, err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate enforces per-environment rules and required fields.
func validate(cfg *Config) error {
	var errs []error

	if !cfg.App.Environment.IsValid() {
		errs = append(errs, fmt.Errorf(
			"app.environment %q is invalid: accepted values are dev, staging, prod",
			cfg.App.Environment,
		))
	}

	if cfg.App.Port <= 0 {
		errs = append(errs, errors.New("app.port must be a positive integer"))
	}

	if cfg.Database.Host == "" {
		errs = append(errs, errors.New("database.host is required"))
	}

	if cfg.Database.Port <= 0 {
		errs = append(errs, errors.New("database.port must be a positive integer"))
	}

	if cfg.Database.Name == "" {
		errs = append(errs, errors.New("database.name is required"))
	}

	if cfg.Database.User == "" {
		errs = append(errs, errors.New("database.user is required"))
	}

	// Token is required for staging and prod.
	if cfg.App.Environment.RequiresToken() && cfg.Pipefy.Token == "" {
		errs = append(errs, fmt.Errorf(
			"pipefy.token is required for environment %q",
			cfg.App.Environment,
		))
	}

	// Simulate is not allowed in prod: all mutations must be real.
	if cfg.App.Environment == Prod && cfg.Pipefy.Simulate {
		errs = append(errs, errors.New(
			"pipefy.simulate must be false in prod environment",
		))
	}

	return errors.Join(errs...)
}
