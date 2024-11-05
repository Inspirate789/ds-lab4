package usecase

import (
	"context"
	"github.com/Inspirate789/ds-lab2/internal/models"
	"log/slog"
)

type Repository interface {
	HealthCheck(ctx context.Context) error
	GetCars(ctx context.Context, offset, limit uint64, showAll bool) (res []models.Car, totalCount uint64, err error)
	GetCar(ctx context.Context, carUID string) (res models.Car, found bool, err error)
	LockCar(ctx context.Context, carUID string) (res models.Car, found, success bool, err error)
	UnlockCar(ctx context.Context, carUID string) (err error)
}

type UseCase struct {
	repo   Repository
	logger *slog.Logger
}

func New(repo Repository, logger *slog.Logger) *UseCase {
	return &UseCase{
		repo:   repo,
		logger: logger,
	}
}

func (u *UseCase) HealthCheck(ctx context.Context) error {
	return u.repo.HealthCheck(ctx)
}

func (u *UseCase) GetCars(ctx context.Context, offset, limit uint64, showAll bool) (res []models.Car, totalCount uint64, err error) {
	return u.repo.GetCars(ctx, offset, limit, showAll)
}

func (u *UseCase) GetCar(ctx context.Context, carUID string) (res models.Car, found bool, err error) {
	return u.repo.GetCar(ctx, carUID)
}

func (u *UseCase) LockCar(ctx context.Context, carUID string) (res models.Car, found, success bool, err error) {
	return u.repo.LockCar(ctx, carUID)
}

func (u *UseCase) UnlockCar(ctx context.Context, carUID string) (err error) {
	return u.repo.UnlockCar(ctx, carUID)
}
