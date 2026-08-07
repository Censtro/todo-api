package core_postgres_pool

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type BaseConfig struct {
	Host     string `envconfig:"HOST" required:"true"`
	Port     string `envconfig:"PORT" default:"5432"`
	User     string `envconfig:"USER" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
	Database string `envconfig:"DB" required:"true"`
}

func NewConfig() (BaseConfig, error) {
	var cfg BaseConfig
	if err := envconfig.Process("POSTGRES", &cfg); err != nil {
		return BaseConfig{}, fmt.Errorf("process envconfig: %w", err)
	}
	return cfg, nil
}

func NewConfigMust() BaseConfig {
	cfg, err := NewConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}
