package api_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/api"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"
	"github.com/huytran2000-hcmus/grpc-microservices/order/mocks/ports"
)

func TestPlaceOrder(t *testing.T) {
	payment := ports.NewMockPaymentPort(t)
	payment.EXPECT().Charge(mock.Anything).Return(nil)

	db := ports.NewMockDBPort(t)
	db.EXPECT().Save(mock.Anything).Return(nil)

	app := api.NewApplication(db, payment)
	_, err := app.PlaceOrder(domain.Order{
		CustomerID: 1,
		Status:     "Pending",
		OrderItems: []domain.OrderItem{
			{
				ProductCode: "camera",
				UnitPrice:   12.3,
				Quantity:    3,
			},
		},
	})

	assert.NoError(t, err)
}

func TestReturnStatusWhenPaymentFailed(t *testing.T) {
	var userID int64 = 1

	errMess := "insufficent balance"
	failedErr := errors.New(errMess)
	payment := ports.NewMockPaymentPort(t)
	payment.EXPECT().Charge(mock.Anything).Return(failedErr)

	db := ports.NewMockDBPort(t)
	db.EXPECT().Save(mock.Anything).Return(nil)

	app := api.NewApplication(db, payment)
	_, err := app.PlaceOrder(domain.NewOrder(userID, []domain.OrderItem{
		{
			ProductCode: "camera",
			UnitPrice:   12.3,
			Quantity:    3,
		},
	}))

	t.Log(err)
	assert.ErrorIs(t, failedErr, err)
}
