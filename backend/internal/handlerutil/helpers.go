package handlerutil

import (
	"errors"
	"net/http"
	"reflect"
	"sort"
	"strings"

	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/common"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func RespondNotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, "not_found", message)
}

func BindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		if verr := asValidationErrors(err); verr != nil {
			RespondValidationErrorWithFieldErrors(c, http.StatusBadRequest, validationErrorFields(dst, verr), validationFieldErrors(dst, verr))
			return false
		}
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return false
	}
	return true
}

func BindQuery(c *gin.Context, dst any) bool {
	if err := c.ShouldBindQuery(dst); err != nil {
		if verr := asValidationErrors(err); verr != nil {
			RespondValidationErrorWithFieldErrors(c, http.StatusBadRequest, validationErrorFields(dst, verr), validationFieldErrors(dst, verr))
			return false
		}
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return false
	}
	return true
}

func ParseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	idStr := c.Param(name)
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "invalid_id", "Invalid UUID format")
		return uuid.Nil, false
	}
	return id, true
}

func ParseUUIDParamWithCode(c *gin.Context, name, code string) (uuid.UUID, bool) {
	idStr := c.Param(name)
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(c, http.StatusBadRequest, code, "Invalid UUID format")
		return uuid.Nil, false
	}
	return id, true
}

func asValidationErrors(err error) validator.ValidationErrors {
	verr, _ := errors.AsType[validator.ValidationErrors](err)
	return verr
}

func validationErrorFields(dst any, verr validator.ValidationErrors) map[string]string {
	fields := make(map[string]string, len(verr))
	for _, fe := range verr {
		name := validationFieldPath(dst, fe)
		if name == "" {
			name = fe.Field()
		}
		fields[name] = validationMessage(fe)
	}
	return fields
}

func validationFieldErrors(dst any, verr validator.ValidationErrors) []dto.FieldErrorResponse {
	if len(verr) == 0 {
		return nil
	}
	fieldErrors := make([]dto.FieldErrorResponse, 0, len(verr))
	for _, fe := range verr {
		path := validationFieldPath(dst, fe)
		if path == "" {
			path = fe.Field()
		}
		fieldErrors = append(fieldErrors, dto.FieldErrorResponse{
			Path:         path,
			Code:         fe.Tag(),
			Message:      validationMessage(fe),
			LocalizedKey: validationLocalizedKey(fe),
		})
	}
	sort.Slice(fieldErrors, func(i, j int) bool {
		return fieldErrors[i].Path < fieldErrors[j].Path
	})
	return fieldErrors
}

func validationFieldPath(dst any, fe validator.FieldError) string {
	namespace := fe.StructNamespace()
	if namespace == "" {
		namespace = fe.Namespace()
	}
	if namespace == "" {
		return jsonFieldName(reflect.TypeOf(dst), fe.StructField())
	}

	root := indirectType(reflect.TypeOf(dst))
	parts := strings.Split(namespace, ".")
	if len(parts) > 0 && root != nil && parts[0] == root.Name() {
		parts = parts[1:]
	}
	if len(parts) == 0 {
		return jsonFieldName(root, fe.StructField())
	}

	current := root
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		fieldName, indexes := splitNamespacePart(part)
		jsonName := jsonFieldName(current, fieldName)
		if jsonName == "" {
			jsonName = fieldName
		}
		out = append(out, jsonName+indexes)

		if current == nil {
			continue
		}
		if field, ok := current.FieldByName(fieldName); ok {
			current = indirectType(field.Type)
			for current != nil && (current.Kind() == reflect.Slice || current.Kind() == reflect.Array) {
				current = indirectType(current.Elem())
			}
			continue
		}
		current = nil
	}

	return strings.Join(out, ".")
}

func splitNamespacePart(part string) (string, string) {
	index := strings.Index(part, "[")
	if index < 0 {
		return part, ""
	}
	return part[:index], part[index:]
}

func jsonFieldName(t reflect.Type, fieldName string) string {
	t = indirectType(t)
	if t == nil || t.Kind() != reflect.Struct || fieldName == "" {
		return ""
	}
	field, ok := t.FieldByName(fieldName)
	if !ok {
		return ""
	}
	jsonTag := field.Tag.Get("json")
	if jsonTag == "-" {
		return ""
	}
	name, _, _ := strings.Cut(jsonTag, ",")
	if name == "" {
		name = field.Name
	}
	return name
}

func indirectType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "min":
		return "min " + fe.Param()
	case "max":
		return "max " + fe.Param()
	case "len":
		return "length " + fe.Param()
	case "oneof":
		return "must be one of: " + fe.Param()
	case "email":
		return "must be a valid email"
	default:
		return "invalid"
	}
}

func validationLocalizedKey(fe validator.FieldError) string {
	if fe == nil || fe.Tag() == "" {
		return ""
	}
	return "validation." + fe.Tag()
}
