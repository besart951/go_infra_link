# Datatable

Generic Svelte 5 datatable primitives for domain lists that need the field-device table behavior without coupling the UI to facility concepts or a specific backend.

## Shape

- `DataTable.svelte` renders rows, columns, sorting, search, selection, expansion, column visibility, loading, empty, error, and pagination.
- `DataTableController.svelte.ts` owns transport-agnostic state and can call a `DataTableSource`. A source can wrap `fetch`, Wails bindings, SvelteKit load data, or any FastAPI client.
- `dataTableModel.ts` contains pure row processing helpers. Keep filtering, sorting, visibility, and pagination logic here unless the behavior needs DOM state.

## Minimal Usage

```svelte
<script lang="ts">
  import { DataTable, type DataTableColumn } from '$lib/components/datatable';

  type User = { id: string; name: string; email: string };

  const columns: DataTableColumn<User>[] = [
    { key: 'name', label: 'Name', sortable: true },
    { key: 'email', label: 'Email', sortable: true }
  ];

  let query = $state({ search: '', page: 1, pageSize: 25, filters: {}, columnVisibility: {} });
  let selectedIds = $state(new Set<string>());
</script>

<DataTable
  rows={users}
  {columns}
  {query}
  {selectedIds}
  selectable
  onQueryChange={(next) => (query = next)}
  onSelectionChange={(next) => (selectedIds = next)}
/>
```

## Design Notes

The first version intentionally follows common modern table capabilities without adding a large table dependency:

- Controlled query state for server-side sorting, searching, filtering, and pagination.
- Pure client-side fallback for small static lists.
- Snippet-based cells, row actions, toolbar, filters, and expanded rows.
- Selection and expansion as external `Set<string>` state for easy bulk actions.
- Column visibility and resizing hooks that fit the existing table primitives.

Common feature references: TanStack Table features, shadcn-svelte Data Table, MUI X Data Grid sorting/pagination, and AG Grid grouping docs.
