package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	core_logger "github.com/Censtro/todo-api/internal/core/logger"
	core_pgx_pool "github.com/Censtro/todo-api/internal/core/repository/postgres/pool/pgx"
	task_events "github.com/Censtro/todo-api/internal/features/tasks/events"
	task_events_kafka "github.com/Censtro/todo-api/internal/features/tasks/events/kafka"
	tasks_postgres_repository "github.com/Censtro/todo-api/internal/features/tasks/repository/postgres"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	log, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to initialize logger:", err)
		os.Exit(1)
	}
	defer log.Close()

	pool, err := core_pgx_pool.NewConnectionPool(ctx, core_pgx_pool.NewConfigMust())
	if err != nil {
		log.Fatal("failed to initialize PostgreSQL pool", zap.Error(err))
	}
	defer pool.Close()

	config := task_events.NewConfigMust()
	repository := tasks_postgres_repository.NewTaskRepository(pool)
	consumer := task_events_kafka.NewConsumer(config, repository, log.Logger)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Error("failed to close Kafka consumer", zap.Error(err))
		}
	}()

	log.Info("starting task audit consumer", zap.Strings("brokers", config.Brokers), zap.String("topic", config.Topic))
	if err := consumer.Run(ctx); err != nil {
		log.Fatal("task audit consumer stopped", zap.Error(err))
	}
}
