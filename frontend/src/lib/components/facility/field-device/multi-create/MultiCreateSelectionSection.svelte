<script lang="ts">
  import { Label } from '$lib/components/ui/label/index.js';
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import FieldDevicePreselection from '../FieldDevicePreselection.svelte';
  import ObjectDataBacnetPreview from './ObjectDataBacnetPreview.svelte';
  import * as Card from '$lib/components/ui/card/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  import type { ObjectData, SPSControllerSystemType } from '$lib/domain/facility/index.js';
  import type { MultiCreateSelection } from '$lib/domain/facility/fieldDeviceMultiCreate.js';
  import type { FieldDevicePreselection as PreselectionType } from '$lib/domain/facility/preselectionFilter.js';

  type Props = {
    projectId?: string;
    systemTypeRefreshKey?: string | number;
    selection: MultiCreateSelection;
    preselectionValue: PreselectionType;
    submitting: boolean;
    selectedObjectData: ObjectData | null;
    loadingObjectDataPreview: boolean;
    objectDataPreviewError?: string;
    onSpsSystemTypeChange: (value: string) => void;
    onPreselectionChange: (value: PreselectionType) => void;
    fetchSpsControllerSystemTypes: (search: string) => Promise<SPSControllerSystemType[]>;
    fetchSpsControllerSystemTypeById: (id: string) => Promise<SPSControllerSystemType | null>;
    formatSpsControllerSystemTypeLabel: (item: SPSControllerSystemType) => string;
  };

  let {
    projectId,
    systemTypeRefreshKey,
    selection,
    preselectionValue,
    submitting,
    selectedObjectData,
    loadingObjectDataPreview,
    objectDataPreviewError,
    onSpsSystemTypeChange,
    onPreselectionChange,
    fetchSpsControllerSystemTypes,
    fetchSpsControllerSystemTypeById,
    formatSpsControllerSystemTypeLabel
  }: Props = $props();

  const t = createTranslator();
</script>

<Card.Root class="p-6">
  <div class="grid gap-4 md:grid-cols-2">
    <div class="space-y-2">
      <Label for="sps-system-type"
        >{$t('field_device.multi_create.selection.sps_system_type')}</Label
      >
      <AsyncCombobox
        id="sps-system-type"
        refreshKey={systemTypeRefreshKey}
        placeholder={$t('field_device.multi_create.selection.sps_system_type_placeholder')}
        searchPlaceholder={$t('field_device.multi_create.selection.sps_system_type_search')}
        emptyText={$t('field_device.multi_create.selection.sps_system_type_empty')}
        fetcher={fetchSpsControllerSystemTypes}
        fetchById={fetchSpsControllerSystemTypeById}
        labelKey="system_type_name"
        labelFormatter={formatSpsControllerSystemTypeLabel}
        width="w-full"
        value={selection.spsControllerSystemTypeId}
        onValueChange={onSpsSystemTypeChange}
        clearable
        clearText={$t('field_device.multi_create.selection.sps_system_type_clear')}
        disabled={submitting}
      />
    </div>
  </div>

  <div class="mt-4">
    <FieldDevicePreselection
      value={preselectionValue}
      onChange={onPreselectionChange}
      {projectId}
      disabled={!selection.spsControllerSystemTypeId || submitting}
      className="grid grid-cols-1 gap-4 md:grid-cols-3"
    />
  </div>

  <ObjectDataBacnetPreview
    objectData={selectedObjectData}
    loading={loadingObjectDataPreview}
    error={objectDataPreviewError}
  />
</Card.Root>
