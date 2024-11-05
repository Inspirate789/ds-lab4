package models

type PaymentStatus string

const (
	PaymentPaid     PaymentStatus = "PAID"
	PaymentCanceled PaymentStatus = "CANCELED"
)

type Payment struct {
	ID         int64
	PaymentUID string
	Status     PaymentStatus
	Price      uint64
}
