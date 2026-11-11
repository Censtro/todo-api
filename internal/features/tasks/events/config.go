package task_events

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Brokers       []string      `envconfig:"BROKERS" default:"localhost:9094"`
	Topic         string        `envconfig:"TOPIC" default:"todo.task-events.v1"`
	DeadLetter    string        `envconfig:"DEAD_LETTER_TOPIC" default:"todo.task-events.v1.dlq"`
	ConsumerGroup string        `envconfig:"CONSUMER_GROUP" default:"todo-task-audit-v1"`
	PollInterval  time.Duration `envconfig:"POLL_INTERVAL" default:"500ms"`
	LeaseDuration time.Duration `envconfig:"LEASE_DURATION" default:"30s"`
	BatchSize     int           `envconfig:"BATCH_SIZE" default:"50"`
}

func NewConfig() (Config, error) {
	var cfg Config
	if err := envconfig.Process("TASK_EVENTS", &cfg); err != nil {
		return Config{}, fmt.Errorf("process task events config: %w", err)
	}
	if len(cfg.Brokers) == 0 {
		return Config{}, fmt.Errorf("at least one Kafka broker is required")
	}
	if cfg.Topic == "" || cfg.DeadLetter == "" || cfg.ConsumerGroup == "" {
		return Config{}, fmt.Errorf("Kafka topic, dead-letter topic, and consumer group are required")
	}
	if cfg.PollInterval <= 0 || cfg.LeaseDuration <= 0 || cfg.BatchSize <= 0 {
		return Config{}, fmt.Errorf("Kafka poll interval, lease duration, and batch size must be positive")
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
