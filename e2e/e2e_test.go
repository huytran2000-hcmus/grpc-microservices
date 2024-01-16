package e2e

import (
	"context"
	"testing"

	order_proto "github.com/huytran2000-hcmus/grpc-microservices-proto/golang/order"
	"github.com/stretchr/testify/suite"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CreateOrderSuite struct {
	suite.Suite
}

func TestCreateOrder(t *testing.T) {
	suite.Run(t, &CreateOrderSuite{})
}

func (s *CreateOrderSuite) TestCreateOrder() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "localhost:3000", opts...)
	s.Require().NoError(err)
	defer func() {
		s.NoError(conn.Close())
	}()

	client := order_proto.NewOrderClient(conn)
	createReq := &order_proto.CreateOrderRequest{
		UserId: 1,
		OrderItems: []*order_proto.OrderItem{
			{
				ProductCode: "CAM123",
				Quantity:    3,
				UnitPrice:   1.23,
			},
		},
	}
	createResp, err := client.Create(ctx, createReq)
	s.Require().NoError(err)

	getResp, err := client.Get(ctx, &order_proto.GetOrderRequest{
		OrderId: createResp.OrderId,
	})
	s.Require().NoError(err)

	s.Equal(createReq.UserId, getResp.UserId)
	s.Require().Equal(len(createReq.OrderItems), len(getResp.OrderItems))

	getItemReq := createReq.OrderItems[0]
	getItemResp := getResp.OrderItems[0]
	s.Equal(getItemReq.ProductCode, getItemResp.ProductCode)
	s.Equal(getItemReq.Quantity, getItemResp.Quantity)
	s.Equal(getItemReq.UnitPrice, getItemResp.UnitPrice)
}

func (s *CreateOrderSuite) SetupSuite() {
	id := tc.StackIdentifier("order_creation_e2e")
	compose, err := tc.NewDockerComposeWith(tc.WithStackFiles("./resources/docker-compose.yaml"), id)
	s.Require().NoError(err)

	ctx := context.Background()
	s.T().Cleanup(func() {
		s.Require().NoError(compose.Down(ctx, tc.RemoveImagesLocal, tc.RemoveOrphans(true), tc.RemoveVolumes(true)), "Compose Down")
	})

	// toCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	// s.T().Cleanup(cancel)

	err = compose.Up(context.Background(), tc.Wait(true))
	s.Require().NoError(err, "Compose Up")
}
