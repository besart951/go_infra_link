package projectsql

import "github.com/google/uuid"

const (
	projectLinkIDFilterChunkSize                = 5000
	projectFieldDeviceSystemTypeFilterChunkSize = 100
)

const deterministicProjectLinkIDExpression = `(
	substr(link_hash.value, 1, 8) || '-' ||
	substr(link_hash.value, 9, 4) || '-' ||
	substr(link_hash.value, 13, 4) || '-' ||
	substr(link_hash.value, 17, 4) || '-' ||
	substr(link_hash.value, 21, 12)
)::uuid`

func projectLinkSeed(kind string, projectID uuid.UUID) string {
	return kind + ":" + projectID.String() + ":"
}

func uuidChunks(ids []uuid.UUID, chunkSize int) [][]uuid.UUID {
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
