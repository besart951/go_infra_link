package handlerutil

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CopyOperationIDHeader = "X-Copy-Operation-ID"

// ParseCopyOperationID reads the client-generated idempotency key for a
// long-running copy. Older clients without the header still receive a new job.
func ParseCopyOperationID(c *gin.Context) (uuid.UUID, error) {
	raw := strings.TrimSpace(c.GetHeader(CopyOperationIDHeader))
	if raw == "" {
		return uuid.New(), nil
	}
	operationID, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid copy operation ID: %w", err)
	}
	if operationID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("invalid copy operation ID: must not be nil")
	}
	return operationID, nil
}
