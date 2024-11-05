package gateway_test

import (
	"context"
	"github.com/Inspirate789/ds-lab2/internal/models"
	"github.com/stretchr/testify/mock"
)

type rentalApiMock struct {
	mock.Mock
}

func (api *rentalApiMock) HealthCheck(ctx context.Context) error {
	return api.Called(ctx).Error(0)
}

func (api *rentalApiMock) GetUserRentals(ctx context.Context, username string, offset, limit uint64) (res []models.Rental, totalCount uint64, err error) {
	args := api.Called(ctx, username, offset, limit)
	return args.Get(0).([]models.Rental), args.Get(1).(uint64), nil
}

func (api *rentalApiMock) GetUserRental(ctx context.Context, rentalUID, username string) (res models.Rental, found, permitted bool, err error) {
	args := api.Called(ctx, rentalUID, username)
	return args.Get(0).(models.Rental), args.Get(1).(bool), args.Get(2).(bool), args.Error(3)
}

func (api *rentalApiMock) CreateRental(ctx context.Context, properties models.RentalProperties) (res models.Rental, err error) {
	args := api.Called(ctx, properties)
	return args.Get(0).(models.Rental), args.Error(1)
}

func (api *rentalApiMock) SetRentalStatus(ctx context.Context, rentalUID string, status models.RentalStatus) (found bool, err error) {
	args := api.Called(ctx, rentalUID, status)
	return args.Bool(0), args.Error(1)
}
