package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Inspirate789/ds-lab2/internal/models"
	"github.com/Inspirate789/ds-lab2/pkg/sqlxutils"
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

func (r *SqlxRepository) GetCars(ctx context.Context, offset, limit uint64, showAll bool) ([]models.Car, uint64, error) {
	cars := make(CarsDTO, 0)

	err := sqlxutils.Select(ctx, r.db, &cars, selectCarsQuery, offset, limit, showAll)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, nil
	} else if err != nil {
		return nil, 0, err
	}

	model, totalCount := cars.ToModel()

	return model, totalCount, nil
}

func (r *SqlxRepository) GetCar(ctx context.Context, carUID string) (models.Car, bool, error) {
	var dto CarDTO

	err := sqlxutils.Get(ctx, r.db, &dto, selectCarQuery, carUID)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Car{}, false, nil
	} else if err != nil {
		return models.Car{}, false, err
	}

	return dto.ToModel(), true, nil
}

func (r *SqlxRepository) LockCar(ctx context.Context, carUID string) (res models.Car, found, success bool, err error) {
	var dto CarDTO

	err = sqlxutils.RunTx(ctx, r.db, sql.LevelDefault, func(tx *sqlx.Tx) error {
		err = sqlxutils.Get(ctx, r.db, &dto, selectCarQuery, carUID)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
			return nil
		} else if err != nil {
			return err
		} else {
			found = true
		}

		return sqlxutils.Get(ctx, r.db, &dto, lockCarQuery, carUID)
	})
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	} else {
		success = true
	}

	return dto.ToModel(), found, success, err
}

func (r *SqlxRepository) UnlockCar(ctx context.Context, carUID string) error {
	_, err := sqlxutils.Exec(ctx, r.db, unlockCarQuery, carUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}

	return err
}
