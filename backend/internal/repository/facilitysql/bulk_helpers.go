package facilitysql

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const uuidFilterChunkSize = 5000

func uuidFilterChunks(ids []uuid.UUID, chunkSize int) [][]uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	if chunkSize <= 0 {
		chunkSize = len(ids)
	}

	chunks := make([][]uuid.UUID, 0, (len(ids)+chunkSize-1)/chunkSize)
	for start := 0; start < len(ids); start += chunkSize {
		end := start + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		chunks = append(chunks, ids[start:end])
	}
	return chunks
}

func assignUUIDColumn(
	ctx context.Context,
	db *gorm.DB,
	model any,
	column string,
	assignments map[uuid.UUID]uuid.UUID,
) error {
	if len(assignments) == 0 {
		return nil
	}

	entityIDs := make([]uuid.UUID, 0, len(assignments))
	for entityID, targetID := range assignments {
		if entityID == uuid.Nil || targetID == uuid.Nil {
			return fmt.Errorf("%s assignment contains a nil ID", column)
		}
		entityIDs = append(entityIDs, entityID)
	}
	sort.Slice(entityIDs, func(i, j int) bool {
		return entityIDs[i].String() < entityIDs[j].String()
	})

	now := time.Now().UTC()
	for _, chunk := range uuidFilterChunks(entityIDs, uuidFilterChunkSize) {
		var expression strings.Builder
		expression.WriteString("CASE id")
		args := make([]any, 0, len(chunk)*2)
		for _, entityID := range chunk {
			expression.WriteString(" WHEN ? THEN ?")
			args = append(args, entityID, assignments[entityID])
		}
		expression.WriteString(" ELSE ")
		expression.WriteString(column)
		expression.WriteString(" END")

		if err := db.WithContext(ctx).
			Model(model).
			Where("id IN ?", chunk).
			Updates(map[string]any{
				column:       gorm.Expr(expression.String(), args...),
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
	}
	return nil
}
