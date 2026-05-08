import { describe, expect, it } from 'vitest';
import {
  importCellMarkerVisualKind,
  importNodeVisualKind,
  importRowMarkerVisualKind,
  importStatusIconClass
} from './fieldDeviceImportPresentation.js';

describe('field device import presentation', () => {
  it('maps node states to stable icon kinds', () => {
    expect(importNodeVisualKind('pending')).toBe('none');
    expect(importNodeVisualKind('failed')).toBe('failed');
    expect(importNodeVisualKind('delta')).toBe('delta');
    expect(importNodeVisualKind('identical')).toBe('identical');
    expect(importNodeVisualKind('success')).toBe('success');
  });

  it('maps row markers to their presentation equivalent', () => {
    expect(importRowMarkerVisualKind(undefined)).toBe('none');
    expect(importRowMarkerVisualKind({ kind: 'error', messages: [] })).toBe('failed');
    expect(importRowMarkerVisualKind({ kind: 'info', messages: [] })).toBe('identical');
    expect(importRowMarkerVisualKind({ kind: 'delta', messages: [] })).toBe('delta');
    expect(importRowMarkerVisualKind({ kind: 'success', messages: [] })).toBe('success');
  });

  it('maps cell markers by diagnostic severity', () => {
    expect(importCellMarkerVisualKind(undefined)).toBe('none');
    expect(importCellMarkerVisualKind({ severity: 'error', messages: [] })).toBe('failed');
    expect(importCellMarkerVisualKind({ severity: 'warning', messages: [] })).toBe('delta');
  });

  it('keeps icon color classes centralized by presentation kind', () => {
    expect(importStatusIconClass('failed')).toContain('text-destructive');
    expect(importStatusIconClass('delta')).toContain('text-warning');
    expect(importStatusIconClass('identical')).toContain('text-info');
    expect(importStatusIconClass('success')).toContain('text-success');
  });
});
