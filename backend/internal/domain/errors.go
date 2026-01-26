package domain

import "errors"

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("entity not found")

// ErrConflict is returned when a write cannot be applied due to a conflicting existing entity.
var ErrConflict = errors.New("entity conflict")

// ErrInvalidArgument is returned when the caller provided an invalid request/payload.
var ErrInvalidArgument = errors.New("invalid argument")

// ValidationError carries field-level validation details.
// Use dot-separated paths for fields, e.g. "fielddevice.apparat_nr".
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "validation_error"
}

func NewValidationError() *ValidationError {
	return &ValidationError{Fields: map[string]string{}}
}

func (e *ValidationError) Add(field, message string) *ValidationError {
	if e.Fields == nil {
		e.Fields = map[string]string{}
	}
	e.Fields[field] = message
	return e
}

func AsValidationError(err error) (*ValidationError, bool) {
	var ve *ValidationError
	if errors.As(err, &ve) {
		return ve, true
	}
	return nil, false
}
