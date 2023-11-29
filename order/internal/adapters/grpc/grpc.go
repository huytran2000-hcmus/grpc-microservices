package grpc

import (
	"context"

	"github.com/huytran2000-hcmus/grpc-microservices-proto/golang/order"

	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"
)

var _ order.OrderServer = (*Adapter)(nil)

func (a *Adapter) Create(ctx context.Context, request *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
	var orderItems []domain.OrderItem
	for _, oItem := range request.OrderItems {
		orderItems = append(orderItems, domain.OrderItem{
			ProductCode: oItem.ProductCode,
			UnitPrice:   oItem.UnitPrice,
			Quantity:    oItem.Quantity,
		})
	}

	newOrder := domain.NewOrder(request.UserId, orderItems)
	res, err := a.api.PlaceOrder(newOrder)
	if err != nil {
		return nil, err
	}

	return &order.CreateOrderResponse{OrderId: res.ID}, nil
}
