package i18n

import "github.com/gin-gonic/gin"

func RegisterRoutes(publicV1 *gin.RouterGroup, handler *I18nHandler) {
	publicV1.GET("/i18n/:locale", handler.GetTranslations)
}
