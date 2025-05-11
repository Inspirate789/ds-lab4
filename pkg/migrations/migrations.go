package migrations

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"log/slog"
)

func Do(connString, migrationsPath string, logger *slog.Logger) error {
	m, err := migrate.New("file://"+migrationsPath, connString)
	if err != nil {
		return err
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return err
	} else if dirty {
		return fmt.Errorf("current database migration version %d is dirty", version)
	}

	logger.Info(fmt.Sprintf("current database migration version is %d; migrate up", version))

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		rollbackErr := m.Force(int(version))
		if rollbackErr != nil {
			logger.Warn(err.Error() + "; rollback")
			logger.Warn("rollback migration")
			return rollbackErr
		} else {
			logger.Warn("rollback migration")
			return err
		}
	} else if errors.Is(err, migrate.ErrNoChange) {
		logger.Info("current database migration version is up to date")
	} else {
		version, dirty, err = m.Version()
		if err != nil {
			return err
		} else if dirty {
			return fmt.Errorf("database migration version %d is dirty", version)
		}

		logger.Info(fmt.Sprintf("migrated to version %d", version))
	}

	return nil
}
