/// <reference lib="webworker" />

import { read, utils } from 'xlsx';
import type { BacnetObjectExcel, ObjectDataExcel } from '$lib/domain/excel/index.js';
import type {
	ExcelReaderWorkerRequest,
	ExcelReaderWorkerResponse
} from './excelReaderWorker.contract.js';

const ctx: DedicatedWorkerGlobalScope = self as DedicatedWorkerGlobalScope;
const TARGET_SHEET_NAME = 'Definition AT Funktionen';
const START_ROW_INDEX = 18;
const SECTOR_SIZE = 50;
const USED_COLUMN_COUNT = 46;

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

type ExcelValue = string | number | boolean;
type ExcelRow = ExcelValue[];

interface BacnetObjectJob {
	row: ExcelRow;
	targetObjectIndex: number;
	isOptional: boolean;
}

function toCellBoolean(value: unknown): boolean {
	if (typeof value === 'boolean') return value;
	if (typeof value === 'number') return value === 1;
	if (typeof value === 'string') {
		const normalized = value.trim().toLowerCase();
		return (
			normalized === '1' || normalized === 'true' || normalized === 'yes' || normalized === 'ja'
		);
	}
	return false;
}

function mapGmsVisible(value: unknown): boolean {
	const normalized = toCellString(value).trim().toLowerCase();
	if (normalized.length === 0) return true;
	if (normalized === 'nein') return false;
	return toCellBoolean(value);
}

function getCell(rows: ExcelRow[], rowIndex: number, colIndex: number): string {
	const row = rows[rowIndex];
	if (!row) return '';
	return toCellString(row[colIndex]).trim();
}

function objectDataIdPart(row: ExcelRow): string {
	return `${toCellString(row[13]).trim()}${toCellString(row[14]).trim()}${toCellString(row[15]).trim()}`;
}

function buildHardwareLabel(prefix: string, value: ExcelValue | undefined): string {
	if (value === null || value === undefined || value === '') return '';

	const numeric = Number(value);
	if (!Number.isNaN(numeric) && Number.isFinite(numeric) && numeric !== 0) {
		const normalized = Math.trunc(numeric);
		return `${prefix}${String(normalized).padStart(2, '0')}`;
	}

	const text = toCellString(value).trim();
	if (text.length === 0) return '';

	const digits = text.match(/\d+/);
	if (digits && digits[0]) {
		return `${prefix}${digits[0].padStart(2, '0')}`;
	}

	return '';
}

function mapHardwareLabel(row: ExcelRow): string {
	const ap = row[41];
	const aq = row[42];
	const ar = row[43];
	const as = row[44];
	const apLabel = buildHardwareLabel('DO', ap);
	if (apLabel.length > 0) return apLabel;
	const aqLabel = buildHardwareLabel('AO', aq);
	if (aqLabel.length > 0) return aqLabel;
	const arLabel = buildHardwareLabel('DI', ar);
	if (arLabel.length > 0) return arLabel;
	const asLabel = buildHardwareLabel('AI', as);
	if (asLabel.length > 0) return asLabel;
	return '';
}

function createBacnetObject(row: ExcelRow, prefix: string, isOptional: boolean): BacnetObjectExcel {
	const alValue = toCellString(row[37]).trim();

	return {
		id: `${prefix}_${objectDataIdPart(row)}_${alValue}`,
		text_fix: toCellString(row[18]).trim(),
		description: toCellString(row[45]).trim(),
		gms_visible: mapGmsVisible(row[6]),
		is_optional: isOptional,
		text_individual: toCellString(row[17]).trim(),
		software_type: alValue.substring(0, 2),
		software_number: alValue.slice(2),
		hardware_label: mapHardwareLabel(row),
		software_reference_label: toCellString(row[38]).trim(),
		state_text_label: toCellString(row[40]).trim(),
		notification_class_label: toCellString(row[4]).trim(),
		alarm_definition_label: toCellString(row[39]).trim(),
		apparat_label: toCellString(row[10]).trim()
	};
}

function findLastValueIndex(values: ExcelRow[]): number {
	for (let index = values.length - 1; index >= 0; index -= 1) {
		if (values[index][13] || values[index][16] || values[index][37]) {
			return index;
		}
	}

	return -1;
}

function getMaxWorkers(): number {
	return Math.max(1, Math.floor(ctx.navigator.hardwareConcurrency || 4));
}

async function mapWithConcurrency<T, R>(
	items: T[],
	concurrency: number,
	mapper: (item: T, index: number) => Promise<R> | R
): Promise<R[]> {
	if (items.length === 0) return [];

	const results = new Array<R>(items.length);
	let nextIndex = 0;
	const runnerCount = Math.max(1, Math.min(concurrency, items.length));

	const runners = Array.from({ length: runnerCount }, async () => {
		while (true) {
			if (cancelled) {
				throw new Error('Read session cancelled.');
			}

			const current = nextIndex;
			nextIndex += 1;

			if (current >= items.length) {
				return;
			}

			results[current] = await mapper(items[current], current);
		}
	});

	await Promise.all(runners);
	return results;
}

