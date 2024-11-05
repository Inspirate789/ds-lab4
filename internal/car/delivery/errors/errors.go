package errors

type CarError string

func (e CarError) Error() string {
	return string(e)
}

func (e CarError) Map() map[string]any {
	return map[string]any{"message": e.Error()}
}

const (
	ErrCarNotFound    CarError = "car not found"
	ErrCarAlreadyRent CarError = "car already rent"
)
