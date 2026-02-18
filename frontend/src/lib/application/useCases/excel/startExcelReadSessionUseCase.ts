import type { ExcelReadProgress, ExcelReadSession } from '$lib/domain/excel/index.js';
import type { ExcelReaderPort } from '$lib/domain/ports/excel/excelReaderPort.js';

const SUPPORTED_EXCEL_EXTENSIONS = ['.xlsx', '.xls', '.xlsm', '.xlsb'];

export class StartExcelReadSessionUseCase {
	constructor(private readonly excelReader: ExcelReaderPort) {}

	async execute(
		file: File,
		onProgress?: (progress: ExcelReadProgress) => void
	): Promise<ExcelReadSession> {
		this.validateFile(file);
		return this.excelReader.read(file, onProgress);
	}

	cancel(): void {
		this.excelReader.cancel();
	}

	private validateFile(file: File): void {
		if (!file) {
			throw new Error('No file selected.');
		}

		const fileName = file.name.toLowerCase();
		const supported = SUPPORTED_EXCEL_EXTENSIONS.some((ext) => fileName.endsWith(ext));

		if (!supported) {
			throw new Error('Only Excel files (.xlsx, .xls, .xlsm, .xlsb) are supported.');
		}
	}
}
