package notification

import (
	"errors"
	"strings"
)

var (
	ErrInvalidEventKey     = errors.New("event key is required")
	ErrInvalidResourceKind = errors.New("resource kind is required")
)

type EventKey string

func NewEventKey(value string) (EventKey, error) {
	normalized := EventKey(strings.TrimSpace(value))
	if normalized == "" {
		return "", ErrInvalidEventKey
	}
	return normalized, nil
}

func (key EventKey) Normalize() (EventKey, error) {
	return NewEventKey(string(key))
}

func (key EventKey) String() string { return string(key) }

type ResourceKind string

func NewResourceKind(value string) (ResourceKind, error) {
	normalized := ResourceKind(strings.TrimSpace(value))
	if normalized == "" {
		return "", ErrInvalidResourceKind
	}
	return normalized, nil
}

func (kind ResourceKind) Normalize() (ResourceKind, error) {
	return NewResourceKind(string(kind))
}

func (kind ResourceKind) String() string { return string(kind) }
