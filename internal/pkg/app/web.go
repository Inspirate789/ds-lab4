package app

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pkg/errors"
	slogfiber "github.com/samber/slog-fiber"
	"log/slog"
	"net"
	"net/http"
	"strings"
)

type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

type Delivery interface {
	HealthChecker
	AddHandlers(router fiber.Router)
}

type WebConfig struct {
	Host       string
	Port       string
	PathPrefix string
}

type FiberApp struct {
	config WebConfig
	fiber  *fiber.App
	logger *slog.Logger
}

func newFiberError(msg string) fiber.Map {
	return fiber.Map{"message": msg}
}

func checkReadiness(delivery HealthChecker) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		err := delivery.HealthCheck(ctx.UserContext())
		if err != nil {
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(newFiberError(err.Error()))
		}

		return ctx.Status(fiber.StatusOK).SendString("healthy")
	}
}

func NewFiberApp(config WebConfig, delivery Delivery, logger *slog.Logger) *FiberApp {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			logger.Error(err.Error())
			msg := strings.SplitN(err.Error(), ":", 2)[0]

			var DNSError *net.DNSError
			if errors.As(err, &DNSError) {
				return ctx.Status(fiber.StatusServiceUnavailable).JSON(newFiberError(msg))
			}

			return ctx.Status(fiber.StatusInternalServerError).JSON(newFiberError(msg))
		},
	})

	app.Use(recover.New())
	app.Use(slogfiber.New(logger))
	app.Use(pprof.New())

	app.Get("/manage/health", checkReadiness(delivery))

	delivery.AddHandlers(app.Group(config.PathPrefix))

	return &FiberApp{
		config: config,
		fiber:  app,
		logger: logger,
	}
}

func (f *FiberApp) Start() error {
	return errors.Wrap(f.fiber.Listen(f.config.Host+":"+f.config.Port), "start web app")
}

func (f *FiberApp) Shutdown(ctx context.Context) error {
	return errors.Wrap(f.fiber.ShutdownWithContext(ctx), "stop web app")
}

func (f *FiberApp) Test(req *http.Request, msTimeout ...int) (*http.Response, error) {
	return f.fiber.Test(req, msTimeout...)
}
