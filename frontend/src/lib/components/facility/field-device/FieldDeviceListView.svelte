<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import ListPlusIcon from '@lucide/svelte/icons/list-plus';
  import { useUnsavedChangesWarning } from '$lib/hooks/useUnsavedChangesWarning.svelte.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import FieldDeviceMultiCreateForm from './FieldDeviceMultiCreateForm.svelte';
  import FieldDeviceFilterCard from './FieldDeviceFilterCard.svelte';
  import FieldDeviceSearchBar from './FieldDeviceSearchBar.svelte';
  import FieldDeviceBulkEditPanel from './FieldDeviceBulkEditPanel.svelte';
  import FieldDeviceTable from './FieldDeviceTable.svelte';
  import FieldDevicePagination from './FieldDevicePagination.svelte';
  import FieldDeviceFloatingSaveBar from './FieldDeviceFloatingSaveBar.svelte';
  import FieldDeviceExportPanel from './FieldDeviceExportPanel.svelte';
  import { provideFieldDeviceState } from './state/context.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';

  const t = createTranslator();

  interface Props {
    projectId?: string;
    refreshKey?: string | number;
    refreshRequest?: import('./state/types.js').FieldDeviceRefreshRequest;
    pageSize?: number;
    systemTypeRefreshKey?: string | number;
    sharedFieldDeviceEditors?: import('./state/types.js').SharedFieldDeviceEditorsByDevice;
    onSharedFieldDeviceStateChange?: (
      state: import('./state/types.js').SharedFieldDeviceDraftState
    ) => void;
    onFieldDevicesSaved?: (devices: import('$lib/domain/facility/index.js').FieldDevice[]) => void;
    onMultiCreateFormVisibilityChange?: (open: boolean) => void;
  }

  const {
    projectId,
    refreshKey,
    refreshRequest,
    pageSize = 300,
    systemTypeRefreshKey,
    sharedFieldDeviceEditors,
    onSharedFieldDeviceStateChange,
    onFieldDevicesSaved,
    onMultiCreateFormVisibilityChange
  }: Props = $props();

  const fieldDeviceState = provideFieldDeviceState({
    projectId: () => projectId,
    pageSize: () => pageSize,
    sharedFieldDeviceEditors: () => sharedFieldDeviceEditors ?? {},
    onSharedFieldDeviceStateChange: (state) => onSharedFieldDeviceStateChange?.(state),
    onFieldDevicesSaved: (devices) => onFieldDevicesSaved?.(devices)
  });

  useUnsavedChangesWarning(() => fieldDeviceState.editing.hasUnsavedChanges);

  let initialized = $state(false);
  let lastRefreshKey: string | number | undefined = $state(undefined);
  let lastRefreshRequestKey: string | number | undefined = $state(undefined);

  onMount(() => {
    initialized = true;
    lastRefreshKey = refreshKey;
    void fieldDeviceState.initialize();
  });

  onDestroy(() => {
    fieldDeviceState.dispose();
  });

  $effect(() => {
    const nextRefreshKey = refreshKey;

    if (!initialized) return;
    if (nextRefreshKey === undefined || nextRefreshKey === lastRefreshKey) {
      lastRefreshKey = nextRefreshKey;
      return;
    }

    lastRefreshKey = nextRefreshKey;
    void fieldDeviceState.reload();
  });

  $effect(() => {
    onMultiCreateFormVisibilityChange?.(fieldDeviceState.showMultiCreateForm);
  });

  $effect(() => {
    const nextRefreshRequest = refreshRequest;

    if (!initialized) return;
    if (!nextRefreshRequest || nextRefreshRequest.key === lastRefreshRequestKey) {
      lastRefreshRequestKey = nextRefreshRequest?.key;
      return;
    }

    lastRefreshRequestKey = nextRefreshRequest.key;
    if (nextRefreshRequest.devices && nextRefreshRequest.devices.length > 0) {
      void fieldDeviceState.applyDeviceDelta(nextRefreshRequest.devices);
      return;
    }

    if (nextRefreshRequest.spsControllers && nextRefreshRequest.spsControllers.length > 0) {
      fieldDeviceState.applySPSControllerDelta(nextRefreshRequest.spsControllers);
      return;
    }

    if (nextRefreshRequest.deviceIds && nextRefreshRequest.deviceIds.length > 0) {
      void fieldDeviceState.refreshDevices(nextRefreshRequest.deviceIds);
      return;
    }

    if (nextRefreshRequest.spsControllerIds && nextRefreshRequest.spsControllerIds.length > 0) {
      void fieldDeviceState.refreshDevicesForSPSControllers(nextRefreshRequest.spsControllerIds);
      return;
    }

    void fieldDeviceState.reload();
  });
</script>

<div class="flex min-w-0 flex-col gap-6">
  <div class="flex justify-end gap-2">
    {#if !fieldDeviceState.showMultiCreateForm && canPerform('create', 'fielddevice')}
      <Button onclick={() => fieldDeviceState.openMultiCreateForm()}>
        <ListPlusIcon class="size-4" />
        {$t('field_device.actions.multi_create')}
      </Button>
    {/if}
  </div>

  {#if fieldDeviceState.showMultiCreateForm}
    <Card.Root class="bg-background">
      <Card.Header>
        <Card.Title>{$t('field_device.multi_create.title')}</Card.Title>
        <Card.Description>
          {$t('field_device.multi_create.description')}
        </Card.Description>
      </Card.Header>
      <Card.Content>
        <FieldDeviceMultiCreateForm
          {projectId}
          {systemTypeRefreshKey}
          onSuccess={(createdDevices) => fieldDeviceState.handleMultiCreateSuccess(createdDevices)}
          onCancel={() => fieldDeviceState.closeMultiCreateForm()}
        />
      </Card.Content>
    </Card.Root>
  {/if}

  <div class="flex min-w-0 flex-col gap-4">
    <FieldDeviceSearchBar />

    <div class:hidden={!fieldDeviceState.showFilterPanel}>
      <FieldDeviceFilterCard showProjectFilter={!projectId} />
    </div>

    {#if fieldDeviceState.showExportPanel}
      <FieldDeviceExportPanel {projectId} />
    {/if}

    {#if fieldDeviceState.showBulkEditPanel}
      <FieldDeviceBulkEditPanel />
    {/if}

    {#if fieldDeviceState.error}
      <div
        class="rounded-md border border-destructive/50 bg-destructive/15 px-4 py-3 text-destructive"
      >
        <p class="font-medium">{$t('common.error')}</p>
        <p class="text-sm">{fieldDeviceState.error}</p>
      </div>
    {/if}

    <FieldDeviceTable />
    <FieldDevicePagination />
  </div>

  <FieldDeviceFloatingSaveBar />
</div>
