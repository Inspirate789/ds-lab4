package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Inspirate789/ds-lab2/internal/models"
	"github.com/Inspirate789/ds-lab2/pkg/sqlxutils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"log/slog"
)

type SqlxRepository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewSqlxRepository(db *sqlx.DB, logger *slog.Logger) *SqlxRepository {
	return &SqlxRepository{
		db:     db,
		logger: logger,
	}
}

func (r *SqlxRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *SqlxRepository) CreatePayment(ctx context.Context, price uint64) (res models.Payment, err error) {
	payment := PaymentDTO{
		ID:         0,
		PaymentUID: uuid.New().String(),
		Status:     models.PaymentPaid,
		Price:      price,
	}

	err = sqlxutils.NamedGet(ctx, r.db, &payment, insertPaymentQuery, &payment)
	if err != nil {
		return models.Payment{}, err
	}

	return payment.ToModel(), nil
}

func (r *SqlxRepository) GetPayment(ctx context.Context, paymentUID string) (models.Payment, bool, error) {
	var dto PaymentDTO

	err := sqlxutils.Get(ctx, r.db, &dto, selectPaymentQuery, paymentUID)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Payment{}, false, nil
	} else if err != nil {
		return models.Payment{}, false, err
	}

	return dto.ToModel(), true, nil
}

func (r *SqlxRepository) SetPaymentStatus(ctx context.Context, paymentUID string, status models.PaymentStatus) (found bool, err error) {
	res, err := sqlxutils.Exec(ctx, r.db, updatePaymentStatusQuery, paymentUID, status)
	if err != nil {
		return false, err
	}

	rowsCount, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsCount != 0, err
}
