<script lang="ts">
  import type { SpreadsheetDisplayRow, WorksheetPreview } from '$lib/domain/excel/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { ImportCellMarker } from './fieldDeviceExportImporter.js';

  interface Props {
    worksheet: WorksheetPreview | null;
    rows: SpreadsheetDisplayRow[];
    columnLabels: string[];
    isTruncated?: boolean;
    visibleRowLimit: number;
    cellMarkers?: Record<string, ImportCellMarker>;
  }

  let {
    worksheet,
    rows,
    columnLabels,
    isTruncated = false,
    visibleRowLimit,
    cellMarkers = {}
  }: Props = $props();
  const t = createTranslator();

  function markerFor(rowNumber: number, columnIndex: number): ImportCellMarker | undefined {
    return cellMarkers[`${rowNumber}:${columnIndex}`];
  }

  function markerClass(marker: ImportCellMarker | undefined): string {
    if (!marker) return '';
    return marker.severity === 'error'
      ? 'bg-destructive/15 text-destructive ring-1 ring-inset ring-destructive/40'
      : 'bg-warning-muted text-warning-muted-foreground ring-1 ring-inset ring-warning-border';
  }

  function cellTitle(cell: string, marker: ImportCellMarker | undefined): string {
    if (!marker) return cell;
    return [cell, ...marker.messages].filter(Boolean).join('\n');
  }
</script>

{#if !worksheet}
  <div class="rounded-lg border border-dashed bg-background p-6 text-sm text-muted-foreground">
    {$t('excel.worksheet_preview.grid.no_workbook')}
  </div>
{:else if worksheet.isEmpty}
  <div class="rounded-lg border border-dashed bg-background p-6 text-sm text-muted-foreground">
    {$t('excel.worksheet_preview.grid.empty_worksheet')}
  </div>
{:else}
  <div class="rounded-lg border bg-background">
    <div class="flex flex-col gap-1 p-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="min-w-0">
        <h3 class="truncate text-sm font-semibold">{worksheet.name}</h3>
        <p class="text-xs text-muted-foreground">
          {$t('excel.worksheet_preview.grid.dimensions', {
            rows: worksheet.rowCount,
            columns: worksheet.columnCount
          })}
        </p>
      </div>
      {#if isTruncated}
        <p class="text-xs text-muted-foreground">
          {$t('excel.worksheet_preview.grid.truncated', { count: visibleRowLimit })}
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
                {@const marker = markerFor(row.rowNumber, index)}
                <td
                  class={`h-8 max-w-64 truncate border-r border-b px-2 align-middle whitespace-nowrap ${markerClass(marker)}`}
                  title={cellTitle(cell, marker)}
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
