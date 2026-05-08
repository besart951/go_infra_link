<script lang="ts">
  import { Input } from '$lib/components/ui/input/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { FieldDeviceImportService } from './FieldDeviceImportService.svelte.js';
  import type { FieldDeviceImportDevicePlan } from './fieldDeviceExportImporter.js';
  import type { FieldDeviceImportViewState } from './fieldDeviceImportPresentation.js';
  import ImportStatusBadge from './ImportStatusBadge.svelte';
  import ImportStatusIcon from './ImportStatusIcon.svelte';
  import NotificationClassSelect from './NotificationClassSelect.svelte';

  type BacnetObjectPlan = FieldDeviceImportDevicePlan['bacnetObjects'][number];

  interface Props {
    deviceKey: string;
    object: BacnetObjectPlan;
    service: FieldDeviceImportService;
    view: FieldDeviceImportViewState;
  }

  let { deviceKey, object, service, view }: Props = $props();
  const t = createTranslator();

  const node = $derived(view.node(object.key));
  const notificationClassMessage = $derived(view.notificationClassMessage(object));
</script>

<div
  class={`grid min-w-0 gap-2 rounded-md border px-2 py-1 md:grid-cols-[minmax(160px,1fr)_110px_220px_minmax(160px,1fr)] ${node.className}`}
>
  <Input
    value={object.textFix}
    aria-label={$t('field_device.importer.tree.fields.text_fix')}
    placeholder={$t('field_device.importer.tree.no_text_fix')}
    class="h-8 text-xs"
    oninput={(event) => service.updateBacnetTextFix(deviceKey, object.key, view.inputValue(event))}
  />
  <Input
    value={object.address}
    aria-label={$t('field_device.importer.tree.fields.address')}
    placeholder={object.sourceCell.address}
    class="h-8 font-mono text-xs"
    oninput={(event) => service.updateBacnetAddress(deviceKey, object.key, view.inputValue(event))}
  />
  <NotificationClassSelect
    {object}
    options={service.notificationClassOptions}
    value={service.notificationClassSelectionValue(object)}
    {view}
    onChange={(value) => service.updateBacnetNotificationClass(deviceKey, object.key, value)}
  />
  <div class="flex min-w-0 items-center gap-2">
    <ImportStatusIcon
      kind={node.visualKind}
      messages={notificationClassMessage ? [] : node.message}
    />
    <ImportStatusBadge {node} />
  </div>
</div>
