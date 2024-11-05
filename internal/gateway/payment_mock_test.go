package gateway_test

import (
	"context"
	"github.com/Inspirate789/ds-lab2/internal/models"
	"github.com/stretchr/testify/mock"
)

type paymentApiMock struct {
	mock.Mock
}

func (api *paymentApiMock) HealthCheck(ctx context.Context) error {
	return api.Called(ctx).Error(0)
}

func (api *paymentApiMock) CreatePayment(ctx context.Context, price uint64) (res models.Payment, err error) {
	args := api.Called(ctx, price)
	return args.Get(0).(models.Payment), args.Error(1)
}

func (api *paymentApiMock) SetPaymentStatus(ctx context.Context, paymentUID string, status models.PaymentStatus) (found bool, err error) {
	args := api.Called(ctx, paymentUID, status)
	return args.Bool(0), args.Error(1)
}

func (api *paymentApiMock) GetPayment(ctx context.Context, paymentUID string) (res models.Payment, found bool, err error) {
	args := api.Called(ctx, paymentUID)
	return args.Get(0).(models.Payment), args.Bool(1), nil
}
