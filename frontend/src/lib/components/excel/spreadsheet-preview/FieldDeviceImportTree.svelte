<script lang="ts">
  import { Input } from '$lib/components/ui/input/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import {
    CONTROL_CABINET_IMPORT_NODE_KEY,
    SPS_CONTROLLER_IMPORT_NODE_KEY,
    type FieldDeviceImportService
  } from './FieldDeviceImportService.svelte.js';
  import type { FieldDeviceImportViewState } from './fieldDeviceImportPresentation.js';
  import BacnetObjectImportRow from './BacnetObjectImportRow.svelte';
  import ImportStatusBadge from './ImportStatusBadge.svelte';
  import ImportStatusIcon from './ImportStatusIcon.svelte';

  interface Props {
    service: FieldDeviceImportService;
    view: FieldDeviceImportViewState;
  }

  let { service, view }: Props = $props();
  const t = createTranslator();
</script>

{#if service.plan && service.visibleSystemTypes.length > 0}
  {@const rootNode = view.node(CONTROL_CABINET_IMPORT_NODE_KEY)}
  {@const spsNode = view.node(SPS_CONTROLLER_IMPORT_NODE_KEY)}
  <div class="mt-4 space-y-2">
    <h4 class="text-sm font-medium">{$t('field_device.importer.tree.title')}</h4>
    <details class={`rounded-md border p-3 ${rootNode.className}`} open>
      <summary class="cursor-pointer text-sm font-medium">
        <span class="flex min-w-0 flex-wrap items-center gap-2">
          <ImportStatusIcon kind={rootNode.visualKind} messages={rootNode.message} />
          <span class="min-w-0 truncate">
            {$t('field_device.importer.tree.root', {
              controlCabinet:
                service.plan.controller.controlCabinetNr || $t('common.not_available'),
              sps:
                service.plan.controller.spsControllerRequest?.ga_device ??
                $t('common.not_available')
            })}
          </span>
          <ImportStatusBadge node={rootNode} />
        </span>
      </summary>
      <div class="mt-3 space-y-3">
        <div class={`rounded-md border px-3 py-2 text-xs ${spsNode.className}`}>
          <div class="flex min-w-0 flex-wrap items-center gap-2">
            <ImportStatusIcon kind={spsNode.visualKind} messages={spsNode.message} />
            <span class="font-medium">
              {service.plan.controller.spsControllerRequest?.device_name ??
                $t('common.not_available')}
            </span>
            <ImportStatusBadge node={spsNode} />
          </div>
        </div>

        {#each service.visibleSystemTypes as systemType (systemType.key)}
          {@const systemTypeNode = view.node(systemType.key)}
          <details class={`rounded-md border p-3 ${systemTypeNode.className}`} open>
            <summary class="cursor-pointer text-sm">
              <span class="flex min-w-0 flex-wrap items-center gap-2">
                <ImportStatusIcon
                  kind={systemTypeNode.visualKind}
                  messages={systemTypeNode.message}
                />
                <span class="font-medium">
                  {systemType.number} · {systemType.systemTypeName}
                </span>
                <span class="text-muted-foreground">({systemType.fieldDeviceCount})</span>
                <ImportStatusBadge node={systemTypeNode} />
              </span>
            </summary>
            <div class="mt-2 space-y-2">
              {#each view.devicesForSystemType(systemType.key) as device (device.key)}
                {@const deviceNode = view.node(device.key)}
                <details class={`rounded-md border p-2 ${deviceNode.className}`}>
                  <summary class="cursor-pointer text-xs font-medium">
                    <span class="flex min-w-0 flex-wrap items-center gap-2">
                      <ImportStatusIcon
                        kind={deviceNode.visualKind}
                        messages={deviceNode.message}
                      />
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
                      <ImportStatusBadge node={deviceNode} />
                    </span>
                  </summary>
                  <div class="mt-2 grid gap-2 md:grid-cols-[120px_1fr_110px]">
                    <Input
                      value={device.request.bmk ?? ''}
                      aria-label={$t('field_device.importer.tree.fields.bmk')}
                      placeholder={$t('field_device.importer.tree.fields.bmk')}
                      class="h-8 text-xs"
                      oninput={(event) =>
                        service.updateFieldDeviceBmk(device.key, view.inputValue(event))}
                    />
                    <Input
                      value={device.request.description ?? ''}
                      aria-label={$t('field_device.importer.tree.fields.description')}
                      placeholder={$t('field_device.importer.tree.fields.description')}
                      class="h-8 text-xs"
                      oninput={(event) =>
                        service.updateFieldDeviceDescription(device.key, view.inputValue(event))}
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
                        service.updateFieldDeviceApparatNr(device.key, view.inputValue(event))}
                    />
                  </div>
                  <div class="mt-2 grid gap-1 text-xs text-muted-foreground">
                    {#each device.bacnetObjects as object (object.key)}
                      <BacnetObjectImportRow deviceKey={device.key} {object} {service} {view} />
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