async function processRows(
	rows: ExcelRow[],
	basePrefix: string,
	existingData: ObjectDataExcel[],
	currentOptionalFlag: boolean,
	maxWorkers: number
): Promise<{ results: ObjectDataExcel[]; currentOptionalFlag: boolean }> {
	const results = existingData;
	let optionalFlag = currentOptionalFlag;
	const bacnetJobs: BacnetObjectJob[] = [];

	for (const row of rows) {
		const colQRaw = toCellString(row[16]).trim();
		const colQ = colQRaw.toLowerCase();
		const colAL = toCellString(row[37]).trim();

		const isOptionalMarker = colQ.includes('optional') && colAL.length === 0;
		const isNewObjectData = colQRaw.length > 0 && colAL.length === 0 && !isOptionalMarker;

		if (isNewObjectData) {
			optionalFlag = false;
			results.push({
				id: `${basePrefix}_${objectDataIdPart(row)}`,
				description: toCellString(row[2]).trim(),
				is_optional_anchor: false,
				bacnet_objects: []
			});
			continue;
		}

		if (isOptionalMarker) {
			optionalFlag = true;
			continue;
		}

		if (colAL.length > 0 && results.length > 0) {
			bacnetJobs.push({
				row,
				targetObjectIndex: results.length - 1,
				isOptional: optionalFlag
			});
		}
	}

	const resolvedBacnetObjects = await mapWithConcurrency(bacnetJobs, maxWorkers, async (job) => ({
		targetObjectIndex: job.targetObjectIndex,
		bacnetObject: createBacnetObject(job.row, basePrefix, job.isOptional)
	}));

	for (const item of resolvedBacnetObjects) {
		results[item.targetObjectIndex]?.bacnet_objects.push(item.bacnetObject);
	}

	return { results, currentOptionalFlag: optionalFlag };
}

function assembleBasePrefix(rows: ExcelRow[]): string {
	const c16 = getCell(rows, 15, 2);
	const c17 = getCell(rows, 16, 2);
	const c18 = getCell(rows, 17, 2);
	const c2 = getCell(rows, 1, 2);
	return `${c16}_${c17}_${c18}_${c2}`;
}

async function scanSectors(
	rows: ExcelRow[],
	basePrefix: string,
	maxWorkers: number
): Promise<ObjectDataExcel[]> {
	const results: ObjectDataExcel[] = [];
	let optionalFlag = false;
	const totalSectors = Math.max(
		1,
		Math.ceil(Math.max(0, rows.length - START_ROW_INDEX) / SECTOR_SIZE)
	);

	for (
		let sectorIndex = 0, start = START_ROW_INDEX;
		start < rows.length;
		sectorIndex += 1, start += SECTOR_SIZE
	) {
		if (cancelled) {
			throw new Error('Read session cancelled.');
		}

		const sectorRows = rows
			.slice(start, start + SECTOR_SIZE)
			.map((row) => row.slice(0, USED_COLUMN_COUNT) as ExcelRow);
		const lastActiveIndex = findLastValueIndex(sectorRows);

		if (lastActiveIndex === -1) {
			break;
		}

		const processedRows = sectorRows.slice(0, lastActiveIndex + 1);
		const processedResult = await processRows(
			processedRows,
			basePrefix,
			results,
			optionalFlag,
			maxWorkers
		);

		optionalFlag = processedResult.currentOptionalFlag;

		const percent = Math.min(95, 15 + Math.round(((sectorIndex + 1) / totalSectors) * 80));
		postMessageToClient({
			type: 'progress',
			payload: {
				percent,
				currentSheet: 1,
				totalSheets: 1,
				message: `Scanning objects ${sectorIndex + 1}/${totalSectors} (${maxWorkers} workers)`
			}
		});

		await Promise.resolve();
	}

	return results;
}

async function handleRead(payload: Extract<ExcelReaderWorkerRequest, { type: 'read' }>['payload']) {
	try {
		cancelled = false;

		postMessageToClient({
			type: 'progress',
			payload: {
				percent: 5,
				currentSheet: 1,
				totalSheets: 1,
				message: 'Loading workbook...'
			}
		});

		const workbook = read(payload.buffer, { type: 'array' });
		const sheet = workbook.Sheets[TARGET_SHEET_NAME];

		if (!sheet) {
			throw new Error(`Sheet "${TARGET_SHEET_NAME}" was not found.`);
		}

		postMessageToClient({
			type: 'progress',
			payload: {
				percent: 12,
				currentSheet: 1,
				totalSheets: 1,
				message: `Reading sheet "${TARGET_SHEET_NAME}"...`
			}
		});

		const rows = utils.sheet_to_json<ExcelRow>(sheet, {
			header: 1,
			blankrows: true,
			defval: ''
		});

		if (rows.length <= START_ROW_INDEX) {
			throw new Error('No scanner data found in the target sheet.');
		}

		const basePrefix = assembleBasePrefix(rows);
		const maxWorkers = getMaxWorkers();
		const objectDataExcel = await scanSectors(rows, basePrefix, maxWorkers);

		if (cancelled) {
			postMessageToClient({ type: 'cancelled' });
			return;
		}

		postMessageToClient({
			type: 'done',
			payload: {
				fileName: payload.fileName,
				fileSize: payload.fileSize,
				objectDataExcel,
				createdAt: new Date().toISOString()
			}
		});

		postMessageToClient({
			type: 'progress',
			payload: {
				percent: 100,
				currentSheet: 1,
				totalSheets: 1,
				message: 'Read session prepared.'
			}
		});
	} catch (error) {
		if (error instanceof Error && error.message === 'Read session cancelled.') {
			postMessageToClient({ type: 'cancelled' });
			return;
		}

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
		return;
	}

	void handleRead(message.payload);
};
