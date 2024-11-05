package delivery

import "github.com/Inspirate789/ds-lab2/internal/models"

type CarDTO struct {
	ID                 int64          `json:"id"`
	CarUID             string         `json:"car_uid"`
	Brand              string         `json:"brand"`
	Model              string         `json:"model"`
	RegistrationNumber string         `json:"registrationNumber"`
	Power              uint64         `json:"power"`
	Price              uint64         `json:"price"`
	Type               models.CarType `json:"type"`
	Availability       bool           `json:"availability"`
}

func NewCarDTO(car models.Car) CarDTO {
	return CarDTO{
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

type CarsDTO struct {
	Items []CarDTO `json:"items"`
	Count uint64   `json:"count"`
}

func NewCarsDTO(cars []models.Car, totalCount uint64) CarsDTO {
	items := make([]CarDTO, 0, len(cars))

	for _, car := range cars {
		items = append(items, NewCarDTO(car))
	}

	return CarsDTO{
		Items: items,
		Count: totalCount,
	}
}

func (cars CarsDTO) ToModel() []models.Car {
	model := make([]models.Car, 0, len(cars.Items))

	for _, car := range cars.Items {
		model = append(model, car.ToModel())
	}

	return model
}
