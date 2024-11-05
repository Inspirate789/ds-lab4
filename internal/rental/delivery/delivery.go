package delivery

import (
	"context"
	"github.com/Inspirate789/ds-lab2/internal/models"
	"github.com/Inspirate789/ds-lab2/internal/pkg/app"
	"github.com/Inspirate789/ds-lab2/internal/rental/delivery/errors"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"math"
	"strconv"
)

type UseCase interface {
	app.HealthChecker
	GetUserRentals(ctx context.Context, username string, offset, limit uint64) (res []models.Rental, totalCount uint64, err error)
	GetUserRental(ctx context.Context, rentalUID, username string) (res models.Rental, found, permitted bool, err error)
	CreateRental(ctx context.Context, properties models.RentalProperties) (res models.Rental, err error)
	SetRentalStatus(ctx context.Context, rentalUID string, status models.RentalStatus) (found bool, err error)
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
	router.Get("/", d.getRentals)
	router.Post("/", d.createRental)
	router.Get("/:rentalUID", d.getRental)
	router.Put("/:rentalUID/status", d.updateRentalStatus)
}

func (d *Delivery) getRentals(ctx *fiber.Ctx) error {
	offset, err := strconv.ParseUint(ctx.Query("offset"), 10, 64)
	if err != nil {
		d.logger.Debug("rentals offset not set, use default 0")
		offset = 0
	}

	limit, err := strconv.ParseUint(ctx.Query("limit"), 10, 64)
	if err != nil {
		d.logger.Debug("rentals limit not set, return all rentals")
		limit = math.MaxInt64
	}

	username := ctx.Get("X-User-Name")

	rentals, totalCount, err := d.useCase.GetUserRentals(ctx.Context(), username, offset, limit)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(NewRentalsDTO(rentals, totalCount))
}

func (d *Delivery) createRental(ctx *fiber.Ctx) error {
	var dto RentalPropertiesDTO

	err := ctx.BodyParser(&dto)
	if err != nil {
		d.logger.Error(err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(errors.ErrInvalidRentalRequest.Map())
	}

	rentalProperties, err := dto.ToModel()
	if err != nil {
		d.logger.Error(err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(errors.ErrConvertRentalRequest.Map())
	}

	rental, err := d.useCase.CreateRental(ctx.Context(), rentalProperties)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(NewRentalDTO(rental))
}

func (d *Delivery) getRental(ctx *fiber.Ctx) error {
	rentalUID := ctx.Params("rentalUID")
	username := ctx.Get("X-User-Name")

	rental, found, permitted, err := d.useCase.GetUserRental(ctx.Context(), rentalUID, username)
	if err != nil {
		return err
	} else if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(errors.ErrRentalNotFound.Map())
	} else if !permitted {
		return ctx.Status(fiber.StatusForbidden).JSON(errors.ErrRentalNotPermitted.Map())
	}

	return ctx.Status(fiber.StatusOK).JSON(NewRentalDTO(rental))
}

func (d *Delivery) updateRentalStatus(ctx *fiber.Ctx) error {
	rentalUID := ctx.Params("rentalUID")
	status := models.RentalStatus(ctx.Body())

	found, err := d.useCase.SetRentalStatus(ctx.Context(), rentalUID, status)
	if err != nil {
		return err
	} else if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(errors.ErrRentalNotFound.Map())
	}

	return ctx.SendStatus(fiber.StatusOK)
}
