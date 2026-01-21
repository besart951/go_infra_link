package domain

import "errors"

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("entity not found")

// ErrConflict is returned when a write cannot be applied due to a conflicting existing entity.
var ErrConflict = errors.New("entity conflict")

// ErrInvalidArgument is returned when the caller provided an invalid request/payload.
var ErrInvalidArgument = errors.New("invalid argument")
