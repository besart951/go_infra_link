/**
 * Value object representing search parameters
 */
export interface SearchQuery {
	text: string;
}

/**
 * Factory function to create empty search query
 */
export function createSearchQuery(text = ''): SearchQuery {
	return { text };
}

/**
 * Check if search query is empty
 */
export function isSearchEmpty(query: SearchQuery): boolean {
	return query.text.trim().length === 0;
}
