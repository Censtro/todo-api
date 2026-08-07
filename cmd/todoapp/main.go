package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	core_logger "github.com/Censtro/todo-api/internal/core/logger"
	core_pgx_pool "github.com/Censtro/todo-api/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "github.com/Censtro/todo-api/internal/core/transport/http/middleware"
	core_http_server "github.com/Censtro/todo-api/internal/core/transport/http/server"
	user_postgres_repository "github.com/Censtro/todo-api/internal/features/users/repository/postgres"
	users_service "github.com/Censtro/todo-api/internal/features/users/service"
	user_transport_http "github.com/Censtro/todo-api/internal/features/users/transport/http"
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
		fmt.Println("failed to init logger: ", err)
		os.Exit(1)
	}
	defer log.Close()
	log.Debug("start app")

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

	userroutes := usertransport.Routes()
	apiversionrouter := core_http_server.NewAPIVersionRouter(core_http_server.APIVersion1)
	apiversionrouter.RegisterRoute(userroutes...)

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
