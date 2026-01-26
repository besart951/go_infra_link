package facility

import "github.com/besart951/go_infra_link/backend/internal/domain"

func normalizePagination(page, limit int) (int, int) {
	return domain.NormalizePagination(page, limit, 10)
}
