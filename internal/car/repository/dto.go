package repository

import "github.com/Inspirate789/ds-lab2/internal/models"

type CarDTO struct {
	ID                 int64          `db:"id"`
	CarUID             string         `db:"car_uid"`
	Brand              string         `db:"brand"`
	Model              string         `db:"model"`
	RegistrationNumber string         `db:"registration_number"`
	Power              uint64         `db:"power"`
	Price              uint64         `db:"price"`
	Type               models.CarType `db:"type"`
	Availability       bool           `db:"availability"`
	TotalCount         uint64         `db:"total_count"`
}

func (car CarDTO) ToModel() models.Car {
	return models.Car{
		ID:                 car.ID,
		CarUID:             car.CarUID,
		Brand:              car.Brand,
		Model:              car.Model,
		RegistrationNumber: car.RegistrationNumber,
		Power:              car.Power,
		Price:              car.Price,
		Type:               car.Type,
		Availability:       car.Availability,
	}
}

type CarsDTO []CarDTO

func (cars CarsDTO) ToModel() ([]models.Car, uint64) {
	result := make([]models.Car, 0, len(cars))

	for _, car := range cars {
		result = append(result, car.ToModel())
	}

	if len(cars) != 0 {
		return result, cars[0].TotalCount
	} else {
		return result, 0
	}
}
