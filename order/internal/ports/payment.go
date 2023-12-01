package ports

import "github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"

type PaymentPort interface {
	Charge(*domain.Order) error
}
