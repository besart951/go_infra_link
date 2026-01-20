package facility

func normalizePagination(page, limit int) (int, int) {
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}

	return page, limit
}
