import type { SpreadsheetWorkbook } from '$lib/domain/excel/index.js';

export interface WorkbookParserPort {
  parse(file: File): Promise<SpreadsheetWorkbook>;
}
