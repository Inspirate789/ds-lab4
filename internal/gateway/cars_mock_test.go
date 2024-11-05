package gateway_test

import (
	"context"
	"github.com/Inspirate789/ds-lab2/internal/models"
	"github.com/stretchr/testify/mock"
)

type carsApiMock struct {
	mock.Mock
}

func (api *carsApiMock) HealthCheck(ctx context.Context) error {
	return api.Called(ctx).Error(0)
}

func (api *carsApiMock) GetCars(ctx context.Context, offset, limit uint64, showAll bool) (res []models.Car, totalCount uint64, err error) {
	args := api.Called(ctx, offset, limit, showAll)
	return args.Get(0).([]models.Car), args.Get(1).(uint64), args.Error(2)
}

func (api *carsApiMock) GetCar(ctx context.Context, carUID string) (res models.Car, found bool, err error) {
	args := api.Called(ctx, carUID)
	return args.Get(0).(models.Car), args.Bool(1), args.Error(2)
}

func (api *carsApiMock) LockCar(ctx context.Context, carUID string) (res models.Car, found, success bool, err error) {
	args := api.Called(ctx, carUID)
	return args.Get(0).(models.Car), args.Bool(1), args.Bool(2), args.Error(3)
}

func (api *carsApiMock) UnlockCar(ctx context.Context, carUID string) (err error) {
	args := api.Called(ctx, carUID)
	return args.Error(0)
}
