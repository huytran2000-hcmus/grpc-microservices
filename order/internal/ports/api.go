package ports

import (
	"context"

	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"
)

type APIPort interface {
	PlaceOrder(order domain.Order) (domain.Order, error)
	GetOrder(ctx context.Context, id int64) (domain.Order, error)
}
