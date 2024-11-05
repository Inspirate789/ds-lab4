package gateway

import (
	"context"
	carErrors "github.com/Inspirate789/ds-lab2/internal/car/delivery/errors"
	"github.com/Inspirate789/ds-lab2/internal/gateway/errors"
	"github.com/Inspirate789/ds-lab2/internal/models"
	paymentErrors "github.com/Inspirate789/ds-lab2/internal/payment/delivery/errors"
	"github.com/Inspirate789/ds-lab2/internal/pkg/app"
	rentalErrors "github.com/Inspirate789/ds-lab2/internal/rental/delivery/errors"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/multierr"
	"log/slog"
	"math"
	"strconv"
	"time"
)

type CarsAPI interface {
	app.HealthChecker
	GetCars(ctx context.Context, offset, limit uint64, showAll bool) (res []models.Car, totalCount uint64, err error)
	GetCar(ctx context.Context, carUID string) (res models.Car, found bool, err error)
	LockCar(ctx context.Context, carUID string) (res models.Car, found, success bool, err error)
	UnlockCar(ctx context.Context, carUID string) (err error)
}

type RentalsAPI interface {
	app.HealthChecker
	GetUserRentals(ctx context.Context, username string, offset, limit uint64) (res []models.Rental, totalCount uint64, err error)
	GetUserRental(ctx context.Context, rentalUID, username string) (res models.Rental, found, permitted bool, err error)
	CreateRental(ctx context.Context, properties models.RentalProperties) (res models.Rental, err error)
	SetRentalStatus(ctx context.Context, rentalUID string, status models.RentalStatus) (found bool, err error)
}

type PaymentsAPI interface {
	app.HealthChecker
	CreatePayment(ctx context.Context, price uint64) (res models.Payment, err error)
	SetPaymentStatus(ctx context.Context, paymentUID string, status models.PaymentStatus) (found bool, err error)
	GetPayment(ctx context.Context, paymentUID string) (res models.Payment, found bool, err error)
}

type Gateway struct {
	carsAPI     CarsAPI
	rentalsAPI  RentalsAPI
	paymentsAPI PaymentsAPI
	logger      *slog.Logger
}

func New(carsAPI CarsAPI, rentalsAPI RentalsAPI, paymentsAPI PaymentsAPI, logger *slog.Logger) app.Delivery {
	return &Gateway{
		carsAPI:     carsAPI,
		rentalsAPI:  rentalsAPI,
		paymentsAPI: paymentsAPI,
		logger:      logger,
	}
}

func (gateway *Gateway) HealthCheck(ctx context.Context) error {
	return multierr.Combine(
		gateway.carsAPI.HealthCheck(ctx),
		gateway.rentalsAPI.HealthCheck(ctx),
		gateway.paymentsAPI.HealthCheck(ctx),
	)
}

func (gateway *Gateway) AddHandlers(router fiber.Router) {
	router.Get("/cars", gateway.getCars)
	router.Get("/rental", gateway.getRentals)
	router.Post("/rental", gateway.startCarRental)
	router.Get("/rental/:rentalUID", gateway.getRental)
	router.Post("/rental/:rentalUID/finish", gateway.finishCarRental)
	router.Delete("/rental/:rentalUID", gateway.cancelCarRental)
}

func (gateway *Gateway) getCars(ctx *fiber.Ctx) error {
	page, err := strconv.ParseUint(ctx.Query("page"), 10, 64)
	if err != nil {
		gateway.logger.Debug("car list page not set, use default 1")
		page = 1
	} else if page == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(errors.ErrInvalidPage.Map())
	}

	size, err := strconv.ParseUint(ctx.Query("size"), 10, 64)
	if err != nil {
		gateway.logger.Debug("car list size not set, return all cars")
		size = math.MaxInt64
	}

	showAll, err := strconv.ParseBool(ctx.Query("showAll"))
	if err != nil {
		gateway.logger.Debug("'showAll' for car list not set, return only available cars")
		showAll = false
	}

	cars, totalCount, err := gateway.carsAPI.GetCars(ctx.Context(), (page-1)*size, size, showAll)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(NewCarsDTO(cars, page, size, totalCount))
}

