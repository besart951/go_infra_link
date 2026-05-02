<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import ControlCabinetSelect from '../selects/ControlCabinetSelect.svelte';
  import SystemTypeSelect from '../selects/SystemTypeSelect.svelte';
  import { spsControllerFormService } from './state/SPSControllerFormService.js';
  import {
    buildSPSControllerDeviceName,
    buildSPSControllerSystemTypeLabel,
    collectSystemTypeLabelFallbacks,
    collectUniqueSystemTypeIds,
    getNextAvailableSystemTypeNumber,
    getSPSControllerSystemTypeAddState,
    toSPSControllerSystemTypeEntries,
    toSPSControllerSystemTypeInput,
    type SPSControllerSystemTypeEntry,
    updateSPSControllerSystemTypeEntry
  } from './state/SPSControllerFormDraft.js';
  import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
  import type {
    Building,
    ControlCabinet,
    SPSController,
    SPSControllerSystemTypeInput,
    SystemType
  } from '$lib/domain/facility/index.js';
  import { useLiveValidation } from '$lib/hooks/useLiveValidation.svelte.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { t as translate } from '$lib/i18n/index.js';
  import CopyIcon from '@lucide/svelte/icons/copy';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';

  interface Props {
    initialData?: SPSController;
    projectId?: string;
    fixedControlCabinetId?: string;
    controlCabinetRefreshKey?: string | number;
    onSuccess?: (controller: SPSController) => void;
    onCancel?: () => void;
  }

  let {
    initialData,
    projectId,
    fixedControlCabinetId,
    controlCabinetRefreshKey,
    onSuccess,
    onCancel
  }: Props = $props();

  const t = createTranslator();

  let ga_device = $state('');
  let device_name = $state('');
  let ip_address = $state('');
  let subnet = $state('');
  let gateway = $state('');
  let vlan = $state('');
  let control_cabinet_id = $state('');
  let system_type_id = $state('');
  let systemTypes: SPSControllerSystemTypeEntry[] = $state([]);
  let systemTypeLabels: Record<string, string> = $state({});
  let systemTypeDetails: Record<string, SystemType> = $state({});
  let systemTypeDetailsLoading = $state(false);
  let systemTypesLoading = $state(false);
  let lastLoadedSystemTypesFor: string | null = $state(null);
  let controlCabinet: ControlCabinet | null = $state(null);
  let building: Building | null = $state(null);
  let cabinetLoading = $state(false);
  let buildingLoading = $state(false);
  let lastLoadedCabinetId: string | null = $state(null);
  let gaDeviceTouched = $state(false);
  let lastGADeviceControlCabinetId: string | null = $state(null);
  let nextGADevice = $state<string | null>(null);
  let gaDeviceSuggestionLoading = $state(false);
  let gaDeviceCheckTimer: ReturnType<typeof setTimeout> | null = null;

  let loading = $state(false);
  let error = $state('');
  let fieldErrors = $state<Record<string, string>>({});
  const liveValidation = useLiveValidation(
    (data: {
      id?: string;
      control_cabinet_id: string;
      ga_device?: string;
      device_name: string;
      ip_address?: string;
      subnet?: string;
      gateway?: string;
      vlan?: string;
    }) => spsControllerFormService.validate(data),
    { debounceMs: 400 }
  );

  $effect(() => {
    if (!initialData) {
      return;
    }
    ga_device = initialData.ga_device ?? '';
    device_name = initialData.device_name;
    ip_address = initialData.ip_address ?? '';
    subnet = initialData.subnet ?? '';
    gateway = initialData.gateway ?? '';
    vlan = initialData.vlan ?? '';
    control_cabinet_id = initialData.control_cabinet_id;
    gaDeviceTouched = false;
    nextGADevice = null;
  });

  $effect(() => {
    if (!fixedControlCabinetId) return;
    if (control_cabinet_id === fixedControlCabinetId) return;
    control_cabinet_id = fixedControlCabinetId;
  });

  $effect(() => {
    if (initialData?.id && lastLoadedSystemTypesFor !== initialData.id) {
      lastLoadedSystemTypesFor = initialData.id;
      loadSystemTypes();
    }
    if (!initialData?.id && lastLoadedSystemTypesFor !== null) {
      lastLoadedSystemTypesFor = null;
      systemTypes = [];
      systemTypeLabels = {};
    }
  });

  $effect(() => {
    if (!control_cabinet_id) {
      nextGADevice = null;
      controlCabinet = null;
      building = null;
      device_name = '';
      if (gaDeviceCheckTimer) {
        clearTimeout(gaDeviceCheckTimer);
        gaDeviceCheckTimer = null;
      }
      return;
    }
    triggerValidation();
    if (lastGADeviceControlCabinetId !== control_cabinet_id) {
      lastGADeviceControlCabinetId = control_cabinet_id;
      if (!initialData) {
        ga_device = '';
        gaDeviceTouched = false;
        void refreshNextGADevice(true);
      } else {
        void refreshNextGADevice(false);
      }
    } else if (!nextGADevice && !gaDeviceSuggestionLoading) {
      void refreshNextGADevice(false);
    }
  });

  $effect(() => {
    if (!control_cabinet_id) return;
    if (lastLoadedCabinetId === control_cabinet_id) return;
    lastLoadedCabinetId = control_cabinet_id;
    void loadCabinetAndBuilding(control_cabinet_id);
  });

  $effect(() => {
    const nextName = buildSPSControllerDeviceName(controlCabinet, building, ga_device);
    if (nextName === null) return;
    if (nextName !== device_name) {
      device_name = nextName;
      triggerValidation();
    }
  });

  const fieldError = (name: string) => getFieldError(fieldErrors, name, ['spscontroller']);
  const liveFieldError = (name: string) =>
    getFieldError(liveValidation.fieldErrors, name, ['spscontroller']);
  const combinedFieldError = (name: string) => liveFieldError(name) || fieldError(name);
  const gaDeviceIsSuboptimal = $derived.by(() => {
    if (!gaDeviceTouched || !nextGADevice) return false;
    const current = ga_device.trim().toUpperCase();
    return current !== '' && current !== nextGADevice;
  });

  function triggerValidation() {
    if (!control_cabinet_id) return;
    liveValidation.trigger({
      id: initialData?.id,
      control_cabinet_id,
      ga_device: ga_device || undefined,
      device_name,
      ip_address: ip_address || undefined,
      subnet: subnet || undefined,
      gateway: gateway || undefined,
      vlan: vlan || undefined
    });
  }

  function scheduleNextGADeviceCheck() {
    if (!control_cabinet_id) return;
    if (gaDeviceCheckTimer) {
      clearTimeout(gaDeviceCheckTimer);
    }
    gaDeviceCheckTimer = setTimeout(() => {
      gaDeviceCheckTimer = null;
      void refreshNextGADevice(false);
    }, 400);
  }

  async function fetchNextGADevice(): Promise<string | null> {
    if (!control_cabinet_id) return null;
    gaDeviceSuggestionLoading = true;
    try {
      const res = await spsControllerFormService.getNextGADevice(
        control_cabinet_id,
        initialData?.id
      );
      nextGADevice = res?.ga_device ?? null;
      return nextGADevice;
    } catch (e) {
      console.error(e);
      return null;
    } finally {
      gaDeviceSuggestionLoading = false;
    }
  }

  async function refreshNextGADevice(applyIfEmpty: boolean) {
    const next = await fetchNextGADevice();
    if (applyIfEmpty && next) {
      ga_device = next;
      gaDeviceTouched = false;
      triggerValidation();
    }
  }

  async function applySuggestedGADevice() {
    if (!control_cabinet_id) return;
    if (!nextGADevice) {
      await refreshNextGADevice(false);
    }
    if (nextGADevice) {
      ga_device = nextGADevice;
      gaDeviceTouched = false;
      triggerValidation();
    }
  }

  async function loadSystemTypes() {
    if (!initialData?.id) return;
    systemTypesLoading = true;
    try {
      const res = await spsControllerFormService.listSystemTypes({
        pagination: { page: 1, pageSize: 100 },
        search: { text: '' },
        filters: { sps_controller_id: initialData.id }
      });
      const items = res.items;
      const labelFallbacks = collectSystemTypeLabelFallbacks(items);
      const uniqueIds = collectUniqueSystemTypeIds(items);
      systemTypes = toSPSControllerSystemTypeEntries(items);
      systemTypeLabels = await buildSystemTypeLabels(uniqueIds, labelFallbacks);
      await Promise.all(uniqueIds.map((id) => ensureSystemTypeDetails(id)));
    } catch (e) {
      console.error(e);
    } finally {
      systemTypesLoading = false;
    }
  }

  $effect(() => {
    if (!system_type_id) return;
    if (systemTypeDetails[system_type_id]) return;
    void ensureSystemTypeDetails(system_type_id);
  });

  async function addSystemType() {
    if (!system_type_id) {
      return;
    }
    try {
      const systemType = await ensureSystemTypeDetails(system_type_id);
      if (!systemType) return;
      systemTypeLabels = {
        ...systemTypeLabels,
        [system_type_id]: buildSPSControllerSystemTypeLabel(
          systemType.name,
          systemType.number_min,
          systemType.number_max
        )
      };
      const nextNumber = getNextAvailableSystemTypeNumber(
        systemTypes,
        system_type_id,
        systemType.number_min,
        systemType.number_max
      );
      if (nextNumber == null) {
        return;
      }
      systemTypes = [
        ...systemTypes,
        {
          system_type_id,
          number: nextNumber
        }
      ];
    } catch (e) {
      console.error(e);
    }
  }

  async function ensureSystemTypeDetails(id: string): Promise<SystemType | null> {
    if (systemTypeDetails[id]) return systemTypeDetails[id];
    systemTypeDetailsLoading = true;
    try {
      const systemType = await spsControllerFormService.getSystemType(id);
      systemTypeDetails = { ...systemTypeDetails, [id]: systemType };
      return systemType;
    } catch (e) {
      console.error(e);
      return null;
    } finally {
      systemTypeDetailsLoading = false;
    }
  }

  async function loadCabinetAndBuilding(cabinetId: string) {
    cabinetLoading = true;
    buildingLoading = true;
    try {
      const cabinet = await spsControllerFormService.getControlCabinet(cabinetId);
      if (control_cabinet_id !== cabinetId) return;
      controlCabinet = cabinet;
      if (!cabinet?.building_id) {
        building = null;
        return;
      }
      const b = await spsControllerFormService.getBuilding(cabinet.building_id);
      if (control_cabinet_id !== cabinetId) return;
      building = b;
    } catch (e) {
      console.error(e);
      controlCabinet = null;
      building = null;
    } finally {
      cabinetLoading = false;
      buildingLoading = false;
    }
  }

  async function buildSystemTypeLabels(
    ids: string[],
    fallbacks: Record<string, string>
  ): Promise<Record<string, string>> {
    const entries = await Promise.all(
      ids.map(async (id) => {
        try {
          const systemType = await spsControllerFormService.getSystemType(id);
          return [
            id,
            buildSPSControllerSystemTypeLabel(
              systemType.name,
              systemType.number_min,
              systemType.number_max
            )
          ] as const;
        } catch (error) {
          console.error('Failed to load system type details:', error);
          return [id, fallbacks[id] ?? id] as const;
        }
      })
    );
    return Object.fromEntries(entries);
  }

  async function removeSystemType(index: number) {
    const entry = systemTypes[index];
    if (!entry) return;
    if (!entry.id) {
      systemTypes = systemTypes.filter((_, i) => i !== index);
      return;
    }
    try {
      await spsControllerFormService.deleteSystemType(entry.id);
      systemTypes = systemTypes.filter((_, i) => i !== index);
    } catch (e) {
      console.error(e);
      error = getErrorMessage(e);
    }
  }

  function updateSystemTypeField(
    index: number,
    field: keyof SPSControllerSystemTypeInput,
    value: string
  ) {
    systemTypes = systemTypes.map((item, i) =>
      i === index ? updateSPSControllerSystemTypeEntry(item, field, value) : item
    );
  }

  async function copySystemType(index: number) {
    const entry = systemTypes[index];
    if (!entry?.id) return;
    try {
      const copied = projectId
        ? await spsControllerFormService.copyProjectSystemType(projectId, entry.id)
        : await spsControllerFormService.copySystemType(entry.id);
      const systemTypeId = copied.system_type_id;
      if (!systemTypeDetails[systemTypeId]) {
        const details = await ensureSystemTypeDetails(systemTypeId);
        if (details) {
          systemTypeLabels = {
            ...systemTypeLabels,
            [systemTypeId]: buildSPSControllerSystemTypeLabel(
              details.name,
              details.number_min,
              details.number_max
            )
          };
        }
      }
      systemTypes = [
        ...systemTypes,
        {
          id: copied.id,
          system_type_id: systemTypeId,
          number: copied.number ?? undefined,
          document_name: copied.document_name ?? undefined
        }
      ];
    } catch (e) {
      console.error(e);
      error = getErrorMessage(e);
    }
  }

  const systemTypeAddState = $derived.by(() => {
    return getSPSControllerSystemTypeAddState(
      system_type_id,
      systemTypeDetails,
      systemTypes,
      systemTypeDetailsLoading,
      translate
    );
  });

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    loading = true;
    error = '';
    fieldErrors = {};

    if (!control_cabinet_id) {
      error = translate('facility.forms.sps_controller.control_cabinet_required');
      loading = false;
      return;
    }

    try {
      if (initialData) {
        const res = await spsControllerFormService.update(initialData.id, {
          id: initialData.id,
          ga_device,
          device_name,
          ip_address: ip_address || undefined,
          subnet: subnet || undefined,
          gateway: gateway || undefined,
          vlan: vlan || undefined,
          control_cabinet_id,
          system_types: toSPSControllerSystemTypeInput(systemTypes)
        });
        onSuccess?.(res);
      } else {
        const res = await spsControllerFormService.create({
          ga_device,
          device_name,
          ip_address: ip_address || undefined,
          subnet: subnet || undefined,
          gateway: gateway || undefined,
          vlan: vlan || undefined,
          control_cabinet_id,
          system_types: toSPSControllerSystemTypeInput(systemTypes)
        });
        onSuccess?.(res);
      }
    } catch (e) {
      console.error(e);
      fieldErrors = getFieldErrors(e);
      error = Object.keys(fieldErrors).length ? '' : getErrorMessage(e);
    } finally {
      loading = false;
    }
  }
