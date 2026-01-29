export interface Pagination<T> {
	items: T[];
	total: number;
	page: number;
	limit: number;
	total_pages: number;
}