func (gateway *Gateway) getRentals(ctx *fiber.Ctx) error {
	page, err := strconv.ParseUint(ctx.Query("page"), 10, 64)
	if err != nil {
		gateway.logger.Debug("car list page not set, use default 1")
		page = 1
	} else if page == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(errors.ErrInvalidPage.Map())
	}

	size, err := strconv.ParseUint(ctx.Query("size"), 10, 64)
	if err != nil {
		gateway.logger.Debug("car list size not set, return all cars")
		size = math.MaxInt64
	}

	username := ctx.Get("X-User-Name")

	rentals, totalCount, err := gateway.rentalsAPI.GetUserRentals(ctx.Context(), username, (page-1)*size, size)
	if err != nil {
		return err
	}

	cars := make(map[string]models.Car)
	for _, rental := range rentals {
		car, found, err := gateway.carsAPI.GetCar(ctx.Context(), rental.CarUID)
		if err != nil {
			return err
		} else if !found {
			return ctx.Status(fiber.StatusNotFound).JSON(carErrors.ErrCarNotFound.Map())
		}

		cars[rental.CarUID] = car
	}

	payments := make([]models.Payment, 0)
	for _, rental := range rentals {
		payment, found, err := gateway.paymentsAPI.GetPayment(ctx.Context(), rental.PaymentUID)
		if err != nil {
			return err
		} else if !found {
			return ctx.Status(fiber.StatusNotFound).JSON(paymentErrors.ErrPaymentNotFound.Map())
		}

		payments = append(payments, payment)
	}

	return ctx.Status(fiber.StatusOK).JSON(NewRentalsDTO(rentals, cars, payments, page, size, totalCount))
}

func (gateway *Gateway) getRental(ctx *fiber.Ctx) error {
	rentalUID := ctx.Params("rentalUID")
	username := ctx.Get("X-User-Name")

	rental, found, permitted, err := gateway.rentalsAPI.GetUserRental(ctx.Context(), rentalUID, username)
	if err != nil {
		return err
	} else if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(rentalErrors.ErrRentalNotFound.Map())
	} else if !permitted {
		return ctx.Status(fiber.StatusForbidden).JSON(rentalErrors.ErrRentalNotPermitted.Map())
	}

	car, found, err := gateway.carsAPI.GetCar(ctx.Context(), rental.CarUID)
	if err != nil {
		return err
	} else if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(carErrors.ErrCarNotFound.Map())
	}

	payment, found, err := gateway.paymentsAPI.GetPayment(ctx.Context(), rental.PaymentUID)
	if err != nil {
		return err
	} else if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(paymentErrors.ErrPaymentNotFound.Map())
	}

	return ctx.Status(fiber.StatusOK).JSON(NewRentalDTO(rental, car, payment))
}

