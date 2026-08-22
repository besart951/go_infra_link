package handlerutil

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const IdempotencyKeyHeader = "Idempotency-Key"

// ParseIdempotencyKey returns a stable operation ID. When the client omits the
// key, a fresh ID is generated for the new operation.
func ParseIdempotencyKey(c *gin.Context) (uuid.UUID, error) {
	raw := strings.TrimSpace(c.GetHeader(IdempotencyKeyHeader))
	if raw == "" {
		return uuid.New(), nil
	}
	operationID, err := uuid.Parse(raw)
	if err != nil || operationID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("invalid idempotency key")
	}
	return operationID, nil
}
