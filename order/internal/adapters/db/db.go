package db

import (
	"context"
	"fmt"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"
	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/ports"
)

var _ ports.DBPort = &Adapter{}

type Order struct {
	gorm.Model
	CustomerID int64
	Status     string
	OrderItems []OrderItem
}

type OrderItem struct {
	gorm.Model
	ProductCode string
	UnitPrice   float32
	Quantity    int32
	OrderID     uint
}

type Adapter struct {
	db *gorm.DB
}

func NewAdapter(connStr string) (*Adapter, error) {
	db, err := gorm.Open(mysql.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open db connection %s: %w", connStr, err)
	}

	err = db.Use(otelgorm.NewPlugin(otelgorm.WithDBName("order")))
	if err != nil {
		return nil, fmt.Errorf("use opentelemetry plugin: %w", err)
	}

	err = db.AutoMigrate(&Order{}, &OrderItem{})
	if err != nil {
		return nil, fmt.Errorf("migrate table: %w", err)
	}

	return &Adapter{
		db: db,
	}, nil
}

func (a *Adapter) Get(ctx context.Context, id int64) (domain.Order, error) {
	var orderEntity Order
	err := a.db.WithContext(ctx).Preload("OrderItems").First(&orderEntity, id).Error
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order: %w", err)
	}
	var orderItems []domain.OrderItem
	for _, item := range orderEntity.OrderItems {
		orderItems = append(orderItems, orderItemEntityToDomain(item))
	}

	order := domain.Order{
		ID:         int64(orderEntity.ID),
		CustomerID: orderEntity.CustomerID,
		Status:     orderEntity.Status,
		OrderItems: orderItems,
		CreatedAt:  orderEntity.CreatedAt.Unix(),
	}

	return order, nil
}

func (a *Adapter) Save(ctx context.Context, order *domain.Order) error {
	var orderItemEntities []OrderItem
	for _, item := range order.OrderItems {
		orderItemEntities = append(orderItemEntities, OrderItem{
			ProductCode: item.ProductCode,
			UnitPrice:   item.UnitPrice,
			Quantity:    item.Quantity,
		})
	}

	orderEntity := Order{
		CustomerID: order.CustomerID,
		Status:     order.Status,
		OrderItems: orderItemEntities,
	}

	err := a.db.WithContext(ctx).Create(&orderEntity).Error
	if err != nil {
		return fmt.Errorf("insert db order: %w", err)
	}

	order.ID = int64(orderEntity.ID)

	return nil
}

func orderItemEntityToDomain(item OrderItem) domain.OrderItem {
	return domain.OrderItem{
		ProductCode: item.ProductCode,
		UnitPrice:   item.UnitPrice,
		Quantity:    item.Quantity,
	}
}
