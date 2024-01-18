package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/huytran2000-hcmus/grpc-microservices/instrumentation/metric"
	"github.com/huytran2000-hcmus/grpc-microservices/instrumentation/trace"
	"go.uber.org/zap"

	"github.com/huytran2000-hcmus/grpc-microservices/order/config"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/adapters/db"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/adapters/grpc"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/adapters/payment"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/api"
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

	if config.GetOTLPEndpoint() != "" {
		otelTraceShutdown, err := trace.SetupOtelSDK(context.Background(), serviceName, serviceVersion, config.GetOTLPEndpoint())
		if err != nil {
			log.Fatalf("setup opentelemetry: %s", err)
		}
		defer func() {
			err := otelTraceShutdown(context.Background())
			log.Fatal(err)
		}()
	}

	otelMetricShutdown, err := metric.SetupOtelSDK(context.Background(), serviceName, serviceVersion)
	if err != nil {
		logger.Fatal(fmt.Sprintf("setup metric sdk: %s", err))
	}
	defer func() {
		err := otelMetricShutdown(context.Background())
		logger.Fatal(err.Error())
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
