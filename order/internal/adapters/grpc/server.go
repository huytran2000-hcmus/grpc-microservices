package grpc

import (
	"fmt"
	"log"
	"net"

	"github.com/huytran2000-hcmus/grpc-microservices-proto/golang/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/huytran2000-hcmus/grpc-microservices/order/config"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/ports"
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
		log.Fatal("failed to listen on port %d, error: %v", a.port, err)
	}

	grpcSrv := grpc.NewServer()
	order.RegisterOrderServer(grpcSrv, a)
	if config.GetEnv() == "development" {
		reflection.Register(grpcSrv)
	}

	err = grpcSrv.Serve(listen)
	if err != nil {
		log.Fatalf("serve grpc on port: %d, %s", a.port, err)
	}
}
