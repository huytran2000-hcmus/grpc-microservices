package ports

import (
	"context"

	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"
)

type PaymentPort interface {
	Charge(context.Context, *domain.Order) error
}
