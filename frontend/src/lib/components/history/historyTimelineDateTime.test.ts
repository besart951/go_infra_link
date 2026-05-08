import { describe, expect, it } from 'vitest';
import { buildTimelineDateTimeISO } from './historyTimelineDateTime.js';

describe('buildTimelineDateTimeISO', () => {
  it('returns undefined when no date is provided', () => {
    expect(buildTimelineDateTimeISO('', '09:30', 'start')).toBeUndefined();
  });

  it('defaults the start boundary to the beginning of the day', () => {
    expect(buildTimelineDateTimeISO('2026-05-06', '', 'start')).toBe(
      new Date('2026-05-06T00:00').toISOString()
    );
  });

  it('defaults the end boundary to the end of the day minute range', () => {
    expect(buildTimelineDateTimeISO('2026-05-06', '', 'end')).toBe(
      new Date('2026-05-06T23:59').toISOString()
    );
  });

  it('uses the provided time when present', () => {
    expect(buildTimelineDateTimeISO('2026-05-06', '14:45', 'end')).toBe(
      new Date('2026-05-06T14:45').toISOString()
    );
  });
});
