package ports

import (
	"context"

	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"
)

type DBPort interface {
	Get(ctx context.Context, id int64) (domain.Order, error)
	Save(ctx context.Context, order *domain.Order) error
}