func (gateway *Gateway) startCarRental(ctx *fiber.Ctx) error {
	// 0. Read request data
	username := ctx.Get("X-User-Name")
	var dto CarRentalRequest

	err := ctx.BodyParser(&dto)
	if err != nil {
		gateway.logger.Error(err.Error())
		parseErr := errors.ErrInvalidRentalRequest(err.Error())

		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(parseErr.Map())
	}

	dateFrom, err := time.Parse(time.DateOnly, dto.DateFrom)
	if err != nil {
		gateway.logger.Error(err.Error())
		parseErr := errors.ErrInvalidDateFrom(err.Error())

		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(parseErr.Map())
	}

	dateTo, err := time.Parse(time.DateOnly, dto.DateTo)
	if err != nil {
		gateway.logger.Error(err.Error())
		parseErr := errors.ErrInvalidDateTo(err.Error())

		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(parseErr.Map())
	}

	if !dateTo.After(dateFrom) {
		dateErr := errors.ErrInvalidRentalPeriod(dto.DateFrom, dto.DateTo)
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(dateErr.Map())
	}

	// 1. Lock car
	car, found, success, err := gateway.carsAPI.LockCar(ctx.Context(), dto.CarUID)
	if err != nil {
		return err
	} else if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(carErrors.ErrCarNotFound.Map())
	} else if !success {
		return ctx.Status(fiber.StatusLocked).JSON(carErrors.ErrCarAlreadyRent.Map())
	}

	defer func() {
		if err != nil {
			rollbackErr := gateway.carsAPI.UnlockCar(ctx.Context(), dto.CarUID)
			err = multierr.Append(err, errors.ErrRollbackWrap(rollbackErr))
		}
	}()

	// 2. Create payment
	price := uint64(dateTo.Sub(dateFrom)/(24*time.Hour)) * car.Price

	payment, err := gateway.paymentsAPI.CreatePayment(ctx.Context(), price)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_, rollbackErr := gateway.paymentsAPI.SetPaymentStatus(ctx.Context(), payment.PaymentUID, models.PaymentCanceled)
			err = multierr.Append(err, errors.ErrRollbackWrap(rollbackErr))
		}
	}()

	// 3. Create rental
	rental, err := gateway.rentalsAPI.CreateRental(ctx.Context(), models.RentalProperties{
		Username:   username,
		PaymentUID: payment.PaymentUID,
		CarUID:     dto.CarUID,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		Status:     models.RentalInProgress,
	}) // dto.CarUID, username, dateFrom, dateTo)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(NewRentalResponse(rental, payment))
}

func (gateway *Gateway) cancelCarRental(ctx *fiber.Ctx) (err error) {
	// 0. Read request data
	username := ctx.Get("X-User-Name")
	rentalUID := ctx.Params("rentalUID")

	// 1. Check rental access
	rental, found, permitted, err := gateway.rentalsAPI.GetUserRental(ctx.Context(), rentalUID, username)
	if err != nil {
		return err
	} else if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(rentalErrors.ErrRentalNotFound.Map())
	} else if !permitted {
		return ctx.Status(fiber.StatusForbidden).JSON(rentalErrors.ErrRentalNotPermitted.Map())
	}

	// 2. Unlock car
	err = gateway.carsAPI.UnlockCar(ctx.Context(), rental.CarUID)
	if err != nil {
		return err
	}

	// 3. Cancel rental
	_, err = gateway.rentalsAPI.SetRentalStatus(ctx.Context(), rentalUID, models.RentalCanceled)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_, rollbackErr := gateway.rentalsAPI.SetRentalStatus(ctx.Context(), rentalUID, models.RentalInProgress)
			err = multierr.Append(err, errors.ErrRollbackWrap(rollbackErr))
		}
	}()

	// 4. Cancel payment
	found, err = gateway.paymentsAPI.SetPaymentStatus(ctx.Context(), rental.PaymentUID, models.PaymentCanceled)
	if err != nil {
		return err
	} else if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(paymentErrors.ErrPaymentNotFound.Map())
	}

	defer func() {
		if err != nil {
			_, rollbackErr := gateway.paymentsAPI.SetPaymentStatus(ctx.Context(), rental.PaymentUID, models.PaymentPaid)
			err = multierr.Append(err, errors.ErrRollbackWrap(rollbackErr))
		}
	}()

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (gateway *Gateway) finishCarRental(ctx *fiber.Ctx) error {
	// 0. Read request data
	username := ctx.Get("X-User-Name")
	rentalUID := ctx.Params("rentalUID")

	// 1. Check rental access
	rental, found, permitted, err := gateway.rentalsAPI.GetUserRental(ctx.Context(), rentalUID, username)
	if err != nil {
		return err
	} else if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(rentalErrors.ErrRentalNotFound.Map())
	} else if !permitted {
		return ctx.Status(fiber.StatusForbidden).JSON(rentalErrors.ErrRentalNotPermitted.Map())
	}

	// 2. Unlock car
	err = gateway.carsAPI.UnlockCar(ctx.Context(), rental.CarUID)
	if err != nil {
		return err
	}

	// 3. Finish rental
	_, err = gateway.rentalsAPI.SetRentalStatus(ctx.Context(), rentalUID, models.RentalFinished)
	if err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
