import { fireEvent, render, screen } from '@testing-library/svelte';
import DataTable from './DataTable.svelte';
import type { DataTableColumn } from './types.js';

interface Row {
  id: string;
  name: string;
  count: number;
}

const Table = DataTable as any;

const rows: Row[] = [
  { id: '1', name: 'Valve 10', count: 10 },
  { id: '2', name: 'Valve 2', count: 2 },
  { id: '3', name: 'Pump 1', count: 1 }
];

const columns: DataTableColumn<Row>[] = [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'count', label: 'Count', sortable: true, align: 'right' }
];

describe('DataTable', () => {
  it('filters rows with the built-in search input', async () => {
    render(Table, {
      rows,
      columns,
      pagination: false
    });

    await fireEvent.input(screen.getByRole('searchbox'), {
      target: { value: 'pump' }
    });

    expect(screen.getByText('Pump 1')).toBeInTheDocument();
    expect(screen.queryByText('Valve 10')).not.toBeInTheDocument();
    expect(screen.queryByText('Valve 2')).not.toBeInTheDocument();
  });

  it('sorts rows when a sortable header is clicked', async () => {
    const rendered = render(Table, {
      rows,
      columns,
      pagination: false
    });

    await fireEvent.click(screen.getByRole('button', { name: 'Sort ascending: Name' }));

    const bodyRows = Array.from(rendered.container.querySelectorAll('tbody tr'));
    expect(bodyRows[0]).toHaveTextContent('Pump 1');
    expect(bodyRows[1]).toHaveTextContent('Valve 2');
    expect(bodyRows[2]).toHaveTextContent('Valve 10');
  });

  it('reports row selection changes', async () => {
    const onSelectionChange = vi.fn();
    render(Table, {
      rows,
      columns,
      selectable: true,
      pagination: false,
      onSelectionChange
    });

    const rowCheckbox = screen.getAllByRole('checkbox')[1];
    await fireEvent.click(rowCheckbox);

    const [selectedIds, selectedRows] = onSelectionChange.mock.calls[0];
    expect([...selectedIds]).toEqual(['1']);
    expect(selectedRows).toEqual([rows[0]]);
  });

  it('paginates uncontrolled client-side rows', async () => {
    const pagedRows = Array.from({ length: 26 }, (_, index) => ({
      id: String(index + 1),
      name: `Row ${index + 1}`,
      count: index + 1
    }));

    render(Table, {
      rows: pagedRows,
      columns
    });

    expect(screen.getByText('Row 1')).toBeInTheDocument();
    expect(screen.queryByText('Row 26')).not.toBeInTheDocument();

    await fireEvent.click(screen.getByRole('button', { name: 'Next page' }));

    expect(screen.getByText('Row 26')).toBeInTheDocument();
  });
});
