package facility

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type validationRule func(*domain.ValidationBuilder)

type validationCheck func(*domain.ValidationBuilder) error

func validateRules(rules ...validationRule) error {
	builder := domain.NewValidationBuilder()
	for _, rule := range rules {
		if rule == nil {
			continue
		}
		rule(builder)
	}
	return builder.Err()
}

func validateChecks(checks ...validationCheck) error {
	builder := domain.NewValidationBuilder()
	for _, check := range checks {
		if check == nil {
			continue
		}
		if err := check(builder); err != nil {
			return err
		}
	}
	return builder.Err()
}

func requiredTrimmed(field domain.ValidationField, value string) validationRule {
	return func(builder *domain.ValidationBuilder) {
		field.RequireTrimmed(builder, value)
	}
}

func requiredTrimmedExact(field domain.ValidationField, value string, length int) validationRule {
	return func(builder *domain.ValidationBuilder) {
		field.RequireTrimmedExact(builder, value, length)
	}
}

func requiredTrimmedMax(field domain.ValidationField, value string, max int) validationRule {
	return func(builder *domain.ValidationBuilder) {
		field.RequireTrimmedMax(builder, value, max)
	}
}

func requiredTrimmedPtrMax(field domain.ValidationField, value *string, max int) validationRule {
	return func(builder *domain.ValidationBuilder) {
		trimmed := field.RequireTrimmedPtr(builder, value)
		field.MaxLength(builder, trimmed, max)
	}
}

func requiredUUID(field domain.ValidationField, value uuid.UUID) validationRule {
	return func(builder *domain.ValidationBuilder) {
		field.RequireUUID(builder, value)
	}
}

func requiredNonZero(field domain.ValidationField, value int) validationRule {
	return func(builder *domain.ValidationBuilder) {
		field.RequireNonZero(builder, value)
	}
}

func optionalMaxLength(field domain.ValidationField, value *string, max int) validationRule {
	return func(builder *domain.ValidationBuilder) {
		field.MaxLengthPtr(builder, value, max)
	}
}

func optionalExactLength(field domain.ValidationField, value *string, length int) validationRule {
	return func(builder *domain.ValidationBuilder) {
		if value == nil || strings.TrimSpace(*value) == "" {
			return
		}
		field.ExactLength(builder, *value, length)
	}
}

func addIf(field domain.ValidationField, condition bool, message string) validationRule {
	return func(builder *domain.ValidationBuilder) {
		if condition {
			field.Add(builder, message)
		}
	}
}

func mergeValidation(err error) validationCheck {
	return func(builder *domain.ValidationBuilder) error {
		return builder.Merge(err)
	}
}

func uniqueIfPresent(field domain.ValidationField, value string, exists func() (bool, error)) validationCheck {
	return func(builder *domain.ValidationBuilder) error {
		if strings.TrimSpace(value) == "" {
			return nil
		}
		found, err := exists()
		if err != nil {
			return err
		}
		if found {
			field.Unique(builder)
		}
		return nil
	}
}

func uniqueWithinIfPresent(field domain.ValidationField, scope, value string, exists func() (bool, error)) validationCheck {
	return func(builder *domain.ValidationBuilder) error {
		if strings.TrimSpace(value) == "" {
			return nil
		}
		found, err := exists()
		if err != nil {
			return err
		}
		if found {
			field.UniqueWithin(builder, scope)
		}
		return nil
	}
}

func referenceExists[T any](ctx context.Context, repo domain.Reader[T], id uuid.UUID) validationCheck {
	return func(_ *domain.ValidationBuilder) error {
		return domain.EnsureReferenceExists(ctx, repo, id)
	}
}
