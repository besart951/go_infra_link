<script lang="ts">
  import { Label } from '$lib/components/ui/label/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import FieldDevicePreselection from '../FieldDevicePreselection.svelte';
  import ObjectDataBacnetPreview from './ObjectDataBacnetPreview.svelte';
  import * as Card from '$lib/components/ui/card/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { FieldDeviceMultiCreateState } from './FieldDeviceMultiCreateState.svelte.js';

  type Props = {
    state: FieldDeviceMultiCreateState;
    systemTypeRefreshKey?: string | number;
  };

  let { state, systemTypeRefreshKey }: Props = $props();

  const t = createTranslator();

  const comboboxRefreshKey = $derived(`${systemTypeRefreshKey ?? ''}-${state.projectOnly}`);

  function handleProjectOnlyCheckedChange(checked: boolean | 'indeterminate') {
    if (typeof checked !== 'boolean') {
      return;
    }

    state.handleProjectOnlyChange(checked);
  }
</script>

<Card.Root class="p-6">
  <div class="grid gap-4 md:grid-cols-2">
    <div class="space-y-2">
      <Label for="sps-system-type"
        >{$t('field_device.multi_create.selection.sps_system_type')}</Label
      >
      <div class="flex items-center gap-2">
        {#if state.projectId}
          <Checkbox
            id="project-only-filter"
            checked={state.projectOnly}
            onCheckedChange={handleProjectOnlyCheckedChange}
            disabled={state.submitting}
          />
          <Label
            for="project-only-filter"
            class="shrink-0 cursor-pointer text-xs text-muted-foreground"
          >
            {$t('field_device.multi_create.selection.project_only')}
          </Label>
        {/if}
      </div>
      <AsyncCombobox
        id="sps-system-type"
        refreshKey={comboboxRefreshKey}
        placeholder={$t('field_device.multi_create.selection.sps_system_type_placeholder')}
        searchPlaceholder={$t('field_device.multi_create.selection.sps_system_type_search')}
        emptyText={$t('field_device.multi_create.selection.sps_system_type_empty')}
        fetcher={(search) => state.fetchSpsControllerSystemTypes(search)}
        fetchById={(id) => state.fetchSpsControllerSystemTypeById(id)}
        labelKey="system_type_name"
        labelFormatter={(item) => state.formatSpsControllerSystemTypeLabel(item)}
        width="w-full"
        value={state.selection.spsControllerSystemTypeId}
        onValueChange={(v) => state.handleSpsSystemTypeChange(v)}
        clearable
        popupWidth="w-[360px]"
        clearText={$t('field_device.multi_create.selection.sps_system_type_clear')}
        disabled={state.submitting}
      />
    </div>
  </div>

  <div class="mt-4">
    <FieldDevicePreselection
      value={state.preselectionValue}
      onChange={(v) => state.handlePreselectionChange(v)}
      projectId={state.projectId}
      disabled={!state.selection.spsControllerSystemTypeId || state.submitting}
      className="grid grid-cols-1 gap-4 md:grid-cols-3"
    />
  </div>

  <ObjectDataBacnetPreview
    objectData={state.selectedObjectData}
    loading={state.loadingObjectDataPreview}
    error={state.objectDataPreviewError}
  />
</Card.Root>
