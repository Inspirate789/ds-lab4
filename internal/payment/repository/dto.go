package repository

import "github.com/Inspirate789/ds-lab2/internal/models"

type PaymentDTO struct {
	ID         int64                `db:"id"`
	PaymentUID string               `db:"payment_uid"`
	Status     models.PaymentStatus `db:"status"`
	Price      uint64               `db:"price"`
}

func (car PaymentDTO) ToModel() models.Payment {
	return models.Payment{
		ID:         car.ID,
		PaymentUID: car.PaymentUID,
		Status:     car.Status,
		Price:      car.Price,
	}
}
