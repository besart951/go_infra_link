<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import type { WorksheetPreview } from '$lib/domain/excel/index.js';
  import {
    AlertTriangle,
    CheckCircle2,
    GitBranch,
    LoaderCircle,
    Play,
    WandSparkles,
    XCircle
  } from '@lucide/svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import {
    CONTROL_CABINET_IMPORT_NODE_KEY,
    SPS_CONTROLLER_IMPORT_NODE_KEY,
    type FieldDeviceImportService,
    type ImportNodeStatus
  } from './FieldDeviceImportService.svelte.js';
  import type {
    FieldDeviceImportDevicePlan,
    ImportDiagnostic
  } from './fieldDeviceExportImporter.js';
  import type { FieldDeviceImportReport } from './FieldDeviceImportService.svelte.js';

  interface Props {
    worksheet: WorksheetPreview | null;
    service: FieldDeviceImportService;
  }

  let { worksheet, service }: Props = $props();
  const t = createTranslator();

  const diagnostics = $derived(service.allDiagnostics);
  const blockingDiagnostics = $derived(
    diagnostics.filter((diagnostic) => diagnostic.severity === 'error')
  );
  const warningDiagnostics = $derived(
    diagnostics.filter((diagnostic) => diagnostic.severity === 'warning')
  );

  function devicesForSystemType(key: string): FieldDeviceImportDevicePlan[] {
    return (
      service.plan?.controller.fieldDevices.filter(
        (device) => device.spsControllerSystemTypeKey === key
      ) ?? []
    );
  }

  function diagnosticClass(diagnostic: ImportDiagnostic): string {
    return diagnostic.severity === 'error'
      ? 'border-destructive/40 bg-destructive/10 text-destructive'
      : 'border-warning-border bg-warning-muted text-warning-muted-foreground';
  }

  function reportClass(status: string): string {
    if (status === 'success')
      return 'border-success-border bg-success-muted text-success-muted-foreground';
    if (status === 'partial')
      return 'border-warning-border bg-warning-muted text-warning-muted-foreground';
    return 'border-destructive/40 bg-destructive/10 text-destructive';
  }

  function reportMessage(report: FieldDeviceImportReport): string {
    if (report.status === 'success') {
      return $t('field_device.importer.report.success', {
        count: report.createdFieldDevices
      });
    }

    if (report.status === 'partial') {
      return $t('field_device.importer.report.partial', {
        success: report.createdFieldDevices,
        failed: report.failedFieldDevices
      });
    }

    return report.errorMessage || $t('field_device.importer.report.failed');
  }

  function diagnosticEntityLabel(diagnostic: ImportDiagnostic): string {
    const key = `field_device.importer.entities.${diagnostic.entity}`;
    const label = $t(key);
    return label === key ? diagnostic.entity : label;
  }

  function nodeStatus(key: string): ImportNodeStatus {
    return service.nodeState(key)?.status ?? 'pending';
  }

  function nodeClass(key: string): string {
    const status = nodeStatus(key);
    if (status === 'success' || status === 'existing') {
      return 'border-success-border bg-success-muted/70';
    }
    if (status === 'failed') return 'border-destructive/40 bg-destructive/10';
    return 'border-border bg-background';
  }

  function nodeBadgeClass(key: string): string {
    const status = nodeStatus(key);
    if (status === 'success' || status === 'existing') {
      return 'border-success-border bg-success-muted text-success-muted-foreground';
    }
    if (status === 'failed') return 'border-destructive/40 bg-destructive/10 text-destructive';
    return 'border-border bg-muted text-muted-foreground';
  }

  function nodeStatusLabel(key: string): string {
    return $t(`field_device.importer.tree.status.${nodeStatus(key)}`);
  }

  function nodeMessage(key: string): string {
    const diagnostics = service.diagnosticsForNode(key);
    return diagnostics[0]?.message ?? service.nodeState(key)?.message ?? '';
  }

  function inputValue(event: Event): string {
    return (event.currentTarget as HTMLInputElement).value;
  }
</script>

