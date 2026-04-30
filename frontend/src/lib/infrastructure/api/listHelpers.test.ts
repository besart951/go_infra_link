import { describe, expect, it } from 'vitest';
import { buildListSearchParams, buildListUrl, mapPaginatedResponse } from './listHelpers.js';

describe('API list helpers', () => {
  it('builds query params for page, limit, search, filters, and skips empty values', () => {
    const query = buildListSearchParams({
      pagination: { page: 2, pageSize: 50 },
      search: { text: 'pump' },
      filters: {
        building_id: 'b-1',
        empty: '',
        nil: null as never,
        missing: undefined as never
      }
    });

    expect(query.toString()).toBe('page=2&limit=50&search=pump&building_id=b-1');
  });

  it('builds list URLs only with a query string when params exist', () => {
    expect(
      buildListUrl('/facility/system-types', {
        pagination: { page: 1, pageSize: 25 },
        search: { text: '' }
      })
    ).toBe('/facility/system-types?page=1&limit=25');
  });

  it('maps API pagination responses to repository pagination metadata', () => {
    expect(
      mapPaginatedResponse(
        {
          items: [{ id: 'item-1' }],
          total: 1,
          page: 1,
          limit: 25,
          total_pages: 1
        },
        {
          pagination: { page: 1, pageSize: 50 },
          search: { text: '' }
        }
      )
    ).toEqual({
      items: [{ id: 'item-1' }],
      metadata: {
        total: 1,
        page: 1,
        pageSize: 25,
        totalPages: 1
      }
    });
  });
});
