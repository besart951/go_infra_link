package domain

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("entity not found")

// ErrConflict is returned when a write cannot be applied due to a conflicting existing entity.
var ErrConflict = errors.New("entity conflict")

// ErrInvalidArgument is returned when the caller provided an invalid request/payload.
var ErrInvalidArgument = errors.New("invalid argument")

// RevisionConflict is returned when a client attempts to update a stale
// entity snapshot. Expected is the revision last seen by the client; Current
// is the authoritative revision observed after the compare-and-swap failed.
type RevisionConflict struct {
	EntityID uuid.UUID
	Expected uint64
	Current  uint64
}

func (e *RevisionConflict) Error() string {
	if e == nil {
		return ErrConflict.Error()
	}
	return fmt.Sprintf(
		"revision conflict for %s: expected %d, current %d",
		e.EntityID,
		e.Expected,
		e.Current,
	)
}

func (*RevisionConflict) Unwrap() error {
	return ErrConflict
}

func AsRevisionConflict(err error) (*RevisionConflict, bool) {
	return errors.AsType[*RevisionConflict](err)
}

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
	return errors.AsType[*ValidationError](err)
}
