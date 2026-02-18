export interface ExcelReadProgress {
	percent: number;
	currentSheet: number;
	totalSheets: number;
	message: string;
}

export interface ExcelSheetPreview {
	name: string;
	rowCount: number;
	columnCount: number;
	headers: string[];
	sampleRows: string[][];
}

export interface ExcelReadSession {
	fileName: string;
	fileSize: number;
	totalSheets: number;
	sheets: ExcelSheetPreview[];
	createdAt: string;
}
