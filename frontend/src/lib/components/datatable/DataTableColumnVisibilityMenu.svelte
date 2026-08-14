<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import Settings2 from '@lucide/svelte/icons/settings-2';
  import { isColumnVisible, toggleColumnVisibility } from './dataTableModel.js';
  import type { DataTableColumn, DataTableColumnVisibilityState } from './types.js';

  interface Props<TItem> {
    columns: DataTableColumn<TItem>[];
    columnVisibility: DataTableColumnVisibilityState;
    label: string;
    onChange: (visibility: DataTableColumnVisibilityState) => void;
  }

  let { columns, columnVisibility, label, onChange }: Props<unknown> = $props();

  const hideableColumns = $derived(columns.filter((column) => column.hideable !== false));
</script>

{#if hideableColumns.length > 0}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <Button variant="outline" size="sm" {...props}>
          <Settings2 class="size-4" />
          {label}
        </Button>
      {/snippet}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="end" class="w-48">
      <DropdownMenu.Label>{label}</DropdownMenu.Label>
      <DropdownMenu.Separator />
      {#each hideableColumns as column (column.key)}
        <DropdownMenu.CheckboxItem
          checked={isColumnVisible(column, columnVisibility)}
          onCheckedChange={() => onChange(toggleColumnVisibility(column, columnVisibility))}
        >
          {column.label}
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
