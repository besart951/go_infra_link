<script lang="ts">
  import type { SpreadsheetDisplayRow, WorksheetPreview } from '$lib/domain/excel/index.js';

  interface Props {
    worksheet: WorksheetPreview | null;
    rows: SpreadsheetDisplayRow[];
    columnLabels: string[];
    isTruncated?: boolean;
    visibleRowLimit: number;
  }

  let { worksheet, rows, columnLabels, isTruncated = false, visibleRowLimit }: Props = $props();
</script>

{#if !worksheet}
  <div class="rounded-lg border border-dashed bg-background p-6 text-sm text-muted-foreground">
    Noch keine Arbeitsmappe geladen.
  </div>
{:else if worksheet.isEmpty}
  <div class="rounded-lg border border-dashed bg-background p-6 text-sm text-muted-foreground">
    Das ausgewählte Arbeitsblatt enthält keine Werte.
  </div>
{:else}
  <div class="rounded-lg border bg-background">
    <div class="flex flex-col gap-1 p-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="min-w-0">
        <h3 class="truncate text-sm font-semibold">{worksheet.name}</h3>
        <p class="text-xs text-muted-foreground">
          {worksheet.rowCount} Zeilen | {worksheet.columnCount} Spalten
        </p>
      </div>
      {#if isTruncated}
        <p class="text-xs text-muted-foreground">
          Vorschau auf die ersten {visibleRowLimit} Zeilen
        </p>
      {/if}
    </div>

    <div class="max-h-[560px] overflow-auto border-t">
      <table class="min-w-max border-separate border-spacing-0 text-xs">
        <thead class="sticky top-0 z-20 bg-muted text-muted-foreground">
          <tr>
            <th class="sticky left-0 z-30 h-8 w-12 border-r border-b bg-muted px-2 text-right">
              #
            </th>
            {#each columnLabels as label}
              <th class="h-8 min-w-32 border-r border-b px-2 text-left font-medium">
                {label}
              </th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each rows as row (row.rowNumber)}
            <tr class="hover:bg-muted/30">
              <th
                class="sticky left-0 z-10 h-8 border-r border-b bg-background px-2 text-right font-medium text-muted-foreground"
              >
                {row.rowNumber}
              </th>
              {#each row.cells as cell, index (`${row.rowNumber}-${index}`)}
                <td
                  class="h-8 max-w-64 truncate border-r border-b px-2 align-middle whitespace-nowrap"
                  title={cell}
                >
                  {cell}
                </td>
              {/each}
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
{/if}
