package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// Config is the root application configuration.
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"db"`
	Auth   AuthConfig   `mapstructure:"auth"`
	SMTP   SMTPConfig   `mapstructure:"smtp"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	// server.addr | WHAT2COOK_SERVER_ADDR
	Addr string `mapstructure:"addr"`
	// server.public_url | WHAT2COOK_SERVER_PUBLIC_URL
	PublicURL string `mapstructure:"public_url"`
}

// DBConfig holds database settings.
type DBConfig struct {
	// db.path | WHAT2COOK_DB_PATH
	Path string `mapstructure:"path"`
}

// AuthConfig holds authentication / session settings.
type AuthConfig struct {
	// auth.token_secret | WHAT2COOK_AUTH_TOKEN_SECRET
	TokenSecret string `mapstructure:"token_secret"`
	// auth.session_ttl | WHAT2COOK_AUTH_SESSION_TTL
	SessionTTL time.Duration `mapstructure:"session_ttl"`
	// auth.reset_ttl | WHAT2COOK_AUTH_RESET_TTL
	ResetTTL time.Duration `mapstructure:"reset_ttl"`
	// auth.verify_ttl | WHAT2COOK_AUTH_VERIFY_TTL
	VerifyTTL time.Duration `mapstructure:"verify_ttl"`
}

// SMTPConfig holds outbound email settings. Empty Host = log reset links instead of sending.
type SMTPConfig struct {
	// smtp.host | WHAT2COOK_SMTP_HOST
	Host string `mapstructure:"host"`
	// smtp.port | WHAT2COOK_SMTP_PORT
	Port int `mapstructure:"port"`
	// smtp.user | WHAT2COOK_SMTP_USER
	User string `mapstructure:"user"`
	// smtp.password | WHAT2COOK_SMTP_PASSWORD
	Password string `mapstructure:"password"`
	// smtp.from | WHAT2COOK_SMTP_FROM
	From string `mapstructure:"from"`
}

var global *Config

// Get returns the loaded configuration. Panics if Load was not called.
func Get() *Config {
	if global == nil {
		panic("config: not loaded")
	}
	return global
}

// Load reads configuration from file and environment variables.
// Env prefix is WHAT2COOK_; nested keys use underscores (server.addr → WHAT2COOK_SERVER_ADDR).
func Load(cfgFile string) error {
	v := viper.New()

	v.SetDefault("server.addr", ":8080")
	v.SetDefault("server.public_url", "http://localhost:8080")
	v.SetDefault("db.path", "what2cook.db")
	v.SetDefault("auth.token_secret", "change-me")
	v.SetDefault("auth.session_ttl", "168h")
	v.SetDefault("auth.reset_ttl", "1h")
	v.SetDefault("auth.verify_ttl", "24h")
	v.SetDefault("smtp.host", "")
	v.SetDefault("smtp.port", 587)
	v.SetDefault("smtp.user", "")
	v.SetDefault("smtp.password", "")
	v.SetDefault("smtp.from", "")

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
	}

	v.SetEnvPrefix("WHAT2COOK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		// Missing config file is OK when defaults / env are enough.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only ignore "not found" when using search paths; explicit file must exist.
			if cfgFile != "" {
				return fmt.Errorf("read config: %w", err)
			}
			// Non-not-found errors when searching also fail.
			if !isConfigFileNotFound(err) {
				return fmt.Errorf("read config: %w", err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	))); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	if strings.TrimSpace(cfg.Auth.TokenSecret) == "" {
		return fmt.Errorf("auth.token_secret must not be empty")
	}
	if cfg.Auth.SessionTTL <= 0 {
		return fmt.Errorf("auth.session_ttl must be positive")
	}
	if cfg.Auth.ResetTTL <= 0 {
		return fmt.Errorf("auth.reset_ttl must be positive")
	}
	if cfg.Auth.VerifyTTL <= 0 {
		return fmt.Errorf("auth.verify_ttl must be positive")
	}

	global = &cfg
	return nil
}

func isConfigFileNotFound(err error) bool {
	_, ok := err.(viper.ConfigFileNotFoundError)
	return ok
}
