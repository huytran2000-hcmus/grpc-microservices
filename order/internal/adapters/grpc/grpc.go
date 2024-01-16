package grpc

import (
	"context"
	"fmt"

	"github.com/huytran2000-hcmus/grpc-microservices-proto/golang/order"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"
)

var _ order.OrderServer = (*Adapter)(nil)

func (a Adapter) Get(ctx context.Context, request *order.GetOrderRequest) (*order.GetOrderResponse, error) {
	result, err := a.api.GetOrder(ctx, request.OrderId)
	var orderItems []*order.OrderItem
	for _, orderItem := range result.OrderItems {
		orderItems = append(orderItems, &order.OrderItem{
			ProductCode: orderItem.ProductCode,
			UnitPrice:   orderItem.UnitPrice,
			Quantity:    orderItem.Quantity,
		})
	}
	if err != nil {
		return nil, err
	}

	return &order.GetOrderResponse{UserId: result.CustomerID, OrderItems: orderItems}, nil
}

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
	res, err := a.api.PlaceOrder(ctx, newOrder)
	if err != nil {
		return nil, makeStatusFromErr("payment", err).Err()
	}

	return &order.CreateOrderResponse{OrderId: res.ID}, nil
}

func makeStatusFromErr(context string, err error) *status.Status {
	if st := status.Convert(err); st != nil {
		var fieldErrs []*errdetails.BadRequest_FieldViolation
		for _, detail := range st.Details() {
			switch t := detail.(type) {
			case *errdetails.BadRequest:
				fieldErrs = append(fieldErrs, addContextToFieldViolations(context, t.FieldViolations)...)
			}
		}
		newStatus := status.Newf(st.Code(), "order creation failed: %s", st.Message())
		if len(fieldErrs) != 0 {
			badReq := &errdetails.BadRequest{
				FieldViolations: fieldErrs,
			}
			newStatus, _ = newStatus.WithDetails(badReq)
		}
		return newStatus
	} else {
		st := status.New(codes.Unknown, err.Error())
		return st
	}
}

func addContextToFieldViolations(context string, violations []*errdetails.BadRequest_FieldViolation) []*errdetails.BadRequest_FieldViolation {
	var fieldErrs []*errdetails.BadRequest_FieldViolation
	for _, violation := range violations {
		fieldErrs = append(fieldErrs, &errdetails.BadRequest_FieldViolation{
			Field:       fmt.Sprintf("%s.%s", context, violation.Field),
			Description: violation.Description,
		})
	}

	return fieldErrs
}
