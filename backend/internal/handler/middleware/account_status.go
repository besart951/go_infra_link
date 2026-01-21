package middleware

import (
	"net/http"
	"time"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserStatusService interface {
	GetByID(id uuid.UUID) (*domainUser.User, error)
}

func AccountStatusGuard(userService UserStatusService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		usr, err := userService.GetByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "authorization_failed"})
			c.Abort()
			return
		}
		if usr == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if !usr.IsActive || usr.DisabledAt != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "account_disabled"})
			c.Abort()
			return
		}
		if usr.LockedUntil != nil && usr.LockedUntil.After(time.Now().UTC()) {
			c.JSON(http.StatusLocked, gin.H{"error": "account_locked"})
			c.Abort()
			return
		}

		c.Next()
	}
}
