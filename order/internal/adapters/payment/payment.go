package payment

import (
	"context"
	"fmt"

	"github.com/huytran2000-hcmus/grpc-microservices-proto/golang/payment"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"
)

type Adapter struct {
	payment payment.PaymentClient
}

func NewAdapter(paymentServiceURL string) (*Adapter, error) {
	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	conn, err := grpc.Dial(paymentServiceURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("dial to payment service: %w", err)
	}
	client := payment.NewPaymentClient(conn)

	return &Adapter{
		payment: client,
	}, nil
}

func (a *Adapter) Charge(ctx context.Context, order *domain.Order) error {
	_, err := a.payment.Create(ctx, &payment.CreatePaymentRequest{
		UserId:     order.CustomerID,
		OrderId:    order.ID,
		TotalPrice: order.TotalPrice(),
	})
	if err != nil {
		return fmt.Errorf("failed to create a payment: %w", err)
	}

	return nil
}
