package instrumentation

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials/insecure"
)

func SetupOtelSDK(ctx context.Context, serviceName, serviceVersion string, otlpEndpoint string) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}

		return err
	}

	handleErr := func(err error) error {
		return errors.Join(err, shutdown(ctx))
	}
	defer func() {
		if err != nil {
			err = handleErr(err)
		}
	}()

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return shutdown, err
	}

	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	traceProvider, err := newTraceProvider(res, otlpEndpoint)
	if err != nil {
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, traceProvider.Shutdown)
	otel.SetTracerProvider(traceProvider)

	return shutdown, nil
}

func WrapOtelLogger(logger *zap.Logger) *otelzap.Logger {
	return otelzap.New(logger)
}

func newTraceProvider(res *resource.Resource, otlpEndpoint string) (*trace.TracerProvider, error) {
	ctx := context.Background()
	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(
			traceExporter,
			trace.WithBatchTimeout(3*time.Second),
		),
		trace.WithResource(res),
	)

	return traceProvider, nil
}
