package api

import (
	"context"

	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/ports"
)

var _ ports.APIPort = (*Application)(nil)

type Application struct {
	db      ports.DBPort
	payment ports.PaymentPort
}

func NewApplication(db ports.DBPort, payment ports.PaymentPort) *Application {
	return &Application{
		db:      db,
		payment: payment,
	}
}

func (a *Application) PlaceOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	err := a.db.Save(ctx, &order)
	if err != nil {
		return domain.Order{}, err
	}
	err = a.payment.Charge(ctx, &order)
	if err != nil {
		return domain.Order{}, err
	}

	return order, nil
}

func (a Application) GetOrder(ctx context.Context, id int64) (domain.Order, error) {
	return a.db.Get(ctx, id)
}
