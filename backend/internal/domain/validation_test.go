package domain

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestValidationBuilderAndFieldPolicies(t *testing.T) {
	builder := NewValidationBuilder()
	codeField := ValidationField{Key: "building.iws_code", Name: "iws_code"}
	groupField := ValidationField{Key: "building.building_group", Name: "building_group"}

	trimmed := codeField.RequireTrimmed(builder, "  ")
	codeField.ExactLength(builder, "ABC", 4)
	groupField.RequireNonZero(builder, 0)

	if trimmed != "" {
		t.Fatalf("expected trimmed empty string, got %q", trimmed)
	}

	ve, ok := AsValidationError(builder.Err())
	if !ok {
		t.Fatalf("expected validation error, got %v", builder.Err())
	}
	if ve.Fields["building.iws_code"] != "iws_code must be exactly 4 characters" {
		t.Fatalf("expected exact-length message to win for duplicate field, got %+v", ve.Fields)
	}
	if ve.Fields["building.building_group"] != "building_group is required" {
		t.Fatalf("expected required int message, got %+v", ve.Fields)
	}
}

func TestValidationBuilderMergePreservesNonValidationErrors(t *testing.T) {
	builder := NewValidationBuilder()
	builder.Add("left", "left error")
	if err := builder.Merge(errors.New("boom")); err == nil || err.Error() != "boom" {
		t.Fatalf("expected original error, got %v", err)
	}

	other := NewValidationError().Add("right", "right error")
	if err := builder.Merge(other); err != nil {
		t.Fatalf("expected merge to succeed, got %v", err)
	}

	ve, ok := AsValidationError(builder.Err())
	if !ok {
		t.Fatalf("expected validation error, got %v", builder.Err())
	}
	if ve.Fields["left"] != "left error" || ve.Fields["right"] != "right error" {
		t.Fatalf("expected merged fields, got %+v", ve.Fields)
	}
}

func TestEnsureReferenceExistsSkipsNilAndReturnsLookupErrors(t *testing.T) {
	repo := &validationReaderFake{}
	if err := EnsureReferenceExists(context.Background(), repo, uuid.Nil); err != nil {
		t.Fatalf("expected nil id to be ignored, got %v", err)
	}

	id := uuid.New()
	if err := EnsureReferenceExists(context.Background(), repo, id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

type validationReaderFake struct{}

func (validationReaderFake) GetByIds(context.Context, []uuid.UUID) ([]*struct{}, error) {
	return nil, nil
}
