package errors

type RentalError string

func (e RentalError) Error() string {
	return string(e)
}

func (e RentalError) Map() map[string]any {
	return map[string]any{"message": e.Error()}
}

const (
	ErrRentalNotFound       RentalError = "rental not found"
	ErrRentalNotPermitted   RentalError = "rental not permitted"
	ErrInvalidRentalRequest RentalError = "invalid rental request"
	ErrConvertRentalRequest RentalError = "rental request conversion failed"
)
