export type SpreadsheetCellValue = string | number | boolean | Date | null | undefined;
export type SpreadsheetRow = SpreadsheetCellValue[];

export interface SpreadsheetDisplayRow {
  rowNumber: number;
  cells: string[];
}

function isBlankCell(value: SpreadsheetCellValue): boolean {
  if (value === null || value === undefined) return true;
  if (value instanceof Date) return false;
  return String(value).trim().length === 0;
}

function formatCellValue(value: SpreadsheetCellValue): string {
  if (value === null || value === undefined) return '';
  if (value instanceof Date) return value.toISOString();
  return String(value);
}

export function toSpreadsheetColumnLabel(index: number): string {
  let current = index + 1;
  let label = '';

  while (current > 0) {
    const remainder = (current - 1) % 26;
    label = String.fromCharCode(65 + remainder) + label;
    current = Math.floor((current - 1) / 26);
  }

  return label;
}

function findLastValueRowIndex(rows: readonly (readonly SpreadsheetCellValue[])[]): number {
  for (let rowIndex = rows.length - 1; rowIndex >= 0; rowIndex -= 1) {
    if (rows[rowIndex].some((cell) => !isBlankCell(cell))) {
      return rowIndex;
    }
  }

  return -1;
}

function findLastValueColumnIndex(
  rows: readonly (readonly SpreadsheetCellValue[])[],
  lastRowIndex: number
): number {
  let lastColumnIndex = -1;

  for (let rowIndex = 0; rowIndex <= lastRowIndex; rowIndex += 1) {
    const row = rows[rowIndex];
    for (let columnIndex = row.length - 1; columnIndex >= 0; columnIndex -= 1) {
      if (!isBlankCell(row[columnIndex])) {
        lastColumnIndex = Math.max(lastColumnIndex, columnIndex);
        break;
      }
    }
  }

  return lastColumnIndex;
}

function normalizeRows(rows: readonly (readonly SpreadsheetCellValue[])[]): SpreadsheetRow[] {
  const lastRowIndex = findLastValueRowIndex(rows);
  if (lastRowIndex === -1) return [];

  const lastColumnIndex = findLastValueColumnIndex(rows, lastRowIndex);
  if (lastColumnIndex === -1) return [];

  return rows.slice(0, lastRowIndex + 1).map((row) => {
    const normalized: SpreadsheetRow = [];
    for (let columnIndex = 0; columnIndex <= lastColumnIndex; columnIndex += 1) {
      normalized.push(row[columnIndex] ?? '');
    }
    return normalized;
  });
}

export class WorksheetPreview {
  readonly name: string;
  readonly rows: SpreadsheetRow[];

  constructor(name: string, rows: readonly (readonly SpreadsheetCellValue[])[]) {
    this.name = name;
    this.rows = normalizeRows(rows);
  }

  get rowCount(): number {
    return this.rows.length;
  }

  get columnCount(): number {
    return this.rows[0]?.length ?? 0;
  }

  get isEmpty(): boolean {
    return this.rowCount === 0 || this.columnCount === 0;
  }

  get columnLabels(): string[] {
    return Array.from({ length: this.columnCount }, (_, index) => toSpreadsheetColumnLabel(index));
  }

  getDisplayRows(limit: number): SpreadsheetDisplayRow[] {
    return this.rows.slice(0, limit).map((row, index) => ({
      rowNumber: index + 1,
      cells: row.map(formatCellValue)
    }));
  }
}

export class SpreadsheetWorkbook {
  readonly fileName: string;
  readonly fileSize: number;
  readonly createdAt: string;
  readonly worksheets: WorksheetPreview[];
  private readonly worksheetsByName: Map<string, WorksheetPreview>;

  constructor(fileName: string, fileSize: number, worksheets: WorksheetPreview[]) {
    this.fileName = fileName;
    this.fileSize = fileSize;
    this.createdAt = new Date().toISOString();
    this.worksheets = worksheets;
    this.worksheetsByName = new Map(worksheets.map((worksheet) => [worksheet.name, worksheet]));
  }

  get sheetNames(): string[] {
    return this.worksheets.map((worksheet) => worksheet.name);
  }

  get firstWorksheetName(): string {
    return this.worksheets[0]?.name ?? '';
  }

  hasWorksheet(name: string): boolean {
    return this.worksheetsByName.has(name);
  }

  getWorksheet(name: string): WorksheetPreview | null {
    return this.worksheetsByName.get(name) ?? null;
  }
}
