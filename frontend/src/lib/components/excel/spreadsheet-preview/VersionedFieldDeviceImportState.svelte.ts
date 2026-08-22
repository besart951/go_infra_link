import {
  isVersionedFieldDeviceArchive,
  isVersionedFieldDeviceWorkbook,
  uploadVersionedFieldDeviceWorkbook,
  type FieldDeviceImportResult
} from '$lib/infrastructure/api/versionedFieldDeviceImport.js';

export interface VersionedFieldDeviceImporter {
  (file: File): Promise<FieldDeviceImportResult>;
}

export class VersionedFieldDeviceImportState {
  file = $state.raw<File | null>(null);
  result = $state.raw<FieldDeviceImportResult | null>(null);
  isImporting = $state(false);
  errorMessage = $state<string | null>(null);
  isVersioned = $state(false);
  isArchive = $derived.by(() => (this.file ? isVersionedFieldDeviceArchive(this.file) : false));

  constructor(
    private readonly importer: VersionedFieldDeviceImporter = uploadVersionedFieldDeviceWorkbook
  ) {}

  select(file: File, sheetNames: readonly string[]): void {
    this.file = file;
    this.result = null;
    this.errorMessage = null;
    this.isVersioned = isVersionedFieldDeviceWorkbook(sheetNames);
  }

  selectArchive(file: File): void {
    this.file = file;
    this.result = null;
    this.errorMessage = null;
    this.isVersioned = isVersionedFieldDeviceArchive(file);
  }

  clear(): void {
    this.file = null;
    this.result = null;
    this.errorMessage = null;
    this.isVersioned = false;
  }

  async run(): Promise<FieldDeviceImportResult | null> {
    if (!this.file || !this.isVersioned || this.isImporting) return null;
    this.isImporting = true;
    this.errorMessage = null;
    try {
      this.result = await this.importer(this.file);
      return this.result;
    } catch (error) {
      this.errorMessage = error instanceof Error ? error.message : String(error);
      return null;
    } finally {
      this.isImporting = false;
    }
  }
}
