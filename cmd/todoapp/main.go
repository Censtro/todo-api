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
	tasks_postgres_repository "github.com/Censtro/todo-api/internal/features/tasks/repository/postgres"
	task_service "github.com/Censtro/todo-api/internal/features/tasks/service"
	tasks_transport_http "github.com/Censtro/todo-api/internal/features/tasks/transport/http"
	user_postgres_repository "github.com/Censtro/todo-api/internal/features/users/repository/postgres"
	users_service "github.com/Censtro/todo-api/internal/features/users/service"
	user_transport_http "github.com/Censtro/todo-api/internal/features/users/transport/http"
	"go.uber.org/zap"
)

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

	userroutes := usertransport.Routes()
	taskRoutes := taskTransport.Routes()

	apiversionrouter := core_http_server.NewAPIVersionRouter(core_http_server.APIVersion1)
	apiversionrouter.RegisterRoute(userroutes...)
	apiversionrouter.RegisterRoute(taskRoutes...)

	server := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		log,
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(log),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)

	server.RegisterAPIRouters(apiversionrouter)
	log.Debug("Starting HTTP server")
	if err := server.Run(ctx); err != nil {
		log.Error("http server run error", zap.Error(err))
	}

}
