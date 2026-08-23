package notification

import (
	"errors"
	"testing"
)

func TestEventKeyNormalizesAndRejectsEmptyValues(t *testing.T) {
	key, err := NewEventKey("  project.updated  ")
	if err != nil || key != "project.updated" {
		t.Fatalf("NewEventKey() = %q, %v", key, err)
	}
	if _, err := NewEventKey(" \t "); !errors.Is(err, ErrInvalidEventKey) {
		t.Fatalf("empty key error = %v", err)
	}
}

func TestResourceKindNormalizesAndRejectsEmptyValues(t *testing.T) {
	kind, err := NewResourceKind("  field_device  ")
	if err != nil || kind != "field_device" {
		t.Fatalf("NewResourceKind() = %q, %v", kind, err)
	}
	if _, err := NewResourceKind("\n"); !errors.Is(err, ErrInvalidResourceKind) {
		t.Fatalf("empty kind error = %v", err)
	}
}
