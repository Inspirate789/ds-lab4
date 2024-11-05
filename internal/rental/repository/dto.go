package repository

import (
	"github.com/Inspirate789/ds-lab2/internal/models"
	"time"
)

type RentalPropertiesDTO struct {
	Username   string              `db:"username"`
	PaymentUID string              `db:"payment_uid"`
	CarUID     string              `db:"car_uid"`
	DateFrom   time.Time           `db:"date_from"`
	DateTo     time.Time           `db:"date_to"`
	Status     models.RentalStatus `db:"status"`
	TotalCount uint64              `db:"total_count"`
}

func NewRentalPropertiesDTO(properties models.RentalProperties) RentalPropertiesDTO {
	return RentalPropertiesDTO{
		Username:   properties.Username,
		PaymentUID: properties.PaymentUID,
		CarUID:     properties.CarUID,
		DateFrom:   properties.DateFrom,
		DateTo:     properties.DateTo,
		Status:     properties.Status,
	}
}

func (rental RentalPropertiesDTO) ToModel() models.RentalProperties {
	return models.RentalProperties{
		Username:   rental.Username,
		PaymentUID: rental.PaymentUID,
		CarUID:     rental.CarUID,
		DateFrom:   rental.DateFrom,
		DateTo:     rental.DateTo,
		Status:     rental.Status,
	}
}

type RentalDTO struct {
	ID        int64  `db:"id"`
	RentalUID string `db:"rental_uid"`
	RentalPropertiesDTO
}

func (rental RentalDTO) ToModel() models.Rental {
	return models.Rental{
		ID:               rental.ID,
		RentalUID:        rental.RentalUID,
		RentalProperties: rental.RentalPropertiesDTO.ToModel(),
	}
}

type RentalsDTO []RentalDTO

func (rentals RentalsDTO) ToModel() ([]models.Rental, uint64) {
	result := make([]models.Rental, 0, len(rentals))

	for _, rental := range rentals {
		result = append(result, rental.ToModel())
	}

	if len(rentals) != 0 {
		return result, rentals[0].TotalCount
	} else {
		return result, 0
	}
}
