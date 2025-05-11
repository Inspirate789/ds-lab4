package errors

type GatewayError string

func (e GatewayError) Error() string {
	return string(e)
}

const (
	ErrInvalidPage GatewayError = "page number must be >= 1"
)

// TODO: use errors.Wrap() ?

func (e GatewayError) Map() map[string]any {
	return map[string]any{"message": e.Error()}
}

func ErrInvalidRentalRequest(msg string) GatewayError {
	return GatewayError("invalid rental request: " + msg)
}

func ErrInvalidDateFrom(msg string) GatewayError {
	return GatewayError("invalid period start date: " + msg)
}

func ErrInvalidDateTo(msg string) GatewayError {
	return GatewayError("invalid period end date: " + msg)
}

func ErrInvalidRentalPeriod(dateFrom, dateTo string) GatewayError {
	return GatewayError("invalid period start date: [" + dateFrom + ", " + dateTo + "]")
}

func ErrRollbackWrap(err error) error {
	if err == nil {
		return nil
	}
	return GatewayError("rollback: " + err.Error())
}
