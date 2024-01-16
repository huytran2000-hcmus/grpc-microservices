package grpc

import (
	"context"
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/huytran2000-hcmus/grpc-microservices-proto/golang/order"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/peer"
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
		zap.L().Fatal(fmt.Sprintf("failed to listen on port %d, error: %v", a.port, err))
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
				grpc_zap.UnaryServerInterceptor(lgr, zapOpts...),
				otelZapUnaryInterceptor, // This must go after the grpc_zap interceptor
			),
		),

		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_zap.StreamServerInterceptor(lgr, zapOpts...),
				otelZapStreamInterceptor, // This must go after the grpc_zap interceptor
			),
		),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	}

	grpcSrv := grpc.NewServer(opts...)
	a.server = grpcSrv

	hsrv := health.NewServer()
	hsrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcSrv, hsrv)

	order.RegisterOrderServer(grpcSrv, a)
	if config.GetEnv() == "development" {
		reflection.Register(grpcSrv)
	}

	zap.L().Info(fmt.Sprintf("starting order service on port %d ...", a.port))
	err = grpcSrv.Serve(listen)
	if err != nil {
		zap.L().Info(fmt.Sprintf("serve grpc on port: %d, %s", a.port, err))
	}
}

func (a *Adapter) ShutDown() {
	a.server.GracefulStop()
}

func otelZapUnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	span := trace.SpanFromContext(ctx)
	sCTX := span.SpanContext()
	ctxzap.AddFields(ctx, zap.String("trace.id", sCTX.TraceID().String()))
	ctxzap.AddFields(ctx, zap.String("span.id", sCTX.SpanID().String()))

	peer, ok := peer.FromContext(ctx)
	if ok {
		ctxzap.AddFields(ctx, zap.String("peer.address", peer.Addr.String()))
	}

	return handler(ctx, req)
}

func otelZapStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	span := trace.SpanFromContext(ctx)
	sCTX := span.SpanContext()
	ctxzap.AddFields(ctx, zap.String("trace.id", sCTX.TraceID().String()))
	ctxzap.AddFields(ctx, zap.String("span.id", sCTX.SpanID().String()))

	peer, ok := peer.FromContext(ctx)
	if ok {
		ctxzap.AddFields(ctx, zap.String("peer.address", peer.Addr.String()))
	}

	return handler(ctx, ss)
}
