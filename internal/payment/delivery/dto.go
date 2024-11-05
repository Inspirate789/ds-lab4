package delivery

import "github.com/Inspirate789/ds-lab2/internal/models"

type PaymentDTO struct {
	ID         int64                `json:"id"`
	PaymentUID string               `json:"paymentUid"`
	Status     models.PaymentStatus `json:"status"`
	Price      uint64               `json:"price"`
}

func NewPaymentDTO(car models.Payment) PaymentDTO {
	return PaymentDTO{
		ID:         car.ID,
		PaymentUID: car.PaymentUID,
		Status:     car.Status,
		Price:      car.Price,
	}
}

func (car PaymentDTO) ToModel() models.Payment {
	return models.Payment{
		ID:         car.ID,
		PaymentUID: car.PaymentUID,
		Status:     car.Status,
		Price:      car.Price,
	}
}
