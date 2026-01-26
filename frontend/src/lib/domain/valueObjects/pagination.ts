/**
 * Value object representing pagination parameters
 */
export interface Pagination {
	page: number;
	pageSize: number;
}

/**
 * Value object representing paginated response metadata
 */
export interface PaginationMetadata {
	total: number;
	page: number;
	pageSize: number;
	totalPages: number;
}

/**
 * Factory function to create default pagination
 */
export function createPagination(page = 1, pageSize = 10): Pagination {
	return { page, pageSize };
}

/**
 * Calculate total pages from total items and page size
 */
export function calculateTotalPages(total: number, pageSize: number): number {
	return Math.ceil(total / pageSize);
}
