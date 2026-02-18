/// <reference lib="webworker" />

import { read, utils } from 'xlsx';
import type { ExcelSheetPreview } from '$lib/domain/excel/index.js';
import type {
	ExcelReaderWorkerRequest,
	ExcelReaderWorkerResponse
} from './excelReaderWorker.contract.js';

const ctx: DedicatedWorkerGlobalScope = self as DedicatedWorkerGlobalScope;

let cancelled = false;

function postMessageToClient(message: ExcelReaderWorkerResponse): void {
	ctx.postMessage(message);
}

function toCellString(value: unknown): string {
	if (value === null || value === undefined) return '';
	if (typeof value === 'string') return value;
	if (typeof value === 'number' || typeof value === 'boolean') return String(value);
	if (value instanceof Date) return value.toISOString();
	return String(value);
}

function toSheetPreview(sheetName: string, rows: unknown[][]): ExcelSheetPreview {
	const [headerRow = [], ...bodyRows] = rows;
	const headers = headerRow.map((value) => toCellString(value)).filter((value) => value.length > 0);

	const sampleRows = bodyRows.slice(0, 10).map((row) => row.map((value) => toCellString(value)));

	const rowCount = bodyRows.length;
	const columnCount = rows.reduce((maxColumns, row) => Math.max(maxColumns, row.length), 0);

	return {
		name: sheetName,
		rowCount,
		columnCount,
		headers,
		sampleRows
	};
}

async function handleRead(payload: Extract<ExcelReaderWorkerRequest, { type: 'read' }>['payload']) {
	try {
		cancelled = false;

		postMessageToClient({
			type: 'progress',
			payload: {
				percent: 5,
				currentSheet: 0,
				totalSheets: 0,
				message: 'Loading workbook...'
			}
		});

		const workbook = read(payload.buffer, { type: 'array' });
		const totalSheets = workbook.SheetNames.length;

		if (totalSheets === 0) {
			throw new Error('Workbook does not contain sheets.');
		}

		const sheets: ExcelSheetPreview[] = [];

		for (let index = 0; index < totalSheets; index += 1) {
			if (cancelled) {
				postMessageToClient({ type: 'cancelled' });
				return;
			}

			const sheetName = workbook.SheetNames[index];
			const sheet = workbook.Sheets[sheetName];
			const rows = utils.sheet_to_json<unknown[]>(sheet, {
				header: 1,
				blankrows: false,
				defval: ''
			});

			sheets.push(toSheetPreview(sheetName, rows));

			const percent = Math.min(95, 5 + Math.round(((index + 1) / totalSheets) * 90));
			postMessageToClient({
				type: 'progress',
				payload: {
					percent,
					currentSheet: index + 1,
					totalSheets,
					message: `Reading sheet ${index + 1} of ${totalSheets}`
				}
			});

			await Promise.resolve();
		}

		postMessageToClient({
			type: 'done',
			payload: {
				fileName: payload.fileName,
				fileSize: payload.fileSize,
				totalSheets,
				sheets,
				createdAt: new Date().toISOString()
			}
		});

		postMessageToClient({
			type: 'progress',
			payload: {
				percent: 100,
				currentSheet: totalSheets,
				totalSheets,
				message: 'Read session prepared.'
			}
		});
	} catch (error) {
		postMessageToClient({
			type: 'error',
			payload: {
				message: error instanceof Error ? error.message : 'Failed to read Excel file.'
			}
		});
	}
}

ctx.onmessage = (event: MessageEvent<ExcelReaderWorkerRequest>) => {
	const message = event.data;

	if (message.type === 'cancel') {
		cancelled = true;
		postMessageToClient({ type: 'cancelled' });
		return;
	}

	void handleRead(message.payload);
};
