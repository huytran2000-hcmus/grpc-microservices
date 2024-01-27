package metric

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
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

	res, err := rsc.New(serviceName, serviceVersion)
	if err != nil {
		return shutdown, err
	}

	metricProvider, err := newMeterProvider(res, otlpEndpoint)
	if err != nil {
		return shutdown, err
	}

	shutdownFuncs = append(shutdownFuncs, metricProvider.Shutdown)
	otel.SetMeterProvider(metricProvider)

	return shutdown, nil
}

func newMeterProvider(res *resource.Resource, otlpEndpoint string) (*metric.MeterProvider, error) {
	ctx := context.Background()
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(otlpEndpoint),
		otlpmetricgrpc.WithTLSCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create new exporter: %w", err)
	}

	provider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(
			metric.NewPeriodicReader(exporter, metric.WithInterval(10*time.Second)),
		),
	)

	return provider, nil
}
