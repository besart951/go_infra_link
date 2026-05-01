package facilitysql

import "github.com/google/uuid"

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
