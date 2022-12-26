package typemap

import "errors"

func NewNotFoundError(label string) *NotFoundError {
	return &NotFoundError{
		label: label,
	}
}

type NotFoundError struct {
	label string
}

// Error implements the error interface.
func (e *NotFoundError) Error() string {
	return "typemap: " + e.label
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var e *NotFoundError
	return errors.As(err, &e)
}
