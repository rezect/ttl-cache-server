package httpclient

import (
	"log"
	"time"
)

type Config struct {
	BaseURL    string        `json:"base_url"`
	Timeout    time.Duration `json:"timeout"`
	MaxRetries int           `json:"max_retries"`
	MinDelay   time.Duration `json:"min_delay"`
	MaxDelay   time.Duration `json:"max_delay"`
	Logger     *log.Logger   `json:"logger"`
}

func DefaultConfig() Config {
	return Config{
		BaseURL:    "https://api.github.com/users",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		MinDelay:   1 * time.Second,
		MaxDelay:   30 * time.Second,
		Logger:     log.Default(),
	}
}

func validateConfig(cfg *Config) {
	defaultConfig := DefaultConfig()
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultConfig.BaseURL
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultConfig.Timeout
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = defaultConfig.MaxRetries
	}
	if cfg.MinDelay == 0 {
		cfg.MinDelay = defaultConfig.MinDelay
	}
	if cfg.MaxDelay == 0 {
		cfg.MaxDelay = defaultConfig.MaxDelay
	}
	if cfg.Logger == nil {
		cfg.Logger = defaultConfig.Logger
	}
}