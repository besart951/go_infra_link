package domain

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type ValidationBuilder struct {
	err *ValidationError
}

func NewValidationBuilder() *ValidationBuilder {
	return &ValidationBuilder{err: NewValidationError()}
}

func (b *ValidationBuilder) Add(field, message string) *ValidationBuilder {
	if b == nil {
		return b
	}
	b.err = b.err.Add(field, message)
	return b
}

func (b *ValidationBuilder) Merge(err error) error {
	if err == nil {
		return nil
	}
	ve, ok := AsValidationError(err)
	if !ok {
		return err
	}
	for field, message := range ve.Fields {
		b.Add(field, message)
	}
	return nil
}

func (b *ValidationBuilder) Err() error {
	if b == nil || b.err == nil || len(b.err.Fields) == 0 {
		return nil
	}
	return b.err
}

func (b *ValidationBuilder) ValidationError() *ValidationError {
	if b == nil || b.err == nil || len(b.err.Fields) == 0 {
		return nil
	}
	return b.err
}

type ValidationField struct {
	Key  string
	Name string
}

func (f ValidationField) Add(builder *ValidationBuilder, message string) {
	builder.Add(f.Key, message)
}

func (f ValidationField) RequireTrimmed(builder *ValidationBuilder, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		f.Add(builder, f.requiredMessage())
	}
	return trimmed
}

func (f ValidationField) RequireTrimmedExact(builder *ValidationBuilder, value string, length int) string {
	trimmed := f.RequireTrimmed(builder, value)
	f.ExactLength(builder, trimmed, length)
	return trimmed
}

func (f ValidationField) RequireTrimmedMax(builder *ValidationBuilder, value string, max int) string {
	trimmed := f.RequireTrimmed(builder, value)
	f.MaxLength(builder, trimmed, max)
	return trimmed
}

func (f ValidationField) RequireTrimmedPtr(builder *ValidationBuilder, value *string) string {
	if value == nil {
		f.Add(builder, f.requiredMessage())
		return ""
	}
	return f.RequireTrimmed(builder, *value)
}

func (f ValidationField) RequireUUID(builder *ValidationBuilder, value uuid.UUID) {
	if value == uuid.Nil {
		f.Add(builder, f.requiredMessage())
	}
}

func (f ValidationField) RequireNonZero(builder *ValidationBuilder, value int) {
	if value == 0 {
		f.Add(builder, f.requiredMessage())
	}
}

func (f ValidationField) ExactLength(builder *ValidationBuilder, value string, length int) {
	if value != "" && len(value) != length {
		f.Add(builder, fmt.Sprintf("%s must be exactly %d characters", f.Name, length))
	}
}

func (f ValidationField) ExactLengthPtr(builder *ValidationBuilder, value *string, length int) {
	if value == nil {
		return
	}
	f.ExactLength(builder, *value, length)
}

func (f ValidationField) MaxLength(builder *ValidationBuilder, value string, max int) {
	if value != "" && len(value) > max {
		f.Add(builder, fmt.Sprintf("%s must be at most %d characters", f.Name, max))
	}
}

func (f ValidationField) MaxLengthPtr(builder *ValidationBuilder, value *string, max int) {
	if value == nil {
		return
	}
	f.MaxLength(builder, *value, max)
}

func (f ValidationField) Between(builder *ValidationBuilder, value, min, max int) {
	if value != 0 && (value < min || value > max) {
		f.Add(builder, fmt.Sprintf("%s must be between %d and %d", f.Name, min, max))
	}
}

func (f ValidationField) Unique(builder *ValidationBuilder) {
	f.Add(builder, f.uniqueMessage())
}

func (f ValidationField) UniqueWithin(builder *ValidationBuilder, scope string) {
	f.Add(builder, fmt.Sprintf("%s must be unique within %s", f.Name, scope))
}

func (f ValidationField) UniqueError() error {
	return NewValidationError().Add(f.Key, f.uniqueMessage())
}

func (f ValidationField) UniqueWithinError(scope string) error {
	return NewValidationError().Add(f.Key, fmt.Sprintf("%s must be unique within %s", f.Name, scope))
}

func (f ValidationField) requiredMessage() string {
	return f.Name + " is required"
}

func (f ValidationField) uniqueMessage() string {
	return f.Name + " must be unique"
}

func EnsureReferenceExists[T any](ctx context.Context, repo Reader[T], id uuid.UUID) error {
	if id == uuid.Nil {
		return nil
	}
	_, err := GetByID(ctx, repo, id)
	return err
}