<div class="rounded-lg border bg-background p-4">
  <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
    <div class="min-w-0">
      <h3 class="flex items-center gap-2 text-sm font-semibold">
        <GitBranch class="size-4" />
        {$t('field_device.importer.title')}
      </h3>
      <p class="text-xs text-muted-foreground">
        {worksheet ? worksheet.name : $t('field_device.importer.no_worksheet')}
      </p>
    </div>

    <div class="flex flex-wrap gap-2">
      <Button
        type="button"
        variant="outline"
        disabled={!worksheet || service.isTransforming || service.isImporting}
        onclick={() => service.transform(worksheet)}
      >
        {#if service.isTransforming}
          <LoaderCircle class="mr-2 size-4 animate-spin" />
        {:else}
          <WandSparkles class="mr-2 size-4" />
        {/if}
        {$t('field_device.importer.actions.transform')}
      </Button>

      <Button
        type="button"
        disabled={!service.canImport || service.isTransforming}
        onclick={() => service.importPlan()}
      >
        {#if service.isImporting}
          <LoaderCircle class="mr-2 size-4 animate-spin" />
        {:else}
          <Play class="mr-2 size-4" />
        {/if}
        {service.importReport?.status === 'partial' || service.importReport?.status === 'failed'
          ? $t('field_device.importer.actions.retry')
          : $t('field_device.importer.actions.import')}
      </Button>
    </div>
  </div>

  {#if service.transformError}
    <div
      class="mt-4 rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive"
    >
      {service.transformError}
    </div>
  {/if}

  {#if service.plan}
    <div class="mt-4 grid grid-cols-2 gap-2 text-xs md:grid-cols-5">
      <div class="rounded-md border p-2">
        <div class="text-muted-foreground">
          {$t('field_device.importer.summary.control_cabinet')}
        </div>
        <div class="truncate font-medium">
          {service.plan.controller.controlCabinetNr || $t('common.not_available')}
        </div>
      </div>
      <div class="rounded-md border p-2">
        <div class="text-muted-foreground">{$t('field_device.importer.summary.system_types')}</div>
        <div class="font-medium">{service.plan.controller.systemTypes.length}</div>
      </div>
      <div class="rounded-md border p-2">
        <div class="text-muted-foreground">{$t('field_device.importer.summary.field_devices')}</div>
        <div class="font-medium">{service.plan.fieldDeviceCount}</div>
      </div>
      <div class="rounded-md border p-2">
        <div class="text-muted-foreground">
          {$t('field_device.importer.summary.bacnet_objects')}
        </div>
        <div class="font-medium">{service.plan.bacnetObjectCount}</div>
      </div>
      <div class="rounded-md border p-2">
        <div class="text-muted-foreground">{$t('field_device.importer.summary.validation')}</div>
        <div class="font-medium">
          {$t('field_device.importer.summary.validation_counts', {
            errors: blockingDiagnostics.length,
            warnings: warningDiagnostics.length
          })}
        </div>
      </div>
    </div>

    {#if service.importReport}
      <div class={`mt-4 rounded-md border p-3 text-sm ${reportClass(service.importReport.status)}`}>
        <div class="flex items-center gap-2 font-medium">
          {#if service.importReport.status === 'success'}
            <CheckCircle2 class="size-4" />
          {:else if service.importReport.status === 'partial'}
            <AlertTriangle class="size-4" />
          {:else}
            <XCircle class="size-4" />
          {/if}
          {reportMessage(service.importReport)}
        </div>
        {#if service.importReport.createdControlCabinetLabel || service.importReport.createdSpsControllerLabel}
          <p class="mt-1 text-xs">
            {$t('field_device.importer.report.created_entities', {
              controlCabinet:
                service.importReport.createdControlCabinetLabel ?? $t('common.not_available'),
              spsController:
                service.importReport.createdSpsControllerLabel ?? $t('common.not_available')
            })}
          </p>
        {/if}
      </div>
    {/if}

    {#if diagnostics.length > 0}
      <div class="mt-4 space-y-2">
        <h4 class="text-sm font-medium">{$t('field_device.importer.diagnostics.title')}</h4>
        <div class="max-h-52 space-y-2 overflow-auto pr-1">
          {#each diagnostics as diagnostic (diagnostic.id)}
            <div class={`rounded-md border p-2 text-xs ${diagnosticClass(diagnostic)}`}>
              <div class="font-medium">
                {diagnostic.cell ? `${diagnostic.cell.address}: ` : ''}{diagnostic.message}
              </div>
              <div class="mt-1 opacity-80">{diagnosticEntityLabel(diagnostic)}</div>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    {#if service.plan.controller.systemTypes.length > 0}
      <div class="mt-4 space-y-2">
        <h4 class="text-sm font-medium">{$t('field_device.importer.tree.title')}</h4>
        <details class={`rounded-md border p-3 ${nodeClass(CONTROL_CABINET_IMPORT_NODE_KEY)}`} open>
          <summary class="cursor-pointer text-sm font-medium">
            <span class="flex min-w-0 flex-wrap items-center gap-2">
              <span class="min-w-0 truncate">
                {$t('field_device.importer.tree.root', {
                  controlCabinet:
                    service.plan.controller.controlCabinetNr || $t('common.not_available'),
                  sps:
                    service.plan.controller.spsControllerRequest?.ga_device ??
                    $t('common.not_available')
                })}
              </span>
              <span
                class={`shrink-0 rounded border px-2 py-0.5 text-[11px] font-medium ${nodeBadgeClass(
                  CONTROL_CABINET_IMPORT_NODE_KEY
                )}`}
              >
                {nodeStatusLabel(CONTROL_CABINET_IMPORT_NODE_KEY)}
              </span>
              {#if nodeMessage(CONTROL_CABINET_IMPORT_NODE_KEY)}
                <span class="min-w-0 flex-1 text-xs text-destructive">
                  {nodeMessage(CONTROL_CABINET_IMPORT_NODE_KEY)}
                </span>
              {/if}
            </span>
          </summary>
          <div class="mt-3 space-y-3">
            <div
              class={`rounded-md border px-3 py-2 text-xs ${nodeClass(SPS_CONTROLLER_IMPORT_NODE_KEY)}`}
            >
              <div class="flex min-w-0 flex-wrap items-center gap-2">
                <span class="font-medium">
                  {service.plan.controller.spsControllerRequest?.device_name ??
                    $t('common.not_available')}
                </span>
                <span
                  class={`rounded border px-2 py-0.5 text-[11px] font-medium ${nodeBadgeClass(
                    SPS_CONTROLLER_IMPORT_NODE_KEY
                  )}`}
                >
                  {nodeStatusLabel(SPS_CONTROLLER_IMPORT_NODE_KEY)}
                </span>
                {#if nodeMessage(SPS_CONTROLLER_IMPORT_NODE_KEY)}
                  <span class="min-w-0 flex-1 text-destructive">
                    {nodeMessage(SPS_CONTROLLER_IMPORT_NODE_KEY)}
                  </span>
                {/if}
              </div>
            </div>
            {#each service.plan.controller.systemTypes as systemType (systemType.key)}
              <details class={`rounded-md border p-3 ${nodeClass(systemType.key)}`} open>
                <summary class="cursor-pointer text-sm">
                  <span class="flex min-w-0 flex-wrap items-center gap-2">
                    <span class="font-medium">
                      {systemType.number} · {systemType.systemTypeName}
                    </span>
                    <span class="text-muted-foreground">({systemType.fieldDeviceCount})</span>
                    <span
                      class={`rounded border px-2 py-0.5 text-[11px] font-medium ${nodeBadgeClass(
                        systemType.key
                      )}`}
                    >
                      {nodeStatusLabel(systemType.key)}
                    </span>
                    {#if nodeMessage(systemType.key)}
                      <span class="min-w-0 flex-1 text-xs text-destructive">
                        {nodeMessage(systemType.key)}
                      </span>
                    {/if}
                  </span>
                </summary>
                <div class="mt-2 space-y-2">
                  {#each devicesForSystemType(systemType.key) as device (device.key)}
                    <details class={`rounded-md border p-2 ${nodeClass(device.key)}`}>
                      <summary class="cursor-pointer text-xs font-medium">
                        <span class="flex min-w-0 flex-wrap items-center gap-2">
                          <span>
                            {device.systemPartLabel}{device.apparatLabel}{String(
                              device.apparatNr ?? ''
                            ).padStart(2, '0')}
                          </span>
                          <span class="text-muted-foreground">
                            {$t('field_device.importer.tree.device_meta', {
                              row: device.sourceRowNumber,
                              count: device.bacnetObjects.length
                            })}
                          </span>
                          <span
                            class={`rounded border px-2 py-0.5 text-[11px] font-medium ${nodeBadgeClass(
                              device.key
                            )}`}
                          >
                            {nodeStatusLabel(device.key)}
                          </span>
                          {#if nodeMessage(device.key)}
                            <span class="min-w-0 flex-1 text-destructive">
                              {nodeMessage(device.key)}
                            </span>
                          {/if}
                        </span>
                      </summary>
                      <div class="mt-2 grid gap-2 md:grid-cols-[120px_1fr_110px]">
                        <Input
                          value={device.request.bmk ?? ''}
                          aria-label={$t('field_device.importer.tree.fields.bmk')}
                          placeholder={$t('field_device.importer.tree.fields.bmk')}
                          class="h-8 text-xs"
                          oninput={(event) =>
                            service.updateFieldDeviceBmk(device.key, inputValue(event))}
                        />
                        <Input
                          value={device.request.description ?? ''}
                          aria-label={$t('field_device.importer.tree.fields.description')}
                          placeholder={$t('field_device.importer.tree.fields.description')}
                          class="h-8 text-xs"
                          oninput={(event) =>
                            service.updateFieldDeviceDescription(device.key, inputValue(event))}
                        />
                        <Input
                          type="number"
                          min="1"
                          max="99"
                          value={String(device.request.apparat_nr || '')}
                          aria-label={$t('field_device.importer.tree.fields.apparat_nr')}
                          placeholder={$t('field_device.importer.tree.fields.apparat_nr')}
                          class="h-8 text-xs"
                          oninput={(event) =>
                            service.updateFieldDeviceApparatNr(device.key, inputValue(event))}
                        />
                      </div>
                      <div class="mt-2 grid gap-1 text-xs text-muted-foreground">
                        {#each device.bacnetObjects as object (object.key)}
                          <div
                            class={`grid min-w-0 gap-2 rounded-md border px-2 py-1 md:grid-cols-[1fr_110px_1fr] ${nodeClass(
                              object.key
                            )}`}
                          >
                            <Input
                              value={object.textFix}
                              aria-label={$t('field_device.importer.tree.fields.text_fix')}
                              placeholder={$t('field_device.importer.tree.no_text_fix')}
                              class="h-8 text-xs"
                              oninput={(event) =>
                                service.updateBacnetTextFix(
                                  device.key,
                                  object.key,
                                  inputValue(event)
                                )}
                            />
                            <Input
                              value={object.address}
                              aria-label={$t('field_device.importer.tree.fields.address')}
                              placeholder={object.sourceCell.address}
                              class="h-8 font-mono text-xs"
                              oninput={(event) =>
                                service.updateBacnetAddress(
                                  device.key,
                                  object.key,
                                  inputValue(event)
                                )}
                            />
                            <div class="flex min-w-0 items-center gap-2">
                              <span
                                class={`shrink-0 rounded border px-2 py-0.5 text-[11px] font-medium ${nodeBadgeClass(
                                  object.key
                                )}`}
                              >
                                {nodeStatusLabel(object.key)}
                              </span>
                              {#if nodeMessage(object.key)}
                                <span class="min-w-0 text-destructive">
                                  {nodeMessage(object.key)}
                                </span>
                              {/if}
                            </div>
                          </div>
                        {/each}
                      </div>
                    </details>
                  {/each}
                </div>
              </details>
            {/each}
          </div>
        </details>
      </div>
    {/if}
  {:else}
    <div class="mt-4 rounded-md border border-dashed p-4 text-sm text-muted-foreground">
      {$t('field_device.importer.empty_prompt')}
    </div>
  {/if}
</div>
