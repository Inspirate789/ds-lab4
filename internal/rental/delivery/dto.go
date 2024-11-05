package delivery

import (
	"github.com/Inspirate789/ds-lab2/internal/models"
	"time"
)

type RentalPropertiesDTO struct {
	Username   string              `json:"username"`
	PaymentUID string              `json:"paymentUid"`
	CarUID     string              `json:"carUid"`
	DateFrom   string              `json:"dateFrom"`
	DateTo     string              `json:"dateTo"`
	Status     models.RentalStatus `json:"status"`
}

func NewRentalPropertiesDTO(properties models.RentalProperties) RentalPropertiesDTO {
	return RentalPropertiesDTO{
		Username:   properties.Username,
		PaymentUID: properties.PaymentUID,
		CarUID:     properties.CarUID,
		DateFrom:   properties.DateFrom.Format(time.DateOnly),
		DateTo:     properties.DateTo.Format(time.DateOnly),
		Status:     properties.Status,
	}
}

func (rental RentalPropertiesDTO) ToModel() (models.RentalProperties, error) {
	dateFrom, err := time.Parse(time.DateOnly, rental.DateFrom)
	if err != nil {
		return models.RentalProperties{}, err
	}

	dateTo, err := time.Parse(time.DateOnly, rental.DateTo)
	if err != nil {
		return models.RentalProperties{}, err
	}

	return models.RentalProperties{
		Username:   rental.Username,
		PaymentUID: rental.PaymentUID,
		CarUID:     rental.CarUID,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		Status:     rental.Status,
	}, nil
}

type RentalDTO struct {
	ID        int64  `json:"id"`
	RentalUID string `json:"rentalUid"`
	RentalPropertiesDTO
}

func NewRentalDTO(rental models.Rental) RentalDTO {
	return RentalDTO{
		ID:                  rental.ID,
		RentalUID:           rental.RentalUID,
		RentalPropertiesDTO: NewRentalPropertiesDTO(rental.RentalProperties),
	}
}

func (rental RentalDTO) ToModel() (models.Rental, error) {
	properties, err := rental.RentalPropertiesDTO.ToModel()
	if err != nil {
		return models.Rental{}, err
	}

	return models.Rental{
		ID:               rental.ID,
		RentalUID:        rental.RentalUID,
		RentalProperties: properties,
	}, nil
}

type RentalsDTO struct {
	Items []RentalDTO `json:"items"`
	Count uint64      `json:"count"`
}

func NewRentalsDTO(rentals []models.Rental, totalCount uint64) RentalsDTO {
	items := make([]RentalDTO, 0, len(rentals))

	for _, rental := range rentals {
		items = append(items, NewRentalDTO(rental))
	}

	return RentalsDTO{
		Items: items,
		Count: totalCount,
	}
}

func (rentals RentalsDTO) ToModel() ([]models.Rental, error) {
	res := make([]models.Rental, 0, len(rentals.Items))

	for _, rental := range rentals.Items {
		model, err := rental.ToModel()
		if err != nil {
			return nil, err
		}

		res = append(res, model)
	}

	return res, nil
}
