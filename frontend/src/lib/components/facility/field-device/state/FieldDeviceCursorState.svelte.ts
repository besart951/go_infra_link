import type { DataTableQuery } from '$lib/state/table/contracts.js';
import type { FieldDeviceCursorPage } from '$lib/domain/ports/facility/fieldDeviceRepository.js';
import type { FieldDeviceFilters } from './types.js';

export class FieldDeviceCursorState {
  nextCursor = $state<string | undefined>();
  previousCursor = $state<string | undefined>();

  private currentCursor: string | undefined;
  private queryKey = '';

  get hasNextPage(): boolean {
    return Boolean(this.nextCursor);
  }

  get hasPreviousPage(): boolean {
    return Boolean(this.previousCursor);
  }

  cursorFor(query: DataTableQuery<FieldDeviceFilters>, projectId?: string): string | undefined {
    const nextKey = fieldDeviceCursorQueryKey(query, projectId);
    if (nextKey !== this.queryKey) {
      this.reset(nextKey);
    }
    return this.currentCursor;
  }

  apply(page: FieldDeviceCursorPage): void {
    this.nextCursor = page.nextCursor;
    this.previousCursor = page.previousCursor;
  }

  moveNext(): boolean {
    if (!this.nextCursor) return false;
    this.currentCursor = this.nextCursor;
    return true;
  }

  movePrevious(): boolean {
    if (!this.previousCursor) return false;
    this.currentCursor = this.previousCursor;
    return true;
  }

  firstPage(): void {
    this.currentCursor = undefined;
    this.nextCursor = undefined;
    this.previousCursor = undefined;
  }

  private reset(queryKey: string): void {
    this.queryKey = queryKey;
    this.firstPage();
  }
}

export function fieldDeviceCursorQueryKey(
  query: DataTableQuery<FieldDeviceFilters>,
  projectId?: string
): string {
  return JSON.stringify({
    search: query.searchText.trim(),
    orderBy: query.orderBy ?? '',
    order: query.order ?? '',
    filters: Object.entries(query.filters).sort(([left], [right]) => left.localeCompare(right)),
    projectId: projectId ?? ''
  });
}
