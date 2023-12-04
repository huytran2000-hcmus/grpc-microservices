package grpc

import (
	"context"
	"fmt"

	"github.com/huytran2000-hcmus/grpc-microservices-proto/golang/payment"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/huytran2000-hcmus/grpc-microservices/payment/internal/application/core/domain"
)

func (a Adapter) Create(ctx context.Context, request *payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	log.WithContext(ctx).Info("Creating payment...")
	newPayment := domain.NewPayment(request.UserId, request.OrderId, request.TotalPrice)
	result, err := a.api.Charge(ctx, newPayment)
	// st := status.Newf(codes.InvalidArgument, "failed to charge user: %d", request.UserId)
	// fieldErr := &errdetails.BadRequest_FieldViolation{
	// 	Field:       "user_id",
	// 	Description: "invalid user id",
	// }
	// badReq := &errdetails.BadRequest{
	// 	FieldViolations: []*errdetails.BadRequest_FieldViolation{fieldErr},
	// }
	// st, _ = st.WithDetails(badReq)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, fmt.Sprintf("failed to charge. %v ", err)).Err()
	}
	return &payment.CreatePaymentResponse{PaymentId: result.ID}, nil
}
