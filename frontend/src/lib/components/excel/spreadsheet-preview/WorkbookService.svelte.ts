import type { WorkbookParserPort } from '$lib/domain/ports/excel/workbookParserPort.js';
import type {
  SpreadsheetDisplayRow,
  SpreadsheetWorkbook,
  WorksheetPreview
} from '$lib/domain/excel/index.js';

const SUPPORTED_EXTENSIONS = ['.xlsx', '.xlsm', '.csv'];

export interface WorkbookServiceOptions {
  visibleRowLimit?: number;
}

export class WorkbookService {
  workbook = $state<SpreadsheetWorkbook | null>(null);
  selectedWorksheetName = $state('');
  isLoading = $state(false);
  errorMessage = $state<string | null>(null);

  readonly visibleRowLimit: number;

  sheetNames = $derived.by(() => this.workbook?.sheetNames ?? []);
  selectedWorksheet = $derived.by<WorksheetPreview | null>(
    () => this.workbook?.getWorksheet(this.selectedWorksheetName) ?? null
  );
  columnLabels = $derived.by(() => this.selectedWorksheet?.columnLabels ?? []);
  displayRows = $derived.by<SpreadsheetDisplayRow[]>(
    () => this.selectedWorksheet?.getDisplayRows(this.visibleRowLimit) ?? []
  );
  isPreviewTruncated = $derived.by(
    () => (this.selectedWorksheet?.rowCount ?? 0) > this.visibleRowLimit
  );

  constructor(
    private readonly parser: WorkbookParserPort,
    options: WorkbookServiceOptions = {}
  ) {
    this.visibleRowLimit = options.visibleRowLimit ?? 500;
    this.setupSelectionSync();
  }

  async loadFile(file: File): Promise<void> {
    if (this.isLoading) return;

    this.isLoading = true;
    this.errorMessage = null;

    try {
      this.validateFile(file);

      const workbook = await this.parser.parse(file);
      this.selectedWorksheetName = workbook.firstWorksheetName;
      this.workbook = workbook;

      if (workbook.sheetNames.length === 0) {
        this.errorMessage = 'Die Arbeitsmappe enthält keine Arbeitsblätter.';
      }
    } catch (error) {
      this.workbook = null;
      this.selectedWorksheetName = '';
      this.errorMessage =
        error instanceof Error ? error.message : 'Excel-Datei konnte nicht gelesen werden.';
    } finally {
      this.isLoading = false;
    }
  }

  selectWorksheet(name: string): void {
    if (!this.workbook?.hasWorksheet(name)) return;
    this.selectedWorksheetName = name;
  }

  clear(): void {
    this.workbook = null;
    this.selectedWorksheetName = '';
    this.errorMessage = null;
  }

  private validateFile(file: File): void {
    if (!file) {
      throw new Error('Keine Datei ausgewählt.');
    }

    const fileName = file.name.toLowerCase();
    const supported = SUPPORTED_EXTENSIONS.some((extension) => fileName.endsWith(extension));

    if (!supported) {
      throw new Error(
        'Für die Arbeitsblatt-Vorschau werden nur .xlsx-, .xlsm- und .csv-Dateien unterstützt.'
      );
    }
  }

  private setupSelectionSync(): void {
    $effect(() => {
      const workbook = this.workbook;
      const selectedName = this.selectedWorksheetName;

      if (!workbook) {
        if (selectedName) this.selectedWorksheetName = '';
        return;
      }

      if (!workbook.hasWorksheet(selectedName)) {
        this.selectedWorksheetName = workbook.firstWorksheetName;
      }
    });
  }
}
