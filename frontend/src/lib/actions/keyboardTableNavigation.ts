type KeyboardTableActivation = 'focus' | 'edit';
type HorizontalDirection = 'left' | 'right';
type VerticalDirection = 'up' | 'down';

type KeyboardTableMove =
  | { axis: 'horizontal'; direction: HorizontalDirection }
  | { axis: 'vertical'; direction: VerticalDirection };

interface KeyboardTableCell {
  element: HTMLElement;
  row: string;
  column: string;
}

interface KeyboardTableCellOptions {
  activate?: KeyboardTableActivation;
}

const cellSelector = '[data-keyboard-table-cell]';
const focusableSelector = [
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  'button:not([disabled])',
  '[href]',
  '[tabindex]:not([tabindex="-1"])'
].join(',');

export function keyboardTableCell(
  row: string,
  column: string,
  options: KeyboardTableCellOptions = {}
): Record<string, string> {
  return {
    'data-keyboard-table-cell': '',
    'data-keyboard-table-row': row,
    'data-keyboard-table-column': column,
    'data-keyboard-table-activate': options.activate ?? 'focus'
  };
}

export function keyboardTableNavigation(root: HTMLElement): { destroy: () => void } {
  function handleKeydown(event: KeyboardEvent): void {
    const target = event.target instanceof Element ? event.target : null;
    if (!target || target.closest('[data-keyboard-table-ignore]')) return;

    const currentCell = target.closest(cellSelector);
    if (!(currentCell instanceof HTMLElement) || !root.contains(currentCell)) return;

    const move = getMoveFromKey(event, target);
    if (!move) return;

    const nextCell = findNextCell(root, currentCell, move);
    if (!nextCell) return;

    event.preventDefault();
    event.stopPropagation();

    queueMicrotask(() => {
      if (root.contains(nextCell)) {
        activateCell(nextCell);
      }
    });
  }

  root.addEventListener('keydown', handleKeydown);

  return {
    destroy: () => root.removeEventListener('keydown', handleKeydown)
  };
}

function getMoveFromKey(event: KeyboardEvent, target: Element): KeyboardTableMove | null {
  if (event.altKey || event.ctrlKey || event.metaKey) return null;

  switch (event.key) {
    case 'Enter':
      if (!isEditableTextTarget(target)) return null;
      return { axis: 'horizontal', direction: event.shiftKey ? 'left' : 'right' };
    case 'Tab':
      return { axis: 'vertical', direction: event.shiftKey ? 'up' : 'down' };
    case 'ArrowRight':
      return { axis: 'horizontal', direction: 'right' };
    case 'ArrowLeft':
      return { axis: 'horizontal', direction: 'left' };
    case 'ArrowDown':
      return { axis: 'vertical', direction: 'down' };
    case 'ArrowUp':
      return { axis: 'vertical', direction: 'up' };
    default:
      return null;
  }
}

function isEditableTextTarget(target: Element): boolean {
  if (target instanceof HTMLTextAreaElement) return true;
  if (!(target instanceof HTMLInputElement)) return false;

  const textInputTypes = new Set([
    '',
    'email',
    'number',
    'password',
    'search',
    'tel',
    'text',
    'url'
  ]);
  return textInputTypes.has(target.type);
}

function findNextCell(
  root: HTMLElement,
  currentCell: HTMLElement,
  move: KeyboardTableMove
): HTMLElement | null {
  const cells = collectCells(root);
  const current = cells.find((cell) => cell.element === currentCell);
  if (!current) return null;

  const rows = groupCellsByRow(cells);
  const rowIndex = rows.findIndex((row) => row.key === current.row);
  if (rowIndex < 0) return null;

  if (move.axis === 'horizontal') {
    return findHorizontalCell(rows, rowIndex, current, move.direction);
  }

  return findVerticalCell(rows, rowIndex, current, move.direction);
}

function collectCells(root: HTMLElement): KeyboardTableCell[] {
  return Array.from(root.querySelectorAll(cellSelector))
    .filter((element): element is HTMLElement => element instanceof HTMLElement)
    .map((element) => ({
      element,
      row: element.dataset.keyboardTableRow ?? '',
      column: element.dataset.keyboardTableColumn ?? ''
    }))
    .filter((cell) => cell.row && cell.column && hasEnabledFocusTarget(cell.element));
}

function groupCellsByRow(
  cells: KeyboardTableCell[]
): Array<{ key: string; cells: KeyboardTableCell[] }> {
  const rows: Array<{ key: string; cells: KeyboardTableCell[] }> = [];

  for (const cell of cells) {
    let row = rows.find((candidate) => candidate.key === cell.row);
    if (!row) {
      row = { key: cell.row, cells: [] };
      rows.push(row);
    }
    row.cells.push(cell);
  }

  return rows;
}

function findHorizontalCell(
  rows: Array<{ key: string; cells: KeyboardTableCell[] }>,
  rowIndex: number,
  current: KeyboardTableCell,
  direction: HorizontalDirection
): HTMLElement | null {
  const rowCells = rows[rowIndex]?.cells ?? [];
  const cellIndex = rowCells.findIndex((cell) => cell.element === current.element);
  if (cellIndex < 0) return null;

  const offset = direction === 'right' ? 1 : -1;
  const nextInRow = rowCells[cellIndex + offset];
  if (nextInRow) return nextInRow.element;

  const adjacentRow = rows[rowIndex + offset];
  if (!adjacentRow) return null;

  const wrappedCell =
    direction === 'right' ? adjacentRow.cells[0] : adjacentRow.cells[adjacentRow.cells.length - 1];
  return wrappedCell?.element ?? null;
}

function findVerticalCell(
  rows: Array<{ key: string; cells: KeyboardTableCell[] }>,
  rowIndex: number,
  current: KeyboardTableCell,
  direction: VerticalDirection
): HTMLElement | null {
  const offset = direction === 'down' ? 1 : -1;
  const targetRow = rows[rowIndex + offset];
  if (!targetRow) return null;

  const exactColumn = targetRow.cells.find((cell) => cell.column === current.column);
  if (exactColumn) return exactColumn.element;

  const currentRowCells = rows[rowIndex]?.cells ?? [];
  const currentColumnIndex = currentRowCells.findIndex((cell) => cell.element === current.element);
  return (
    targetRow.cells[Math.max(0, Math.min(currentColumnIndex, targetRow.cells.length - 1))]
      ?.element ?? null
  );
}

function activateCell(cell: HTMLElement): void {
  const target = findFocusTarget(cell);
  if (!target) return;

  target.focus();

  if (cell.dataset.keyboardTableActivate !== 'edit') return;
  if (target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement) {
    target.select();
    return;
  }
  if (target instanceof HTMLButtonElement) {
    target.click();
  }
}

function hasEnabledFocusTarget(cell: HTMLElement): boolean {
  return findFocusTarget(cell) !== null;
}

function findFocusTarget(cell: HTMLElement): HTMLElement | null {
  const target = cell.matches(focusableSelector)
    ? cell
    : cell.querySelector<HTMLElement>(focusableSelector);

  if (!target || target.getAttribute('aria-disabled') === 'true') return null;
  return target;
}
