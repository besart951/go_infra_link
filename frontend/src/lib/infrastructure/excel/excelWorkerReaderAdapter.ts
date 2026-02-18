import type { ExcelReadProgress, ExcelReadSession } from '$lib/domain/excel/index.js';
import type { ExcelReaderPort } from '$lib/domain/ports/excel/excelReaderPort.js';
import type {
	ExcelReaderWorkerRequest,
	ExcelReaderWorkerResponse
} from './excelReaderWorker.contract.js';

export class ExcelWorkerReaderAdapter implements ExcelReaderPort {
	private worker: Worker | null = null;
	private rejectCurrent: ((reason?: unknown) => void) | null = null;

	async read(
		file: File,
		onProgress?: (progress: ExcelReadProgress) => void
	): Promise<ExcelReadSession> {
		this.cancel();

		const buffer = await file.arrayBuffer();
		const worker = new Worker(new URL('./excelReader.worker.ts', import.meta.url), {
			type: 'module'
		});
		this.worker = worker;

		return new Promise<ExcelReadSession>((resolve, reject) => {
			this.rejectCurrent = reject;

			worker.onmessage = (event: MessageEvent<ExcelReaderWorkerResponse>) => {
				const response = event.data;

				if (response.type === 'progress') {
					onProgress?.(response.payload);
					return;
				}

				if (response.type === 'done') {
					this.disposeWorker();
					resolve(response.payload);
					return;
				}

				if (response.type === 'cancelled') {
					this.disposeWorker();
					reject(new Error('Read session cancelled.'));
					return;
				}

				this.disposeWorker();
				reject(new Error(response.payload.message));
			};

			worker.onerror = (event) => {
				this.disposeWorker();
				reject(new Error(event.message || 'Excel worker failed.'));
			};

			const request: ExcelReaderWorkerRequest = {
				type: 'read',
				payload: {
					fileName: file.name,
					fileSize: file.size,
					buffer
				}
			};

			worker.postMessage(request, [buffer]);
		});
	}

	cancel(): void {
		if (!this.worker) return;
		const reject = this.rejectCurrent;
		this.rejectCurrent = null;
		reject?.(new Error('Read session cancelled.'));

		const cancelRequest: ExcelReaderWorkerRequest = { type: 'cancel' };
		this.worker.postMessage(cancelRequest);
		this.disposeWorker();
	}

	private disposeWorker(): void {
		if (!this.worker) return;
		this.worker.terminate();
		this.worker = null;
		this.rejectCurrent = null;
	}
}
