package handlerutil

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/handler/dto"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func RespondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, dto.ErrorResponse{
		Error:   code,
		Message: message,
	})
}

func RespondValidationError(c *gin.Context, fields map[string]string) {
	c.JSON(http.StatusBadRequest, dto.ErrorResponse{
		Error:  "validation_error",
		Fields: fields,
	})
}

func RespondNotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, "not_found", message)
}

func BindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		if verr := asValidationErrors(err); verr != nil {
			RespondValidationError(c, validationErrorFields(dst, verr))
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
			RespondValidationError(c, validationErrorFields(dst, verr))
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
	var verr validator.ValidationErrors
	if errors.As(err, &verr) {
		return verr
	}
	return nil
}

func validationErrorFields(dst any, verr validator.ValidationErrors) map[string]string {
	fields := make(map[string]string, len(verr))
	fieldNames := structJSONFieldMap(dst)
	for _, fe := range verr {
		name := fe.StructField()
		if mapped, ok := fieldNames[name]; ok && mapped != "" {
			name = mapped
		}
		fields[name] = validationMessage(fe)
	}
	return fields
}

func structJSONFieldMap(dst any) map[string]string {
	result := map[string]string{}
	if dst == nil {
		return result
	}
	t := reflect.TypeOf(dst)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return result
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		name := strings.Split(jsonTag, ",")[0]
		if name == "" {
			name = field.Name
		}
		result[field.Name] = name
	}
	return result
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
