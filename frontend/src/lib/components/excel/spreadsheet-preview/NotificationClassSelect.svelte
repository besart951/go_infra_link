<script lang="ts">
  import type { NotificationClass } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { FieldDeviceImportDevicePlan } from './fieldDeviceExportImporter.js';
  import { CREATE_NOTIFICATION_CLASS_VALUE } from './FieldDeviceImportService.svelte.js';
  import type { FieldDeviceImportViewState } from './fieldDeviceImportPresentation.js';
  import ImportStatusIcon from './ImportStatusIcon.svelte';

  type BacnetObjectPlan = FieldDeviceImportDevicePlan['bacnetObjects'][number];

  interface Props {
    object: BacnetObjectPlan;
    options: NotificationClass[];
    value: string;
    view: FieldDeviceImportViewState;
    onChange: (value: string) => void;
  }

  let { object, options, value, view, onChange }: Props = $props();
  const t = createTranslator();

  const message = $derived(view.notificationClassMessage(object));
  const visualKind = $derived(view.notificationClassVisualKind(object));
</script>

<div class="flex min-w-0 items-center gap-1">
  <select
    {value}
    aria-label={$t('field_device.importer.tree.fields.notification_class')}
    class="h-8 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-xs"
    onchange={(event) => onChange(view.inputValue(event))}
  >
    <option value="">
      {$t('field_device.importer.tree.notification.manual_assign')}
    </option>
    {#if object.notificationClass.number}
      <option value={CREATE_NOTIFICATION_CLASS_VALUE}>
        {$t('field_device.importer.tree.notification.create_new', {
          value: object.notificationClass.number
        })}
      </option>
    {/if}
    {#each options as notificationClass (notificationClass.id)}
      <option value={notificationClass.id}>
        {view.notificationClassOptionLabel(notificationClass)}
      </option>
    {/each}
  </select>
  <ImportStatusIcon kind={visualKind} messages={message} />
</div>
