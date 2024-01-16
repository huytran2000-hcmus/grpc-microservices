package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"

	"github.com/huytran2000-hcmus/grpc-microservices/payment/config"
	"github.com/huytran2000-hcmus/grpc-microservices/payment/internal/adapters/db"
	"github.com/huytran2000-hcmus/grpc-microservices/payment/internal/adapters/grpc"
	"github.com/huytran2000-hcmus/grpc-microservices/payment/internal/application/core/api"
	"github.com/huytran2000-hcmus/grpc-microservices/payment/internal/instrumentation"
	"github.com/huytran2000-hcmus/grpc-microservices/payment/internal/logger"
)

const (
	serviceName    = "payment"
	serviceVersion = "0.1.0"
)

func main() {
	logger, err := logger.NewLogger(zap.InfoLevel)
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)

	otelLogger := instrumentation.WrapOtelLogger(logger)
	otelzap.ReplaceGlobals(otelLogger)

	otelShutdown, err := instrumentation.SetupOtelSDK(context.Background(), serviceName, serviceVersion, config.GetOTLPEndpoint())
	if err != nil {
		logger.Fatal(fmt.Sprintf("setup opentelemetry: %s", err))
	}
	defer func() {
		err := otelShutdown(context.Background())
		logger.Fatal(err.Error())
	}()

	dbAdapter, err := db.NewAdapter(config.GetDataSourceURL())
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to connect to database. Error: %v", err))
	}

	api := api.NewApplication(dbAdapter)
	grpcAdapter := grpc.NewAdapter(api, config.GetApplicationPort())

	setupGracefulShutdown(grpcAdapter.Shutdown)
	grpcAdapter.Run()
}

func setupGracefulShutdown(shutdownServer func()) {
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

		s := <-quit
		zap.L().Info(fmt.Sprintf("Receive signal %s", s.String()))
		shutdownServer()
	}()
}
