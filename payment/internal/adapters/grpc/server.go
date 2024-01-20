package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/huytran2000-hcmus/grpc-microservices-proto/golang/payment"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/huytran2000-hcmus/grpc-microservices/payment/config"
	"github.com/huytran2000-hcmus/grpc-microservices/payment/internal/ports"
)

const (
	serverName = "order"
)

type Adapter struct {
	api    ports.APIPort
	port   int
	server *grpc.Server
	payment.UnimplementedPaymentServer
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

	srvMetric := grpcprom.NewServerMetrics()
	prometheus.MustRegister(srvMetric)
	// Setup metric for panic recoveries.
	panicsTotal := promauto.NewCounter(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	})
	grpcPanicRecoveryHandler := func(p any) (err error) {
		panicsTotal.Inc()
		// level.Error(rpcLogger).Log("msg", "recovered from panic", "panic", p, "stack", debug.Stack())
		return status.Errorf(codes.Internal, "%s", p)
	}

	lgr := zap.L().Named(serverName)
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
				grpc_zap.UnaryServerInterceptor(lgr, zapOpts...),
				otelZapUnaryInterceptor, // This must go after the grpc_zap interceptor
				srvMetric.UnaryServerInterceptor(grpcprom.WithExemplarFromContext(traceExamplar)),
			),
		),

		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
				grpc_zap.StreamServerInterceptor(lgr, zapOpts...),
				otelZapStreamInterceptor, // This must go after the grpc_zap interceptor
				srvMetric.StreamServerInterceptor(grpcprom.WithExemplarFromContext(traceExamplar)),
			),
		),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	}

	grpcSrv := grpc.NewServer(opts...)
	a.server = grpcSrv

	hsrv := health.NewServer()
	hsrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcSrv, hsrv)

	payment.RegisterPaymentServer(grpcSrv, a)
	if config.GetEnv() == "development" {
		reflection.Register(grpcSrv)
	}

	zap.L().Info(fmt.Sprintf("starting payment grpc server on port %d ...", a.port))

	shutdown := make(chan struct{}, 1)
	done := make(chan struct{}, 1)
	go func() {
		runPrometheusServer(config.GetMetricAddress(), shutdown)
		done <- struct{}{}
	}()
	err = grpcSrv.Serve(listen)
	if err != nil {
		zap.L().Error(fmt.Sprintf("failed to serve grpc on port %d", a.port))
	}

	shutdown <- struct{}{}

	<-done
}

func (a *Adapter) Shutdown() {
	zap.L().Info("stopping payment grpc server...")
	a.server.GracefulStop()
}

func runPrometheusServer(addr string, shutdownCh <-chan struct{}) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srv := http.Server{
		Handler: mux,
		Addr:    addr,
	}

	errCh := make(chan error, 1)
	var err error

	go func() {
		zap.L().Info(fmt.Sprintf("starting payment http server on address %s ...", addr))
		err = srv.ListenAndServe()
		errCh <- err
	}()

	select {
	case <-shutdownCh:
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		zap.L().Info("stopping payment http server...")
		err = srv.Shutdown(ctx)
		if err != nil {
			zap.L().Error(err.Error())
		}
	case err := <-errCh:
		zap.L().Error(err.Error())
	}
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
