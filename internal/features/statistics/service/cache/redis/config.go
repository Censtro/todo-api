package statistics_cache

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr     string        `envconfig:"ADDR" default:"localhost:6379"`
	Password string        `envconfig:"PASSWORD"`
	DB       int           `envconfig:"DB"`
	TTL      time.Duration `envconfig:"TTL" default:"60s"`
}

func NewConfig() (Config, error) {
	var cfg Config
	if err := envconfig.Process("STATISTICS_CACHE_REDIS", &cfg); err != nil {
		return Config{}, fmt.Errorf("procces Redis cache config: %w", err)
	}

	if cfg.TTL <= 0 {
		return Config{}, fmt.Errorf("Redis cache TTL must be positive")
	}

	return cfg, nil
}

func NewConfigMust() Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}
