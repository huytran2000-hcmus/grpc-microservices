package ports

import "github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"

type APIPort interface {
	PlaceOrder(order domain.Order) (domain.Order, error)
}
