package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_logger "github.com/Censtro/todo-api/internal/core/logger"
	core_pgx_pool "github.com/Censtro/todo-api/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "github.com/Censtro/todo-api/internal/core/transport/http/middleware"
	core_http_server "github.com/Censtro/todo-api/internal/core/transport/http/server"
	statistics_postgres_repository "github.com/Censtro/todo-api/internal/features/statistics/repository/postgres"
	statistics_service "github.com/Censtro/todo-api/internal/features/statistics/service"
	statistics_cache "github.com/Censtro/todo-api/internal/features/statistics/service/cache/redis"
	statistics_transport_http "github.com/Censtro/todo-api/internal/features/statistics/transport/http"
	task_events "github.com/Censtro/todo-api/internal/features/tasks/events"
	task_events_kafka "github.com/Censtro/todo-api/internal/features/tasks/events/kafka"
	tasks_postgres_repository "github.com/Censtro/todo-api/internal/features/tasks/repository/postgres"
	task_service "github.com/Censtro/todo-api/internal/features/tasks/service"
	tasks_transport_http "github.com/Censtro/todo-api/internal/features/tasks/transport/http"
	user_postgres_repository "github.com/Censtro/todo-api/internal/features/users/repository/postgres"
	users_service "github.com/Censtro/todo-api/internal/features/users/service"
	user_transport_http "github.com/Censtro/todo-api/internal/features/users/transport/http"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	_ "github.com/Censtro/todo-api/docs"
)

// @title Todo API
// @version 1.0
// @description Todo application REST API scheme
// @host 127.0.0.1:5050
// @BasePath /api/v1
func main() {
	timeZone := time.UTC

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	log, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init logger: ", err)
		os.Exit(1)
	}
	defer log.Close()
	log.Debug("start app")

	time.Local = timeZone
	log.Debug("Server time zone", zap.Any("zone", timeZone))

	log.Debug("postgres pool initializing")

	pool, err := core_pgx_pool.NewConnectionPool(
		ctx,
		core_pgx_pool.NewConfigMust(),
	)

	if err != nil {
		log.Fatal("failed to init postgres connection pool", zap.Error(err))
	}
	defer pool.Close()

	usersRepository := user_postgres_repository.NewUserRepository(pool)
	usersService := users_service.NewUserService(usersRepository)
	usertransport := user_transport_http.NewUserHTTPHandler(usersService)

	taskRepository := tasks_postgres_repository.NewTaskRepository(pool)
	tasksService := task_service.NewTasksService(taskRepository)
	taskTransport := tasks_transport_http.NewTasksHTTPHandler(tasksService)

	taskEventsConfig := task_events.NewConfigMust()
	eventProducer := task_events_kafka.NewProducer(taskEventsConfig.Brokers, taskEventsConfig.Topic)
	defer func() {
		if err := eventProducer.Close(); err != nil {
			log.Error("failed to close Kafka producer", zap.Error(err))
		}
	}()
	outboxPublisher := task_events.NewOutboxPublisher(taskRepository, eventProducer, log.Logger, taskEventsConfig)
	publisherDone := make(chan struct{})
	go func() {
		defer close(publisherDone)
		outboxPublisher.Run(ctx)
	}()

	statisticsRepository := statistics_postgres_repository.NewStatisticRepository(pool)

	redisConfig := statistics_cache.NewConfigMust()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
	defer redisClient.Close()

	statisticsCache, err := statistics_cache.New(redisClient, redisConfig.TTL)
	if err != nil {
		log.Fatal("failed to init statistics Redis cache")
	}
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Warn("Redis unavailable", zap.Error(err))
	}
	statisticsService := statistics_service.NewStatisticsService(statisticsRepository, statisticsCache)
	statisticsTransport := statistics_transport_http.NewStatisticsHTTPHandler(statisticsService)

	userroutes := usertransport.Routes()
	taskRoutes := taskTransport.Routes()
	statisticsRoutes := statisticsTransport.Routes()

	apiversionrouter := core_http_server.NewAPIVersionRouter(core_http_server.APIVersion1)
	apiversionrouter.RegisterRoute(userroutes...)
	apiversionrouter.RegisterRoute(taskRoutes...)
	apiversionrouter.RegisterRoute(statisticsRoutes...)

	server := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		log,
		core_http_middleware.CORS(),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(log),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)
	server.RegisterHealthChecks(pool.Ping)

	server.RegisterAPIRouters(apiversionrouter)

	server.RegisterSwagger()

	log.Debug("Starting HTTP server")
	if err := server.Run(ctx); err != nil {
		log.Error("http server run error", zap.Error(err))
	}
	cancel()
	<-publisherDone

}
