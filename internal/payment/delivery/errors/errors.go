package errors

type PaymentError string

func (e PaymentError) Error() string {
	return string(e)
}

func (e PaymentError) Map() map[string]any {
	return map[string]any{"message": e.Error()}
}

const (
	ErrPaymentNotFound    PaymentError = "payment not found"
	ErrPaymentPriceNotSet PaymentError = "payment price not set"
)
