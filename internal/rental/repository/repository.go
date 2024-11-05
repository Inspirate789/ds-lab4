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

func (r *SqlxRepository) GetUserRentals(ctx context.Context, username string, offset, limit uint64) ([]models.Rental, uint64, error) {
	rentals := make(RentalsDTO, 0)

	err := sqlxutils.Select(ctx, r.db, &rentals, selectRentalsQuery, offset, limit, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, nil
	} else if err != nil {
		return nil, 0, err
	}

	model, totalCount := rentals.ToModel()

	return model, totalCount, nil
}

func (r *SqlxRepository) CreateRental(ctx context.Context, properties models.RentalProperties) (models.Rental, error) {
	dto := RentalDTO{
		ID:                  0,
		RentalUID:           uuid.New().String(),
		RentalPropertiesDTO: NewRentalPropertiesDTO(properties),
	}

	err := sqlxutils.NamedGet(ctx, r.db, &dto, insertRentalQuery, &dto)
	if err != nil {
		return models.Rental{}, err
	}

	return dto.ToModel(), nil
}

func (r *SqlxRepository) GetUserRental(ctx context.Context, rentalUID, username string) (res models.Rental, found, permitted bool, err error) {
	var dto RentalDTO

	err = sqlxutils.Get(ctx, r.db, &dto, selectRentalQuery, rentalUID)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Rental{}, false, false, nil
	} else if dto.Username != username {
		return models.Rental{}, true, false, nil
	} else if err != nil {
		return models.Rental{}, true, false, err
	}

	return dto.ToModel(), true, true, nil
}

func (r *SqlxRepository) SetRentalStatus(ctx context.Context, rentalUID string, status models.RentalStatus) (found bool, err error) {
	res, err := sqlxutils.Exec(ctx, r.db, updateRentalStatusQuery, rentalUID, status)
	if err != nil {
		return false, err
	}

	rowsCount, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsCount != 0, err
}
