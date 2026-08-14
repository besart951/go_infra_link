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
	Fields        map[string]string
	Codes         map[string]string
	LocalizedKeys map[string]string
}

func (e *ValidationError) Error() string {
	return "validation_error"
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		Fields:        map[string]string{},
		Codes:         map[string]string{},
		LocalizedKeys: map[string]string{},
	}
}

func (e *ValidationError) Add(field, message string) *ValidationError {
	return e.AddCode(field, "invalid", message)
}

// AddCode records a stable machine-readable code alongside the human message.
// Field paths use the JSON names accepted by the API.
func (e *ValidationError) AddCode(field, code, message string) *ValidationError {
	if e.Fields == nil {
		e.Fields = map[string]string{}
	}
	if e.Codes == nil {
		e.Codes = map[string]string{}
	}
	e.Fields[field] = message
	e.Codes[field] = code
	return e
}

func (e *ValidationError) AddLocalized(field, code, message, localizedKey string) *ValidationError {
	e.AddCode(field, code, message)
	if e.LocalizedKeys == nil {
		e.LocalizedKeys = map[string]string{}
	}
	e.LocalizedKeys[field] = localizedKey
	return e
}

func AsValidationError(err error) (*ValidationError, bool) {
	return errors.AsType[*ValidationError](err)
}
