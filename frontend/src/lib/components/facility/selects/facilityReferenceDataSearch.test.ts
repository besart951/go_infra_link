import { matchesApparatSearch, matchesSystemPartSearch } from './facilityReferenceDataSearch.js';
import type { Apparat, SystemPart } from '$lib/domain/facility/index.js';

const apparat: Apparat = {
  id: 'apparat-1',
  short_name: 'AHU',
  name: 'Air Handler',
  description: 'Supply air treatment',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z'
};

const systemPart: SystemPart = {
  id: 'system-part-1',
  short_name: 'SUP',
  name: 'Supply air',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z'
};

describe('facility reference data search', () => {
  it('matches an apparat description', () => {
    expect(matchesApparatSearch(apparat, 'treatment')).toBe(true);
  });

  it('continues to match short names and names', () => {
    expect(matchesApparatSearch(apparat, 'ahu')).toBe(true);
    expect(matchesSystemPartSearch(systemPart, 'supply')).toBe(true);
  });
});
