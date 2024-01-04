package grpc

import (
	"fmt"
	"log"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/huytran2000-hcmus/grpc-microservices-proto/golang/order"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/huytran2000-hcmus/grpc-microservices/order/config"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/ports"
)

const (
	serverName = "order"
)

type Adapter struct {
	api    ports.APIPort
	port   int
	server *grpc.Server
	order.UnimplementedOrderServer
}

func NewAdapter(api ports.APIPort, port int) *Adapter {
	return &Adapter{api: api, port: port}
}

func (a *Adapter) Run() {
	var err error
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Fatalf("failed to listen on port %d, error: %v", a.port, err)
	}
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithDecider(func(fullMethodName string, err error) bool {
			if err == nil && fullMethodName == "/grpc.health.v1.Health/Check" {
				return false
			}

			return true
		}),
	}

	lgr := zap.L().Named(serverName)
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_zap.UnaryServerInterceptor(lgr, zapOpts...),
			),
		),

		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(),
				grpc_zap.StreamServerInterceptor(lgr, zapOpts...),
			),
		),
	}

	grpcSrv := grpc.NewServer(opts...)

	hsrv := health.NewServer()
	hsrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcSrv, hsrv)

	order.RegisterOrderServer(grpcSrv, a)
	if config.GetEnv() == "development" {
		reflection.Register(grpcSrv)
	}
	a.server = grpcSrv

	log.Printf("starting order service on port %d ...", a.port)
	err = grpcSrv.Serve(listen)
	if err != nil {
		log.Fatalf("serve grpc on port: %d, %s", a.port, err)
	}
}
