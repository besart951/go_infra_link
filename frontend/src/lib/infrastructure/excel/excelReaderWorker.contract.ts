import type { ExcelReadProgress, ExcelReadSession } from '$lib/domain/excel/index.js';

export type ExcelReaderWorkerRequest =
	| {
			type: 'read';
			payload: {
				fileName: string;
				fileSize: number;
				buffer: ArrayBuffer;
			};
	  }
	| { type: 'cancel' };

export type ExcelReaderWorkerResponse =
	| {
			type: 'progress';
			payload: ExcelReadProgress;
	  }
	| {
			type: 'done';
			payload: ExcelReadSession;
	  }
	| {
			type: 'error';
			payload: { message: string };
	  }
	| {
			type: 'cancelled';
	  };
