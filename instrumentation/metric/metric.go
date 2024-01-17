package metric

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/huytran2000-hcmus/grpc-microservices/instrumentation/internal/resource"
)

func SetupOtelSDK(ctx context.Context, serviceName, serviceVersion string) (shutdown func(context.Context) error, err error) {
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

	res, err := resource.NewResource(serviceName, serviceVersion)
	if err != nil {
		return nil, err
	}

	exporter, err := prometheus.New(
		prometheus.WithNamespace(serviceName),
	)
	if err != nil {
		return nil, err
	}

	metricProvider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res),
	)

	shutdownFuncs = append(shutdownFuncs, metricProvider.Shutdown)
	otel.SetMeterProvider(metricProvider)

	return shutdown, nil
}
