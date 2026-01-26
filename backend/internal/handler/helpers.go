package handler

import (
	"github.com/besart951/go_infra_link/backend/internal/handlerutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RespondError(c *gin.Context, status int, code, message string) {
	handlerutil.RespondError(c, status, code, message)
}

func RespondValidationError(c *gin.Context, fields map[string]string) {
	handlerutil.RespondValidationError(c, fields)
}

func RespondNotFound(c *gin.Context, message string) {
	handlerutil.RespondNotFound(c, message)
}

func BindJSON(c *gin.Context, dst any) bool {
	return handlerutil.BindJSON(c, dst)
}

func BindQuery(c *gin.Context, dst any) bool {
	return handlerutil.BindQuery(c, dst)
}

func ParseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	return handlerutil.ParseUUIDParam(c, name)
}

func ParseUUIDParamWithCode(c *gin.Context, name, code string) (uuid.UUID, bool) {
	return handlerutil.ParseUUIDParamWithCode(c, name, code)
}
