package ports

import "github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"

type DBPort interface {
	Get(id int64) (domain.Order, error)
	Save(*domain.Order) error
}
