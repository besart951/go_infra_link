import type { ExcelReadProgress, ExcelReadSession } from '$lib/domain/excel/index.js';

export interface ExcelReaderPort {
	read(file: File, onProgress?: (progress: ExcelReadProgress) => void): Promise<ExcelReadSession>;
	cancel(): void;
}
