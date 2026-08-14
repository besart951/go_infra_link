import { DataTableController } from './DataTableController.svelte.js';
import type { DataTableSource } from './types.js';

interface Row {
  id: string;
  name: string;
}

describe('DataTableController', () => {
  it('loads rows through a transport-agnostic source', async () => {
    const source: DataTableSource<Row> = {
      load: vi.fn().mockResolvedValue({
        rows: [{ id: '1', name: 'Pump' }],
        totalRows: 1,
        page: 1,
        pageSize: 25
      })
    };

    const controller = new DataTableController<Row>({ source });

    await controller.load();

    expect(source.load).toHaveBeenCalledWith(
      {
        search: '',
        sort: undefined,
        pagination: { page: 1, pageSize: 25 },
        filters: {}
      },
      expect.any(AbortSignal)
    );
    expect(controller.rows).toEqual([{ id: '1', name: 'Pump' }]);
    expect(controller.totalRows).toBe(1);
    expect(controller.loading).toBe(false);
  });

  it('resets to the first page when search changes', async () => {
    const source: DataTableSource<Row> = {
      load: vi.fn().mockResolvedValue({
        rows: [],
        totalRows: 0
      })
    };

    const controller = new DataTableController<Row>({
      source,
      initialQuery: { page: 3, pageSize: 50 }
    });

    await controller.setSearch('sps');

    expect(controller.query.search).toBe('sps');
    expect(controller.query.page).toBe(1);
    expect(source.load).toHaveBeenLastCalledWith(
      expect.objectContaining({
        search: 'sps',
        pagination: { page: 1, pageSize: 50 }
      }),
      expect.any(AbortSignal)
    );
  });

  it('prunes selected ids after loading a new row set when getRowId is provided', async () => {
    const source: DataTableSource<Row> = {
      load: vi.fn().mockResolvedValue({
        rows: [{ id: 'live', name: 'Live' }],
        totalRows: 1
      })
    };

    const controller = new DataTableController<Row>({
      source,
      getRowId: (row) => row.id
    });
    controller.setSelection(new Set(['live', 'stale']));

    await controller.load();

    expect([...controller.selectedIds]).toEqual(['live']);
  });
});
