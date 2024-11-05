package delivery

import (
	"context"
	"github.com/Inspirate789/ds-lab2/internal/car/delivery/errors"
	"github.com/Inspirate789/ds-lab2/internal/models"
	"github.com/Inspirate789/ds-lab2/internal/pkg/app"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"math"
	"strconv"
)

type UseCase interface {
	app.HealthChecker
	GetCars(ctx context.Context, offset, limit uint64, showAll bool) (res []models.Car, totalCount uint64, err error)
	GetCar(ctx context.Context, carUID string) (res models.Car, found bool, err error)
	LockCar(ctx context.Context, carUID string) (res models.Car, found, success bool, err error)
	UnlockCar(ctx context.Context, carUID string) (err error)
}

type Delivery struct {
	useCase UseCase
	logger  *slog.Logger
}

func New(useCase UseCase, logger *slog.Logger) *Delivery {
	return &Delivery{
		useCase: useCase,
		logger:  logger,
	}
}

func (d *Delivery) HealthCheck(ctx context.Context) error {
	return d.useCase.HealthCheck(ctx)
}

func (d *Delivery) AddHandlers(router fiber.Router) {
	router.Get("/", d.getCars)
	router.Get("/:carUID", d.getCar)
	router.Post("/:carUID/lock", d.lockCar)
	router.Delete("/:carUID/lock", d.unlockCar)
}

func (d *Delivery) getCars(ctx *fiber.Ctx) error {
	offset, err := strconv.ParseUint(ctx.Query("offset"), 10, 64)
	if err != nil {
		d.logger.Debug("cars offset not set, use default 0")
		offset = 0
	}

	limit, err := strconv.ParseUint(ctx.Query("limit"), 10, 64)
	if err != nil {
		d.logger.Debug("cars limit not set, return all cars")
		limit = math.MaxInt64
	}

	showAll, err := strconv.ParseBool(ctx.Query("showAll"))
	if err != nil {
		d.logger.Debug("'showAll' for car list not set, return only available cars")
		showAll = false
	}

	cars, totalCount, err := d.useCase.GetCars(ctx.Context(), offset, limit, showAll)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(NewCarsDTO(cars, totalCount))
}

func (d *Delivery) getCar(ctx *fiber.Ctx) error {
	carUID := ctx.Params("carUID")

	car, found, err := d.useCase.GetCar(ctx.Context(), carUID)
	if err != nil {
		return err
	} else if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(errors.ErrCarNotFound.Map())
	}

	return ctx.Status(fiber.StatusOK).JSON(NewCarDTO(car))
}

func (d *Delivery) lockCar(ctx *fiber.Ctx) error {
	carUID := ctx.Params("carUID")

	car, found, success, err := d.useCase.LockCar(ctx.Context(), carUID)
	if err != nil {
		return err
	} else if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(errors.ErrCarNotFound.Map())
	} else if !success {
		return ctx.Status(fiber.StatusLocked).JSON(errors.ErrCarAlreadyRent.Map())
	}

	return ctx.Status(fiber.StatusOK).JSON(NewCarDTO(car))
}

func (d *Delivery) unlockCar(ctx *fiber.Ctx) error {
	carUID := ctx.Params("carUID")

	err := d.useCase.UnlockCar(ctx.Context(), carUID)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusOK)
}
