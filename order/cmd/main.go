package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"

	"github.com/huytran2000-hcmus/grpc-microservices/order/config"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/adapters/db"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/adapters/grpc"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/adapters/payment"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/api"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/instrumentation"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/logger"
)

const (
	serviceName    = "order"
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
		log.Fatalf("setup opentelemetry: %s", err)
	}
	defer func() {
		err := otelShutdown(context.Background())
		log.Fatal(err)
	}()

	dbAdapter, err := db.NewAdapter(config.GetDataSourceURL())
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to create db adapter: %s", err))
	}

	paymentAdapter, err := payment.NewAdapter(config.GetPaymentServiceURL())
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to create order adapter: %s", err))
	}

	api := api.NewApplication(dbAdapter, paymentAdapter)
	grpcAdapter := grpc.NewAdapter(api, config.GetApplicationPort())

	setupGracefulShutdown(grpcAdapter.ShutDown)
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
