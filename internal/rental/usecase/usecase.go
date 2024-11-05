package usecase

import (
	"context"
	"github.com/Inspirate789/ds-lab2/internal/models"
	"log/slog"
)

type Repository interface {
	HealthCheck(ctx context.Context) error
	GetUserRentals(ctx context.Context, username string, offset, limit uint64) (res []models.Rental, totalCount uint64, err error)
	GetUserRental(ctx context.Context, rentalUID, username string) (res models.Rental, found, permitted bool, err error)
	CreateRental(ctx context.Context, properties models.RentalProperties) (res models.Rental, err error)
	SetRentalStatus(ctx context.Context, rentalUID string, status models.RentalStatus) (found bool, err error)
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

func (u *UseCase) GetUserRentals(ctx context.Context, username string, offset, limit uint64) (res []models.Rental, totalCount uint64, err error) {
	return u.repo.GetUserRentals(ctx, username, offset, limit)
}

func (u *UseCase) GetUserRental(ctx context.Context, rentalUID, username string) (res models.Rental, found, permitted bool, err error) {
	return u.repo.GetUserRental(ctx, rentalUID, username)
}

func (u *UseCase) CreateRental(ctx context.Context, properties models.RentalProperties) (res models.Rental, err error) {
	return u.repo.CreateRental(ctx, properties)
}

func (u *UseCase) SetRentalStatus(ctx context.Context, rentalUID string, status models.RentalStatus) (found bool, err error) {
	return u.repo.SetRentalStatus(ctx, rentalUID, status)
}
