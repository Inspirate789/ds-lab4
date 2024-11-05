package main

import (
	"context"
	"fmt"
	"github.com/Inspirate789/ds-lab2/internal/pkg/app"
	"github.com/Inspirate789/ds-lab2/internal/rental/delivery"
	"github.com/Inspirate789/ds-lab2/internal/rental/repository"
	"github.com/Inspirate789/ds-lab2/internal/rental/usecase"
	"github.com/Inspirate789/ds-lab2/pkg/migrations"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/lmittmann/tint"
	"github.com/spf13/pflag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type WebApp interface {
	Start() error
	Shutdown(ctx context.Context) error
}

func startApp(webApp WebApp, config app.Config, logger *slog.Logger) {
	logger.Debug(fmt.Sprintf("web app starts at %s with configuration: %+v", config.Web.Host+":"+config.Web.Port, config))

	go func() {
		err := webApp.Start()
		if err != nil {
			panic(err)
		}
	}()
}

func shutdownApp(webApp WebApp, logger *slog.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Debug("shutdown web app ...")

	const shutdownTimeout = time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

	err := webApp.Shutdown(ctx)
	if err != nil {
		panic(err)
	}

	cancel()
	logger.Debug("web app exited")
}

func main() {
	var configPath, migrationsPath string
	pflag.StringVarP(&configPath, "config", "c", "configs/rental.yaml", "Config file path")
	pflag.StringVarP(&migrationsPath, "migrations", "", "migrations", "Migrations directory path")
	pflag.Parse()

	config, err := app.ReadLocalConfig(configPath)
	if err != nil {
		panic(err)
	}

	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.Level(config.Logging.Level)}))

	db, err := sqlx.Connect(config.DB.DriverName, config.DB.ConnectionString)
	if err != nil {
		panic(err)
	}

	defer func(db *sqlx.DB) {
		err = db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	err = migrations.Do(config.DB.ConnectionString, migrationsPath, logger)
	if err != nil {
		panic(err)
	}

	repo := repository.NewSqlxRepository(db, logger)
	useCase := usecase.New(repo, logger)
	webApp := app.NewFiberApp(config.Web, delivery.New(useCase, logger), logger)

	startApp(webApp, config, logger)
	shutdownApp(webApp, logger)
}
