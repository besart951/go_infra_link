const RESIZE_HANDLE_SELECTOR = '[data-table-column-resize-handle]';
const DEFAULT_MIN_WIDTH = 64;

interface LeafColumn {
  head: HTMLTableCellElement;
  index: number;
  minWidth: number;
}

export function resizableTableColumns(table: HTMLTableElement): { destroy: () => void } {
  let insertedHandles = new Set<HTMLElement>();
  let restoreBodyCursor: string | null = null;
  let restoreBodyUserSelect: string | null = null;
  const previousWidths = new WeakMap<HTMLTableCellElement, number>();

  function getLeafColumns(): LeafColumn[] {
    const headerRows = table.tHead?.rows;
    const headerRow = headerRows?.[headerRows.length - 1];
    if (!headerRow) return [];

    const columns: LeafColumn[] = [];
    let currentIndex = 0;

    for (const cell of Array.from(headerRow.cells)) {
      const span = cell.colSpan || 1;

      if (span === 1) {
        const head = cell as HTMLTableCellElement;
        const minWidth = Number.parseFloat(head.dataset.tableResizeMinWidth ?? '');
        columns.push({
          head,
          index: currentIndex,
          minWidth: Number.isFinite(minWidth) ? minWidth : DEFAULT_MIN_WIDTH
        });
      }

      currentIndex += span;
    }

    return columns;
  }

  function getColumnCells(columnIndex: number): HTMLTableCellElement[] {
    const cells: HTMLTableCellElement[] = [];

    for (const row of Array.from(table.rows)) {
      let currentIndex = 0;

      for (const cell of Array.from(row.cells)) {
        const span = cell.colSpan || 1;

        if (span === 1 && currentIndex === columnIndex) {
          cells.push(cell as HTMLTableCellElement);
          break;
        }

        currentIndex += span;

        if (currentIndex > columnIndex) {
          break;
        }
      }
    }

    return cells;
  }

  function readColumnWidth(cell: HTMLTableCellElement): number {
    const inlineWidth = Number.parseFloat(cell.style.width);
    if (Number.isFinite(inlineWidth) && inlineWidth > 0) return inlineWidth;

    const rectWidth = cell.getBoundingClientRect().width;
    if (rectWidth > 0) return rectWidth;

    const computedWidth = Number.parseFloat(window.getComputedStyle(cell).width);
    if (Number.isFinite(computedWidth) && computedWidth > 0) return computedWidth;

    return cell.offsetWidth;
  }

  function applyColumnWidth(columnIndex: number, width: number): void {
    const nextWidth = `${Math.round(width)}px`;

    for (const cell of getColumnCells(columnIndex)) {
      cell.style.width = nextWidth;
      cell.style.minWidth = nextWidth;
      cell.style.maxWidth = nextWidth;
    }
  }

  function freezeCurrentWidths(columns: LeafColumn[]): void {
    for (const column of columns) {
      applyColumnWidth(column.index, readColumnWidth(column.head));
    }
  }

  function shouldResize(head: HTMLTableCellElement): boolean {
    return head.dataset.tableResizable !== 'false';
  }

  function getActiveColumnFromHandle(handle: HTMLElement): LeafColumn | undefined {
    const head = handle.parentElement;
    if (!(head instanceof HTMLTableCellElement)) return undefined;

    return getLeafColumns().find((column) => column.head === head);
  }

  function ensureHandle(head: HTMLTableCellElement): void {
    if (!shouldResize(head)) return;
    if (head.querySelector(RESIZE_HANDLE_SELECTOR)) return;

    if (window.getComputedStyle(head).position === 'static') {
      head.style.position = 'relative';
    }

    const handle = document.createElement('span');
    handle.dataset.tableColumnResizeHandle = 'true';
    handle.setAttribute('aria-hidden', 'true');
    handle.className =
      'absolute top-0 right-0 z-10 h-full w-3 cursor-col-resize select-none touch-none bg-transparent';

    head.append(handle);
    insertedHandles.add(handle);
  }

  function removeDanglingHandles(columns: LeafColumn[]): void {
    const activeHeads = new Set(columns.map((column) => column.head));

    for (const handle of insertedHandles) {
      const head = handle.parentElement;

      if (
        !(head instanceof HTMLTableCellElement) ||
        !activeHeads.has(head) ||
        !shouldResize(head)
      ) {
        handle.remove();
        insertedHandles.delete(handle);
      }
    }
  }

  function refreshHandles(): void {
    const columns = getLeafColumns();
    removeDanglingHandles(columns);

    for (const column of columns) {
      ensureHandle(column.head);
    }
  }

  function cleanupDragListeners(
    moveListener: (event: MouseEvent) => void,
    upListener: () => void
  ): void {
    window.removeEventListener('mousemove', moveListener);
    window.removeEventListener('mouseup', upListener);

    if (restoreBodyCursor !== null) {
      document.body.style.cursor = restoreBodyCursor;
      restoreBodyCursor = null;
    }

    if (restoreBodyUserSelect !== null) {
      document.body.style.userSelect = restoreBodyUserSelect;
      restoreBodyUserSelect = null;
    }
  }

  function handleMouseDown(event: MouseEvent): void {
    const target = event.target;
    if (!(target instanceof HTMLElement)) return;

    const handle = target.closest(RESIZE_HANDLE_SELECTOR);
    if (!(handle instanceof HTMLElement) || !table.contains(handle)) return;

    const activeColumn = getActiveColumnFromHandle(handle);
    if (!activeColumn) return;

    event.preventDefault();
    event.stopPropagation();

    const columns = getLeafColumns();

    freezeCurrentWidths(columns);

    const startX = event.clientX;
    const startWidth = readColumnWidth(activeColumn.head);

    restoreBodyCursor = document.body.style.cursor;
    restoreBodyUserSelect = document.body.style.userSelect;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';

    const moveListener = (moveEvent: MouseEvent) => {
      const delta = moveEvent.clientX - startX;
      const nextWidth = Math.max(activeColumn.minWidth, startWidth + delta);
      applyColumnWidth(activeColumn.index, nextWidth);
    };

    const upListener = () => {
      const endWidth = readColumnWidth(activeColumn.head);
      if (Math.round(endWidth) !== Math.round(startWidth)) {
        previousWidths.set(activeColumn.head, startWidth);
      }

      cleanupDragListeners(moveListener, upListener);
    };

    window.addEventListener('mousemove', moveListener);
    window.addEventListener('mouseup', upListener, { once: true });
  }

  function handleDoubleClick(event: MouseEvent): void {
    const target = event.target;
    if (!(target instanceof HTMLElement)) return;

    const handle = target.closest(RESIZE_HANDLE_SELECTOR);
    if (!(handle instanceof HTMLElement) || !table.contains(handle)) return;

    const activeColumn = getActiveColumnFromHandle(handle);
    if (!activeColumn) return;

    const previousWidth = previousWidths.get(activeColumn.head);
    if (!previousWidth) return;

    event.preventDefault();
    event.stopPropagation();

    const currentWidth = readColumnWidth(activeColumn.head);
    applyColumnWidth(activeColumn.index, previousWidth);
    previousWidths.set(activeColumn.head, currentWidth);
  }

  const observer = new MutationObserver(() => {
    refreshHandles();
  });

  refreshHandles();
  observer.observe(table, { childList: true, subtree: true });
  table.addEventListener('mousedown', handleMouseDown);
  table.addEventListener('dblclick', handleDoubleClick);

  return {
    destroy() {
      observer.disconnect();
      table.removeEventListener('mousedown', handleMouseDown);
      table.removeEventListener('dblclick', handleDoubleClick);

      for (const handle of insertedHandles) {
        handle.remove();
      }
      insertedHandles.clear();
    }
  };
}
