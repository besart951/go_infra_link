import { read, utils } from 'xlsx';
import {
  SpreadsheetWorkbook,
  WorksheetPreview,
  type SpreadsheetCellValue
} from '$lib/domain/excel/index.js';
import type { WorkbookParserPort } from '$lib/domain/ports/excel/workbookParserPort.js';

export class SheetJsWorkbookParser implements WorkbookParserPort {
  async parse(file: File): Promise<SpreadsheetWorkbook> {
    const workbook = file.name.toLowerCase().endsWith('.csv')
      ? read(await file.text(), {
          type: 'string',
          cellDates: true
        })
      : read(await file.arrayBuffer(), {
          type: 'array',
          cellDates: true
        });

    const worksheets = workbook.SheetNames.map((sheetName) => {
      const sheet = workbook.Sheets[sheetName];
      const rows = sheet
        ? utils.sheet_to_json<SpreadsheetCellValue[]>(sheet, {
            header: 1,
            blankrows: true,
            defval: '',
            raw: true
          })
        : [];

      return new WorksheetPreview(sheetName, rows);
    });

    return new SpreadsheetWorkbook(file.name, file.size, worksheets);
  }
}
