package gormbase

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type TrigramSearchColumn struct {
	Expression string
}

func SearchColumn(expression string) TrigramSearchColumn {
	return TrigramSearchColumn{Expression: strings.TrimSpace(expression)}
}

func TrigramSearchCallback[T any](columns ...TrigramSearchColumn) SearchCallback[T] {
	return func(query *gorm.DB, search string) *gorm.DB {
		return ApplyTrigramSearch(query, search, columns...)
	}
}

func ApplyTrigramSearch(query *gorm.DB, search string, columns ...TrigramSearchColumn) *gorm.DB {
	return applyTrigramSearchTerm(query, strings.TrimSpace(search), columns...)
}

func ApplyTrigramTokenSearch(query *gorm.DB, search string, columns ...TrigramSearchColumn) *gorm.DB {
	for token := range strings.FieldsSeq(strings.TrimSpace(search)) {
		query = applyTrigramSearchTerm(query, token, columns...)
	}
	return query
}

func applyTrigramSearchTerm(query *gorm.DB, term string, columns ...TrigramSearchColumn) *gorm.DB {
	term = strings.ToLower(strings.TrimSpace(term))
	if query == nil || term == "" || len(columns) == 0 {
		return query
	}

	condition, args := buildTrigramSearchCondition(query, term, columns...)
	if condition == "" {
		return query
	}
	return query.Where(condition, args...)
}

func buildTrigramSearchCondition(query *gorm.DB, term string, columns ...TrigramSearchColumn) (string, []any) {
	conditions := make([]string, 0, len(columns))
	args := make([]any, 0, len(columns))
	pattern := SearchLikePattern(query, term)

	for _, column := range columns {
		expression := strings.TrimSpace(column.Expression)
		if expression == "" {
			continue
		}

		lowerExpression := fmt.Sprintf("LOWER(%s)", expression)
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", lowerExpression))
		args = append(args, pattern)
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return "(" + strings.Join(conditions, " OR ") + ")", args
}

func SearchLikePattern(query *gorm.DB, term string) string {
	return trigramSearchPattern(term, isPostgresDialect(query))
}

func trigramSearchPattern(term string, postgres bool) string {
	if postgres && len([]rune(term)) < 3 {
		return term + "%"
	}
	return "%" + term + "%"
}

func isPostgresDialect(query *gorm.DB) bool {
	return query != nil && query.Dialector != nil && query.Dialector.Name() == "postgres"
}