</script>

<form onsubmit={handleSubmit} class="space-y-4 rounded-md border bg-muted/20 p-4">
  <div class="mb-4 flex items-center justify-between">
    <h3 class="text-lg font-medium">
      {initialData
        ? $t('facility.forms.sps_controller.title_edit')
        : $t('facility.forms.sps_controller.title_new')}
    </h3>
  </div>

  <div class="space-y-2">
    <Label>{$t('facility.forms.sps_controller.control_cabinet_label')}</Label>
    {#if fixedControlCabinetId}
      <Input value={fixedControlCabinetId} readonly disabled />
    {:else}
      <div class="block">
        <ControlCabinetSelect
          bind:value={control_cabinet_id}
          width="w-full"
          refreshKey={controlCabinetRefreshKey}
        />
      </div>
    {/if}
    {#if combinedFieldError('control_cabinet_id')}
      <p class="text-sm text-destructive">{combinedFieldError('control_cabinet_id')}</p>
    {:else if !control_cabinet_id}
      <p class="text-sm text-muted-foreground">
        {$t('facility.forms.sps_controller.control_cabinet_help')}
      </p>
    {/if}
  </div>

  {#if control_cabinet_id}
    <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
      <div class="space-y-2">
        <Label for="ga_device" class="flex items-center gap-2">
          {$t('facility.forms.sps_controller.ga_device_label')}
          <span
            class="inline-flex h-4 w-4 items-center justify-center rounded-full border text-[10px] text-muted-foreground"
            title={$t('facility.forms.sps_controller.ga_device_tooltip')}
          >
            i
          </span>
        </Label>
        <div class="flex items-start gap-2">
          <Input
            id="ga_device"
            value={ga_device}
            required
            maxlength={3}
            class={`flex-1 ${
              gaDeviceIsSuboptimal
                ? 'border-warning-border bg-warning-muted/60 focus-visible:ring-warning-border'
                : ''
            }`}
            oninput={(e) => {
              ga_device = (e.target as HTMLInputElement).value.toUpperCase();
              gaDeviceTouched = true;
              triggerValidation();
              scheduleNextGADeviceCheck();
            }}
          />
          {#if gaDeviceIsSuboptimal}
            <Button
              type="button"
              variant="outline"
              size="sm"
              class="mt-1"
              onclick={applySuggestedGADevice}
            >
              {$t('facility.forms.sps_controller.ga_device_use_next')}
            </Button>
          {/if}
        </div>
        {#if gaDeviceIsSuboptimal && nextGADevice}
          <p class="text-xs text-warning-muted-foreground">
            {$t('facility.forms.sps_controller.ga_device_lowest', { value: nextGADevice })}
          </p>
        {/if}
        {#if gaDeviceSuggestionLoading}
          <p class="text-xs text-muted-foreground">
            {$t('facility.forms.sps_controller.ga_device_checking')}
          </p>
        {/if}
        {#if combinedFieldError('ga_device')}
          <p class="text-sm text-destructive">{combinedFieldError('ga_device')}</p>
        {/if}
      </div>
      <div class="space-y-2">
        <Label for="device_name">{$t('facility.forms.sps_controller.device_name_label')}</Label>
        <Input
          id="device_name"
          bind:value={device_name}
          disabled
          readonly
          required
          maxlength={100}
        />
        {#if cabinetLoading || buildingLoading}
          <p class="text-xs text-muted-foreground">
            {$t('facility.forms.sps_controller.device_name_loading')}
          </p>
        {/if}
        {#if combinedFieldError('device_name')}
          <p class="text-sm text-destructive">{combinedFieldError('device_name')}</p>
        {/if}
      </div>
      <div class="space-y-2">
        <Label for="ip_address">{$t('facility.forms.sps_controller.ip_address_label')}</Label>
        <Input id="ip_address" bind:value={ip_address} maxlength={50} oninput={triggerValidation} />
        {#if combinedFieldError('ip_address')}
          <p class="text-sm text-destructive">{combinedFieldError('ip_address')}</p>
        {/if}
      </div>
      <div class="space-y-2">
        <Label for="subnet">{$t('facility.forms.sps_controller.subnet_label')}</Label>
        <Input id="subnet" bind:value={subnet} maxlength={50} oninput={triggerValidation} />
        {#if combinedFieldError('subnet')}
          <p class="text-sm text-destructive">{combinedFieldError('subnet')}</p>
        {/if}
      </div>
      <div class="space-y-2">
        <Label for="gateway">{$t('facility.forms.sps_controller.gateway_label')}</Label>
        <Input id="gateway" bind:value={gateway} maxlength={50} oninput={triggerValidation} />
        {#if combinedFieldError('gateway')}
          <p class="text-sm text-destructive">{combinedFieldError('gateway')}</p>
        {/if}
      </div>
      <div class="space-y-2">
        <Label for="vlan">{$t('facility.forms.sps_controller.vlan_label')}</Label>
        <Input id="vlan" bind:value={vlan} maxlength={50} oninput={triggerValidation} />
        {#if combinedFieldError('vlan')}
          <p class="text-sm text-destructive">{combinedFieldError('vlan')}</p>
        {/if}
      </div>
    </div>

    <div class="space-y-3 pt-4">
      <div class="flex items-center justify-between border-t pt-4">
        <div>
          <h4 class="text-base font-medium">
            {$t('facility.forms.sps_controller.system_types_title')}
          </h4>
          <p class="text-sm text-muted-foreground">
            {$t('facility.forms.sps_controller.system_types_description')}
          </p>
        </div>
        <div class="flex items-center gap-2">
          <SystemTypeSelect bind:value={system_type_id} width="w-[250px]" />
          <Button
            type="button"
            variant="outline"
            size="sm"
            onclick={addSystemType}
            disabled={systemTypeAddState.disabled}
            title={systemTypeAddState.tooltip}
          >
            {$t('common.add')}
          </Button>
        </div>
      </div>

      {#if systemTypesLoading}
        <p class="text-sm text-muted-foreground">
          {$t('facility.forms.sps_controller.system_types_loading')}
        </p>
      {:else if systemTypes.length === 0}
        <div class="rounded-md border border-dashed p-6 text-center">
          <p class="text-sm text-muted-foreground">
            {$t('facility.forms.sps_controller.system_types_empty')}
          </p>
        </div>
      {:else}
        <div class="max-h-80 space-y-2 overflow-y-auto pr-1">
          {#each systemTypes as st, index (index)}
            <div class="grid grid-cols-1 gap-3 rounded-md border p-3 md:grid-cols-12">
              <div class="md:col-span-4">
                <div class="text-xs text-muted-foreground">
                  {$t('facility.forms.sps_controller.system_type_label')}
                </div>
                <div class="text-sm font-medium">
                  {systemTypeLabels[st.system_type_id] ?? st.system_type_id}
                </div>
              </div>
              <div class="md:col-span-3">
                <Label class="text-xs"
                  >{$t('facility.forms.sps_controller.system_type_number')}</Label
                >
                <Input
                  type="number"
                  value={st.number ?? ''}
                  oninput={(e) =>
                    updateSystemTypeField(index, 'number', (e.target as HTMLInputElement).value)}
                />
              </div>
              <div class="md:col-span-5">
                <Label class="text-xs"
                  >{$t('facility.forms.sps_controller.system_type_document')}</Label
                >
                <div class="relative">
                  <Input
                    value={st.document_name ?? ''}
                    oninput={(e) =>
                      updateSystemTypeField(
                        index,
                        'document_name',
                        (e.target as HTMLInputElement).value
                      )}
                    maxlength={250}
                    class="pr-16"
                  />
                  <div
                    class="pointer-events-none absolute top-1/2 right-2 flex -translate-y-1/2 gap-1"
                  >
                    {#if st.id}
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onclick={() => copySystemType(index)}
                        aria-label={$t('common.copy')}
                        title={$t('common.copy')}
                        class="pointer-events-auto h-7 w-7"
                      >
                        <CopyIcon class="h-4 w-4" />
                        <span class="sr-only">{$t('common.copy')}</span>
                      </Button>
                    {/if}
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onclick={() => removeSystemType(index)}
                      aria-label={$t('common.remove')}
                      title={$t('common.remove')}
                      class="pointer-events-auto h-7 w-7 text-muted-foreground hover:text-destructive"
                    >
                      <Trash2Icon class="h-4 w-4" />
                      <span class="sr-only">{$t('common.remove')}</span>
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          {/each}
        </div>
      {/if}
      {#if combinedFieldError('system_types')}
        <p class="text-sm text-destructive">{combinedFieldError('system_types')}</p>
      {/if}
    </div>
  {/if}

  {#if error || liveValidation.error}
    <p class="text-sm text-destructive">{error || liveValidation.error}</p>
  {/if}

  <div class="flex justify-end gap-2 pt-2">
    <Button type="button" variant="ghost" onclick={onCancel}>{$t('common.cancel')}</Button>
    <Button type="submit" disabled={loading}>
      {initialData ? $t('common.update') : $t('common.create')}
    </Button>
  </div>
</form>
