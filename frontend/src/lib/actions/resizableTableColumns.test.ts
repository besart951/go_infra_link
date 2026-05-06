import { describe, expect, it } from 'vitest';
import { resizableTableColumns } from './resizableTableColumns.js';

function mockWidth(cell: HTMLElement, width: number): void {
  Object.defineProperty(cell, 'getBoundingClientRect', {
    configurable: true,
    value: () =>
      ({
        width,
        height: 32,
        top: 0,
        right: width,
        bottom: 32,
        left: 0,
        x: 0,
        y: 0,
        toJSON: () => ({})
      }) satisfies DOMRect
  });
}

describe('resizableTableColumns', () => {
  it('adds handles, resizes the full column when dragged, and restores the previous width on double click', () => {
    document.body.innerHTML = `
      <table>
        <thead>
          <tr>
            <th>First</th>
            <th data-table-resize-min-width="80">Second</th>
            <th data-table-resizable="false">Fixed</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>A1</td>
            <td>B1</td>
            <td>C1</td>
          </tr>
        </tbody>
      </table>
    `;

    const table = document.querySelector('table');
    const heads = Array.from(document.querySelectorAll('th'));
    const bodyCells = Array.from(document.querySelectorAll('td'));

    if (!(table instanceof HTMLTableElement)) {
      throw new Error('expected table');
    }

    mockWidth(heads[0] as HTMLElement, 120);
    mockWidth(heads[1] as HTMLElement, 180);
    mockWidth(heads[2] as HTMLElement, 90);

    const action = resizableTableColumns(table);

    const handles = document.querySelectorAll('[data-table-column-resize-handle]');
    expect(handles).toHaveLength(2);

    handles[1]?.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, clientX: 180 }));
    window.dispatchEvent(new MouseEvent('mousemove', { clientX: 110 }));
    window.dispatchEvent(new MouseEvent('mouseup', { clientX: 110 }));

    expect((heads[1] as HTMLElement).style.width).toBe('110px');
    expect((bodyCells[1] as HTMLElement).style.width).toBe('110px');
    expect((heads[1] as HTMLElement).style.minWidth).toBe('110px');
    expect((bodyCells[1] as HTMLElement).style.maxWidth).toBe('110px');

    handles[1]?.dispatchEvent(new MouseEvent('dblclick', { bubbles: true }));

    expect((heads[1] as HTMLElement).style.width).toBe('180px');
    expect((bodyCells[1] as HTMLElement).style.width).toBe('180px');

    handles[1]?.dispatchEvent(new MouseEvent('dblclick', { bubbles: true }));

    expect((heads[1] as HTMLElement).style.width).toBe('110px');
    expect((bodyCells[1] as HTMLElement).style.width).toBe('110px');

    action.destroy();
  });
});
