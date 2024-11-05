package gateway

import (
	"github.com/Inspirate789/ds-lab2/internal/models"
	"time"
)

type CarRentalRequest struct {
	CarUID   string `json:"carUid"`   // UUID
	DateFrom string `json:"dateFrom"` // ISO 8601
	DateTo   string `json:"dateTo"`   // ISO 8601
}

type CarRentalPayment struct {
	PaymentUID string               `json:"paymentUid,omitempty"`
	Status     models.PaymentStatus `json:"status,omitempty"`
	Price      uint64               `json:"price,omitempty"`
}

type CarRentalResponse struct {
	RentalUID string              `json:"rentalUid,omitempty"`
	Status    models.RentalStatus `json:"status,omitempty"`
	CarUID    string              `json:"carUid,omitempty"`
	DateFrom  string              `json:"dateFrom,omitempty"`
	DateTo    string              `json:"dateTo,omitempty"`
	Payment   CarRentalPayment    `json:"payment,omitempty"`
}

func NewRentalResponse(rental models.Rental, payment models.Payment) CarRentalResponse {
	return CarRentalResponse{
		RentalUID: rental.RentalUID,
		Status:    rental.Status,
		CarUID:    rental.CarUID,
		DateFrom:  rental.DateFrom.Format(time.DateOnly),
		DateTo:    rental.DateTo.Format(time.DateOnly),
		Payment: CarRentalPayment{
			PaymentUID: payment.PaymentUID,
			Status:     payment.Status,
			Price:      payment.Price,
		},
	}
}

type CarDTO struct {
	CarUID             string         `json:"carUid,omitempty"`
	Brand              string         `json:"brand,omitempty"`
	Model              string         `json:"model,omitempty"`
	RegistrationNumber string         `json:"registrationNumber,omitempty"`
	Power              uint64         `json:"power,omitempty"`
	Price              uint64         `json:"price,omitempty"`
	Type               models.CarType `json:"type,omitempty"`
	Availability       bool           `json:"available,omitempty"`
}

func NewCarsDTO(cars []models.Car, page, pageSize, totalCount uint64) map[string]any {
	items := make([]CarDTO, 0, len(cars))

	for _, car := range cars {
		items = append(items, CarDTO{
			CarUID:             car.CarUID,
			Brand:              car.Brand,
			Model:              car.Model,
			RegistrationNumber: car.RegistrationNumber,
			Power:              car.Power,
			Price:              car.Price,
			Type:               car.Type,
			Availability:       car.Availability,
		})
	}

	return map[string]any{
		"page":          page,
		"pageSize":      pageSize,
		"totalElements": totalCount,
		"items":         items,
	}
}

type RentalCarDTO struct {
	CarUID             string `json:"carUid,omitempty"`
	Brand              string `json:"brand,omitempty"`
	Model              string `json:"model,omitempty"`
	RegistrationNumber string `json:"registrationNumber,omitempty"`
}

type RentalPayment struct {
	PaymentUID string               `json:"paymentUid,omitempty"`
	Status     models.PaymentStatus `json:"status,omitempty"`
	Price      uint64               `json:"price,omitempty"`
}

type RentalDTO struct {
	RentalUID string              `json:"rentalUid,omitempty"`
	DateFrom  string              `json:"dateFrom,omitempty"`
	DateTo    string              `json:"dateTo,omitempty"`
	Status    models.RentalStatus `json:"status,omitempty"`
	Car       RentalCarDTO        `json:"car,omitempty"`
	Payment   RentalPayment       `json:"payment,omitempty"`
}

func NewRentalDTO(rental models.Rental, car models.Car, payment models.Payment) RentalDTO {
	return RentalDTO{
		RentalUID: rental.RentalUID,
		DateFrom:  rental.DateFrom.Format(time.DateOnly),
		DateTo:    rental.DateTo.Format(time.DateOnly),
		Status:    rental.Status,
		Car: RentalCarDTO{
			CarUID:             car.CarUID,
			Brand:              car.Brand,
			Model:              car.Model,
			RegistrationNumber: car.RegistrationNumber,
		},
		Payment: RentalPayment{
			PaymentUID: payment.PaymentUID,
			Status:     payment.Status,
			Price:      payment.Price,
		},
	}
}

func NewRentalsDTO(rentals []models.Rental, cars map[string]models.Car, payments []models.Payment, _, _, _ uint64) []RentalDTO {
	items := make([]RentalDTO, 0, len(rentals))

	for i := range rentals {
		items = append(items, NewRentalDTO(rentals[i], cars[rentals[i].CarUID], payments[i]))
	}

	return items
}
