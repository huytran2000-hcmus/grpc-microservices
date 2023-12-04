package domain

import (
	"time"
)

type OrderItem struct {
	ProductCode string  `json:"product_code"`
	UnitPrice   float32 `json:"unit_price"`
	Quantity    int32   `json:"quantity"`
}

type Order struct {
	ID         int64       `json:"id"`
	CustomerID int64       `json:"customer_id"`
	Status     string      `json:"status"`
	OrderItems []OrderItem `json:"order_items"`
	CreatedAt  int64       `json:"created_at"`
}

func NewOrder(customerID int64, orderItems []OrderItem) Order {
	return Order{
		CustomerID: customerID,
		Status:     "Pending",
		OrderItems: orderItems,
		CreatedAt:  time.Now().Unix(),
	}
}

func (o *Order) TotalPrice() float32 {
	var totalPrice float32
	for _, item := range o.OrderItems {
		totalPrice += float32(item.Quantity) * item.UnitPrice
	}

	return totalPrice
}