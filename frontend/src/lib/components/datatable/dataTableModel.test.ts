import {
  createDefaultDataTableQuery,
  getVisibleColumns,
  nextSortState,
  processDataTableRows,
  toggleColumnVisibility
} from './dataTableModel.js';
import type { DataTableColumn } from './types.js';

interface Row {
  id: string;
  name: string;
  count: number;
  hidden?: string;
}

const rows: Row[] = [
  { id: '1', name: 'Valve 10', count: 10, hidden: 'secret' },
  { id: '2', name: 'Valve 2', count: 2 },
  { id: '3', name: 'Pump 1', count: 1 }
];

const columns: DataTableColumn<Row>[] = [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'count', label: 'Count', sortable: true },
  { key: 'hidden', label: 'Hidden', defaultVisible: false }
];

describe('dataTableModel', () => {
  it('filters, sorts, and paginates client-side rows', () => {
    const result = processDataTableRows(
      rows,
      columns,
      createDefaultDataTableQuery({
        search: 'valve',
        sort: { key: 'name', direction: 'asc' },
        page: 1,
        pageSize: 1
      })
    );

    expect(result.filteredRows).toHaveLength(2);
    expect(result.sortedRows.map((row) => row.name)).toEqual(['Valve 2', 'Valve 10']);
    expect(result.rows.map((row) => row.name)).toEqual(['Valve 2']);
    expect(result.totalRows).toBe(2);
    expect(result.totalPages).toBe(2);
  });

  it('leaves rows untouched in manual modes while using provided totals', () => {
    const result = processDataTableRows(
      rows,
      columns,
      createDefaultDataTableQuery({
        search: 'missing',
        sort: { key: 'name', direction: 'desc' },
        page: 2,
        pageSize: 1
      }),
      { search: true, sort: true, pagination: true },
      200,
      20
    );

    expect(result.rows).toEqual(rows);
    expect(result.totalRows).toBe(200);
    expect(result.totalPages).toBe(20);
  });

  it('cycles sort state through ascending, descending, and none', () => {
    const asc = nextSortState(undefined, 'name');
    const desc = nextSortState(asc, 'name');
    const none = nextSortState(desc, 'name');

    expect(asc).toEqual({ key: 'name', direction: 'asc' });
    expect(desc).toEqual({ key: 'name', direction: 'desc' });
    expect(none).toBeUndefined();
  });

  it('respects default-hidden columns and visibility overrides', () => {
    let visibility = {};
    expect(getVisibleColumns(columns, visibility).map((column) => column.key)).toEqual([
      'name',
      'count'
    ]);

    visibility = toggleColumnVisibility(columns[2], visibility);
    expect(getVisibleColumns(columns, visibility).map((column) => column.key)).toEqual([
      'name',
      'count',
      'hidden'
    ]);
  });
});
