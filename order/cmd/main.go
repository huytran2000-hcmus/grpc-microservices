package main

import (
	"log"

	"go.uber.org/zap"

	"github.com/huytran2000-hcmus/grpc-microservices/order/config"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/adapters/db"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/adapters/grpc"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/adapters/payment"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/api"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/logger"
)

func main() {
	logger, err := logger.NewLogger(zap.InfoLevel)
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)

	dbAdapter, err := db.NewAdapter(config.GetDataSourceURL())
	if err != nil {
		log.Fatalf("failed to create db adapter: %s", err)
	}

	paymentAdapter, err := payment.NewAdapter(config.GetPaymentServiceURL())
	if err != nil {
		log.Fatalf("failed to create order adapter: %s", err)
	}

	api := api.NewApplication(dbAdapter, paymentAdapter)

	grpcAdapter := grpc.NewAdapter(api, config.GetApplicationPort())

	grpcAdapter.Run()
}
