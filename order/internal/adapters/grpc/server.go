package grpc

import (
	"context"
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/huytran2000-hcmus/grpc-microservices-proto/golang/order"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/huytran2000-hcmus/grpc-microservices/order/config"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/ports"
)

var meter = otel.Meter("order_grpc")

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
	logger := zap.L()

	var err error
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to listen on port %d, error: %v", a.port, err))
	}
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithDecider(func(fullMethodName string, err error) bool {
			if err == nil && fullMethodName == "/grpc.health.v1.Health/Check" {
				return false
			}

			return true
		}),
	}

	panicsTotalName := "grpc_req_panics_recovered_count"
	panicsTotal, err := meter.Int64Counter(panicsTotalName,
		metric.WithDescription("Total number of gRPC requests recovered from internal panic."),
		metric.WithUnit("{count}"),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("create %s meter: %v", panicsTotalName, err))
	}
	grpcPanicRecoveryHandler := func(p any) (err error) {
		panicsTotal.Add(context.Background(), 1)
		// level.Error(rpcLogger).Log("msg", "recovered from panic", "panic", p, "stack", debug.Stack())
		return status.Errorf(codes.Internal, "%s", p)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
				grpc_zap.UnaryServerInterceptor(logger, zapOpts...),
				otelZapUnaryInterceptor, // This must go after the grpc_zap interceptor
			),
		),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
				grpc_zap.StreamServerInterceptor(logger, zapOpts...),
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

	logger.Info(fmt.Sprintf("starting order grpc server on port %d ...", a.port))

	err = grpcSrv.Serve(listen)
	if err != nil {
		logger.Error(fmt.Sprintf("serve grpc on port: %d, %s", a.port, err))
	}
}

func (a *Adapter) ShutDown() {
	zap.L().Info("stopping payment grpc server...")
	a.server.GracefulStop()
}

func otelZapUnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	addTraceField(ctx)

	return handler(ctx, req)
}

func otelZapStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	addTraceField(ctx)

	return handler(ctx, ss)
}

func addTraceField(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		sCTX := span.SpanContext()
		ctxzap.AddFields(ctx, zap.String("trace.id", sCTX.TraceID().String()))
		ctxzap.AddFields(ctx, zap.String("span.id", sCTX.SpanID().String()))

	}

	peer, ok := peer.FromContext(ctx)
	if ok {
		ctxzap.AddFields(ctx, zap.String("peer.address", peer.Addr.String()))
	}
}

func traceExamplar(ctx context.Context) prometheus.Labels {
	span := trace.SpanContextFromContext(ctx)
	if !span.IsSampled() {
		return nil
	}

	return prometheus.Labels{"traceID": span.TraceID().String(), "spanID": span.SpanID().String()}
}
