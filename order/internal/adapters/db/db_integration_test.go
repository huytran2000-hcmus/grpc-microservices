// go:build integration
package db

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/huytran2000-hcmus/grpc-microservices/order/internal/appication/core/domain"
)

type OrderDatabaseTestSuite struct {
	suite.Suite
	DataSourceURL string
	container     *mysql.MySQLContainer
}

func TestDB(t *testing.T) {
	suite.Run(t, new(OrderDatabaseTestSuite))
}

func (s *OrderDatabaseTestSuite) TestSaveOrder() {
	adapter, err := NewAdapter(s.DataSourceURL)
	s.Require().NoError(err)
	err = adapter.Save(&domain.Order{})
	s.Require().NoError(err)
}

func (s *OrderDatabaseTestSuite) TestGetOrder() {
	adapter, err := NewAdapter(s.DataSourceURL)
	s.Require().NoError(err)

	order := domain.NewOrder(1, []domain.OrderItem{
		{
			ProductCode: "CAM",
			Quantity:    5,
			UnitPrice:   1.32,
		},
	})
	err = adapter.Save(&order)
	s.Require().NoError(err)

	ord, err := adapter.Get(order.ID)
	s.Require().NoError(err)
	s.Equal(order.CustomerID, ord.CustomerID)
}

func (s *OrderDatabaseTestSuite) SetupSuite() {
	ctx := context.Background()

	user := "order_service"
	password := "somepassword"
	port := "3306/tcp"
	db := "orders"
	url := func(address string) string {
		return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, address, db)
	}
	dbURL := func(host string, port nat.Port) string {
		return url(fmt.Sprintf("%s:%s", host, port.Port()))
	}

	req := testcontainers.ContainerRequest{
		Image:        "mysql:8.0.35",
		ExposedPorts: []string{port, port},
		Env: map[string]string{
			"MYSQL_USER":          user,
			"MYSQL_ROOT_PASSWORD": password,
			"MYSQL_PASSWORD":      password,
			"MYSQL_DATABASE":      db,
		},
		WaitingFor: wait.ForSQL(nat.Port(port), "mysql", dbURL),
	}
	mysqlContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("set up mysql container: %s", err)
	}

	ctnHost, _ := mysqlContainer.Endpoint(ctx, "")
	s.DataSourceURL = url(ctnHost)
}

func (s *OrderDatabaseTestSuite) TeardownSuite() {
	err := s.container.Terminate(context.Background())
	if err != nil {
		log.Fatalf("terminate mysql container: %s", err)
	}
}
