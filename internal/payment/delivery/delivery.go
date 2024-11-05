package delivery

import (
	"context"
	"github.com/Inspirate789/ds-lab2/internal/models"
	"github.com/Inspirate789/ds-lab2/internal/payment/delivery/errors"
	"github.com/Inspirate789/ds-lab2/internal/pkg/app"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"strconv"
)

type UseCase interface {
	app.HealthChecker
	CreatePayment(ctx context.Context, price uint64) (res models.Payment, err error)
	GetPayment(ctx context.Context, paymentUID string) (res models.Payment, found bool, err error)
	SetPaymentStatus(ctx context.Context, paymentUID string, status models.PaymentStatus) (found bool, err error)
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
	router.Post("/", d.createPayment)
	router.Get("/:paymentUID", d.getPayment)
	router.Put("/:paymentUID/status", d.updatePaymentStatus)
}

func (d *Delivery) createPayment(ctx *fiber.Ctx) error {
	price, err := strconv.ParseUint(ctx.Query("price"), 10, 64)
	if err != nil {
		d.logger.Error(err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(errors.ErrPaymentPriceNotSet.Map())
	}

	payment, err := d.useCase.CreatePayment(ctx.Context(), price)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(NewPaymentDTO(payment))
}

func (d *Delivery) getPayment(ctx *fiber.Ctx) error {
	paymentUID := ctx.Params("paymentUID")

	payment, found, err := d.useCase.GetPayment(ctx.Context(), paymentUID)
	if err != nil {
		return err
	} else if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(errors.ErrPaymentNotFound.Map())
	}

	return ctx.Status(fiber.StatusOK).JSON(NewPaymentDTO(payment))
}

func (d *Delivery) updatePaymentStatus(ctx *fiber.Ctx) error {
	paymentUID := ctx.Params("paymentUID")
	status := models.PaymentStatus(ctx.Body())

	found, err := d.useCase.SetPaymentStatus(ctx.Context(), paymentUID, status)
	if err != nil {
		return err
	} else if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(errors.ErrPaymentNotFound.Map())
	}

	return ctx.SendStatus(fiber.StatusOK)
}
