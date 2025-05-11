package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	slogfiber "github.com/samber/slog-fiber"
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

func checkReadiness(delivery HealthChecker, logger *slog.Logger) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		err := delivery.HealthCheck(ctx.UserContext())
		if err != nil {
			logger.Error(err.Error())
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(newFiberError(err.Error()))
		}

		return ctx.Status(fiber.StatusOK).SendString("healthy")
	}
}

func ExtractServiceUnavailableErr(err error) (error, bool) { // TODO: Check by error type
	var DNSError *net.DNSError

	if merr, ok := err.(*multierror.Error); ok {
		fmt.Println(merr.Errors)
		fmt.Printf("%T\n", merr.Errors[0])

		if len(merr.Errors) == 0 {
			return err, false
		}

		srcErr := merr.Errors[0]
		if strings.Contains(strings.ToLower(srcErr.Error()), "service unavailable") ||
			errors.As(srcErr, &DNSError) ||
			errors.Is(srcErr, syscall.ECONNREFUSED) {
			return srcErr, true
		}
	}

	if strings.Contains(strings.ToLower(err.Error()), "service unavailable") ||
		errors.As(err, &DNSError) ||
		errors.Is(err, syscall.ECONNREFUSED) {
		return err, true
	}

	return err, false
}

func NewFiberApp(config WebConfig, delivery Delivery, logger *slog.Logger) *FiberApp {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			logger.Error(err.Error(), slog.String("errorType", fmt.Sprintf("%T", err)))

			if err, ok := ExtractServiceUnavailableErr(err); ok {
				msg := strings.SplitN(err.Error(), ":", 2)[0]
				return ctx.Status(fiber.StatusServiceUnavailable).JSON(newFiberError(msg))
			}

			return ctx.Status(fiber.StatusInternalServerError).JSON(newFiberError(err.Error()))
		},
	})

	app.Use(recover.New())
	app.Use(slogfiber.New(logger))
	app.Use(pprof.New())

	app.Get("/manage/health", checkReadiness(delivery, logger))

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
