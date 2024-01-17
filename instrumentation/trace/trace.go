package trace

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials/insecure"

	rsc "github.com/huytran2000-hcmus/grpc-microservices/instrumentation/internal/resource"
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

	res, err := rsc.NewResource(serviceName, serviceVersion)
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
