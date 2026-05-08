import type { TranslationParams } from '$lib/i18n/index.js';
import type { NotificationClass } from '$lib/domain/facility/index.js';
import type {
  FieldDeviceImportReport,
  FieldDeviceImportService,
  ImportNodeStatus
} from './FieldDeviceImportService.svelte.js';
import type {
  FieldDeviceImportDevicePlan,
  ImportCellMarker,
  ImportDiagnostic,
  ImportRowMarker
} from './fieldDeviceExportImporter.js';

type Translate = (key: string, params?: TranslationParams) => string;

export type ImportStatusVisualKind = 'none' | 'failed' | 'delta' | 'success' | 'identical';

type BacnetObjectPlan = FieldDeviceImportDevicePlan['bacnetObjects'][number];

export class FieldDeviceImportViewState {
  constructor(
    readonly service: FieldDeviceImportService,
    private readonly translate: Translate
  ) {}

  get diagnostics(): ImportDiagnostic[] {
    return this.service.allDiagnostics;
  }

  get blockingDiagnostics(): ImportDiagnostic[] {
    return this.diagnostics.filter((diagnostic) => diagnostic.severity === 'error');
  }

  get warningDiagnostics(): ImportDiagnostic[] {
    return this.diagnostics.filter((diagnostic) => diagnostic.severity === 'warning');
  }

  node(key: string): ImportNodeViewState {
    return new ImportNodeViewState(this.service, this.translate, key);
  }

  devicesForSystemType(key: string): FieldDeviceImportDevicePlan[] {
    return this.service.visibleDevicesForSystemType(key);
  }

  inputValue(event: Event): string {
    return (event.currentTarget as HTMLInputElement | HTMLSelectElement).value;
  }

  diagnosticClass(diagnostic: ImportDiagnostic): string {
    return diagnostic.severity === 'error'
      ? 'border-destructive/40 bg-destructive/10 text-destructive'
      : 'border-warning-border bg-warning-muted text-warning-muted-foreground';
  }

  diagnosticEntityLabel(diagnostic: ImportDiagnostic): string {
    const key = `field_device.importer.entities.${diagnostic.entity}`;
    const label = this.translate(key);
    return label === key ? diagnostic.entity : label;
  }

  reportClass(report: FieldDeviceImportReport): string {
    if (report.status === 'success') {
      return 'border-success-border bg-success-muted text-success-muted-foreground';
    }
    if (report.status === 'partial') {
      return 'border-warning-border bg-warning-muted text-warning-muted-foreground';
    }
    return 'border-destructive/40 bg-destructive/10 text-destructive';
  }

  reportMessage(report: FieldDeviceImportReport): string {
    if (report.status === 'success') {
      return this.translate('field_device.importer.report.success', {
        count: report.createdFieldDevices
      });
    }

    if (report.status === 'partial') {
      return this.translate('field_device.importer.report.partial', {
        success: report.createdFieldDevices,
        failed: report.failedFieldDevices
      });
    }

    return report.errorMessage || this.translate('field_device.importer.report.failed');
  }

  notificationClassOptionLabel(nc: Pick<NotificationClass, 'nc' | 'object_description'>): string {
    return [`NC ${nc.nc}`, nc.object_description].filter(Boolean).join(' - ');
  }

  notificationClassMessage(object: BacnetObjectPlan): string {
    if (object.notificationClass.status === 'invalid') {
      return this.translate('field_device.importer.tree.notification.invalid', {
        value: object.notificationClass.raw
      });
    }
    if (object.notificationClass.status === 'missing') {
      return this.translate('field_device.importer.tree.notification.missing', {
        value: object.notificationClass.number ?? object.notificationClass.raw
      });
    }
    if (object.notificationClass.status === 'create') {
      return this.translate('field_device.importer.tree.notification.create_pending', {
        value: object.notificationClass.number
      });
    }
    return '';
  }

  notificationClassVisualKind(object: BacnetObjectPlan): ImportStatusVisualKind {
    if (object.notificationClass.status === 'create') return 'delta';
    if (
      object.notificationClass.status === 'invalid' ||
      object.notificationClass.status === 'missing'
    ) {
      return 'failed';
    }
    return 'none';
  }
}

export class ImportNodeViewState {
  constructor(
    private readonly service: FieldDeviceImportService,
    private readonly translate: Translate,
    readonly key: string
  ) {}

  get status(): ImportNodeStatus {
    return this.service.nodeState(this.key)?.status ?? 'pending';
  }

  get label(): string {
    return this.translate(`field_device.importer.tree.status.${this.status}`);
  }

  get message(): string {
    const diagnostics = this.service.diagnosticsForNode(this.key);
    return diagnostics[0]?.message ?? this.service.nodeState(this.key)?.message ?? '';
  }

  get className(): string {
    return importNodeClass(this.status);
  }

  get badgeClass(): string {
    return importNodeBadgeClass(this.status);
  }

  get visualKind(): ImportStatusVisualKind {
    return importNodeVisualKind(this.status);
  }
}

export function importNodeClass(status: ImportNodeStatus): string {
  if (status === 'success') return 'border-success-border bg-success-muted/70';
  if (status === 'identical') return 'border-info-border bg-info-muted/70';
  if (status === 'delta') return 'border-warning-border bg-warning-muted/70';
  if (status === 'failed') return 'border-destructive/40 bg-destructive/10';
  return 'border-border bg-background';
}

export function importNodeBadgeClass(status: ImportNodeStatus): string {
  if (status === 'success') {
    return 'border-success-border bg-success-muted text-success-muted-foreground';
  }
  if (status === 'identical') {
    return 'border-info-border bg-info-muted text-info-muted-foreground';
  }
  if (status === 'delta') {
    return 'border-warning-border bg-warning-muted text-warning-muted-foreground';
  }
  if (status === 'failed') return 'border-destructive/40 bg-destructive/10 text-destructive';
  return 'border-border bg-muted text-muted-foreground';
}

export function importStatusIconClass(kind: ImportStatusVisualKind): string {
  if (kind === 'success') return 'text-success-muted-foreground';
  if (kind === 'identical') return 'text-info-muted-foreground';
  if (kind === 'delta') return 'text-warning-muted-foreground';
  if (kind === 'failed') return 'text-destructive';
  return 'text-muted-foreground';
}

export function importNodeVisualKind(status: ImportNodeStatus): ImportStatusVisualKind {
  if (status === 'pending') return 'none';
  if (status === 'failed') return 'failed';
  return status;
}

export function importRowMarkerVisualKind(
  marker: ImportRowMarker | undefined
): ImportStatusVisualKind {
  if (!marker) return 'none';
  if (marker.kind === 'error') return 'failed';
  if (marker.kind === 'info') return 'identical';
  return marker.kind;
}

export function importCellMarkerVisualKind(
  marker: ImportCellMarker | undefined
): ImportStatusVisualKind {
  if (!marker) return 'none';
  return marker.severity === 'error' ? 'failed' : 'delta';
}
