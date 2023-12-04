package ports

import (
	"context"

	"github.com/huytran2000-hcmus/grpc-microservices/payment/internal/application/core/domain"
)

type APIPort interface {
	Charge(ctx context.Context, payment domain.Payment) (domain.Payment, error)
}
