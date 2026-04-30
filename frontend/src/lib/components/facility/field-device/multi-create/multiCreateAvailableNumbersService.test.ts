import { describe, expect, it } from 'vitest';
import { autoAssignApparatNumbers } from './multiCreateAvailableNumbersService.js';

describe('multi-create available number helpers', () => {
  it('auto-assigns unused available numbers to blank rows', () => {
    expect(
      autoAssignApparatNumbers(
        [
          { id: 'row-1', bmk: '', description: '', textFix: null, apparatNr: 11 },
          { id: 'row-2', bmk: '', description: '', textFix: null, apparatNr: null }
        ],
        [11, 12]
      )
    ).toEqual({
      rows: [
        { id: 'row-1', bmk: '', description: '', textFix: null, apparatNr: 11 },
        { id: 'row-2', bmk: '', description: '', textFix: null, apparatNr: 12 }
      ],
      changed: true
    });
  });
});
