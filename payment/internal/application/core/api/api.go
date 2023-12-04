package api

import (
	"context"

	"github.com/huytran2000-hcmus/grpc-microservices/payment/internal/application/core/domain"
	"github.com/huytran2000-hcmus/grpc-microservices/payment/internal/ports"
)

type Application struct {
	db ports.DBPort
}

func NewApplication(db ports.DBPort) *Application {
	return &Application{
		db: db,
	}
}

func (a Application) Charge(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	err := a.db.Save(ctx, &payment)
	if err != nil {
		return domain.Payment{}, err
	}
	return payment, nil
}
