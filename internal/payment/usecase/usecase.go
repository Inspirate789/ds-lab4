package usecase

import (
	"context"
	"github.com/Inspirate789/ds-lab2/internal/models"
	"log/slog"
)

type Repository interface {
	HealthCheck(ctx context.Context) error
	CreatePayment(ctx context.Context, price uint64) (res models.Payment, err error)
	GetPayment(ctx context.Context, paymentUID string) (res models.Payment, found bool, err error)
	SetPaymentStatus(ctx context.Context, paymentUID string, status models.PaymentStatus) (found bool, err error)
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

func (u *UseCase) CreatePayment(ctx context.Context, price uint64) (res models.Payment, err error) {
	return u.repo.CreatePayment(ctx, price)
}

func (u *UseCase) GetPayment(ctx context.Context, paymentUID string) (res models.Payment, found bool, err error) {
	return u.repo.GetPayment(ctx, paymentUID)
}

func (u *UseCase) SetPaymentStatus(ctx context.Context, paymentUID string, status models.PaymentStatus) (found bool, err error) {
	return u.repo.SetPaymentStatus(ctx, paymentUID, status)
}
