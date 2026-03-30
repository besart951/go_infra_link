<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Skeleton } from '$lib/components/ui/skeleton/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import * as Collapsible from '$lib/components/ui/collapsible/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import { addToast } from '$lib/components/toast.svelte';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import { confirm } from '$lib/stores/confirm-dialog.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { t as translate } from '$lib/i18n/index.js';
  import ControlCabinetList from '$lib/components/facility/control-cabinets/ControlCabinetList.svelte';
  import SPSControllerForm from '$lib/components/facility/forms/SPSControllerForm.svelte';
  import FieldDeviceListView from '$lib/components/facility/field-device/FieldDeviceListView.svelte';
  import {
    getProject,
    listProjectControlCabinets,
    addProjectControlCabinet,
    copyProjectControlCabinet,
    removeProjectControlCabinet,
    listProjectSPSControllers,
    addProjectSPSController,
    copyProjectSPSController
  } from '$lib/infrastructure/api/project.adapter.js';
  import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
  import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
  import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
  import type { Project } from '$lib/domain/project/index.js';
  import type {
    ProjectControlCabinetLink,
    ProjectSPSControllerLink
  } from '$lib/domain/project/index.js';
  import type { Building, ControlCabinet, SPSController } from '$lib/domain/facility/index.js';
  import { ArrowLeft, Plus, ChevronDown, Settings } from '@lucide/svelte';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import { canPerform } from '$lib/utils/permissions.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';

  const t = createTranslator();

  const projectId = $derived($page.params.id ?? '');

  let project = $state<Project | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  let controlCabinetLinks = $state<ProjectControlCabinetLink[]>([]);
  let controlCabinetOptions = $state<ControlCabinet[]>([]);
  let controlCabinetLoading = $state(false);
  let showControlCabinetForm = $state(false);
  let editingControlCabinet: ControlCabinet | undefined = $state(undefined);
  let controlCabinetOpen = $state(true);
  let controlCabinetSearch = $state('');
  let controlCabinetPage = $state(1);
  const controlCabinetPageSize = 10;
  let buildingMap = $state(new Map<string, string>());
  const buildingRequests = new Set<string>();

  let spsControllerLinks = $state<ProjectSPSControllerLink[]>([]);
  let spsControllerOptions = $state<SPSController[]>([]);
  let spsControllerLoading = $state(false);
  let showSpsControllerForm = $state(false);
  let spsControllerOpen = $state(true);
  let editingSpsController: SPSController | undefined = $state(undefined);
  let spsControllerSearchText = $state('');
  let spsControllerPage = $state(1);
  const spsControllerPageSize = 10;
  let controlCabinetRefreshKey = $state(0);
  let fieldDeviceRefreshKey = $state(0);
  let systemTypeRefreshKey = $state(0);
  let projectEventsSource: EventSource | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let pendingSseRefresh = $state(false);

  const controlCabinetLinkMap = $derived.by(
    () => new Map(controlCabinetLinks.map((link) => [link.control_cabinet_id, link]))
  );

  const linkedControlCabinets = $derived.by(() =>
    mergeControlCabinetOptions(
      controlCabinetOptions.filter((cabinet) => controlCabinetLinkMap.has(cabinet.id))
    )
  );

  const filteredControlCabinets = $derived.by(() => {
    const query = controlCabinetSearch.trim().toLowerCase();
    if (!query) return linkedControlCabinets;
    return linkedControlCabinets.filter((cabinet) =>
      [cabinet.control_cabinet_nr, getBuildingLabel(cabinet.building_id)]
        .filter(Boolean)
        .some((value) => value!.toLowerCase().includes(query))
    );
  });

  const controlCabinetTotalPages = $derived.by(() =>
    filteredControlCabinets.length === 0
      ? 0
      : Math.ceil(filteredControlCabinets.length / controlCabinetPageSize)
  );

  const controlCabinetPageItems = $derived.by(() => {
    const start = (controlCabinetPage - 1) * controlCabinetPageSize;
    return filteredControlCabinets.slice(start, start + controlCabinetPageSize);
  });

  const controlCabinetListState = $derived.by(() => ({
    items: controlCabinetPageItems,
    total: filteredControlCabinets.length,
    page:
      controlCabinetTotalPages === 0 ? 1 : Math.min(controlCabinetPage, controlCabinetTotalPages),
    pageSize: controlCabinetPageSize,
    totalPages: controlCabinetTotalPages,
    searchText: controlCabinetSearch,
    loading: controlCabinetLoading,
    error: null
  }));

  const spsControllerLinkMap = $derived.by(
    () => new Map(spsControllerLinks.map((link) => [link.sps_controller_id, link]))
  );

  const linkedSpsControllers = $derived.by(() =>
    spsControllerOptions.filter((controller) => spsControllerLinkMap.has(controller.id))
  );

  const filteredSpsControllers = $derived.by(() => {
    const query = spsControllerSearchText.trim().toLowerCase();
    if (!query) return linkedSpsControllers;
    return linkedSpsControllers.filter((controller) =>
      [
        controller.device_name,
        controller.ga_device,
        controller.ip_address,
        controlCabinetLabel(controller.control_cabinet_id)
      ]
        .filter(Boolean)
        .some((value) => value!.toLowerCase().includes(query))
    );
  });

  const spsControllerTotalPages = $derived.by(() =>
    filteredSpsControllers.length === 0
      ? 0
      : Math.ceil(filteredSpsControllers.length / spsControllerPageSize)
  );

  const spsControllerPageItems = $derived.by(() => {
    const start = (spsControllerPage - 1) * spsControllerPageSize;
    return filteredSpsControllers.slice(start, start + spsControllerPageSize);
  });

  const spsControllerListState = $derived.by(() => ({
    items: spsControllerPageItems,
    total: filteredSpsControllers.length,
    page: spsControllerTotalPages === 0 ? 1 : Math.min(spsControllerPage, spsControllerTotalPages),
    pageSize: spsControllerPageSize,
    totalPages: spsControllerTotalPages,
    searchText: spsControllerSearchText,
    loading: spsControllerLoading,
    error: null
  }));

  function uniqueIds(ids: string[]): string[] {
    return Array.from(new Set(ids.filter(Boolean)));
  }

  async function fetchControlCabinetsByIds(ids: string[]): Promise<ControlCabinet[]> {
    const unique = uniqueIds(ids);
    if (unique.length === 0) return [];
    return controlCabinetRepository.getBulk(unique);
  }

  async function fetchSpsControllersByIds(ids: string[]): Promise<SPSController[]> {
    const unique = uniqueIds(ids);
    if (unique.length === 0) return [];
    return spsControllerRepository.getBulk(unique);
  }

  function mergeControlCabinetOptions(items: ControlCabinet[]): ControlCabinet[] {
    return Array.from(new Map(items.map((item) => [item.id, item])).values());
  }

  function formatBuildingLabel(building: Building): string {
    return `${building.iws_code}-${building.building_group}`;
  }

  function getBuildingLabel(buildingId: string | undefined | null): string {
    if (!buildingId) return '-';
    return buildingMap.get(buildingId) ?? buildingId;
  }

  function updateBuildingMap(buildings: Building[]) {
    const next = new Map(buildingMap);
    for (const building of buildings) {
      next.set(building.id, formatBuildingLabel(building));
    }
    buildingMap = next;
  }

  async function ensureBuildingLabels(items: ControlCabinet[]) {
    const uniqueIds = new Set(
      items.map((item) => item.building_id).filter((id): id is string => Boolean(id))
    );
    const missingIds = Array.from(uniqueIds).filter(
      (id) => !buildingMap.has(id) && !buildingRequests.has(id)
    );

    if (missingIds.length === 0) return;

    missingIds.forEach((id) => buildingRequests.add(id));

    try {
      const items = await buildingRepository.getBulk(missingIds);
      updateBuildingMap(items);
    } catch (err) {
      console.error('Failed to load buildings:', err);
    } finally {
      missingIds.forEach((id) => buildingRequests.delete(id));
    }
  }

  function controlCabinetLabel(id: string): string {
    const item = controlCabinetOptions.find((c) => c.id === id);
    return item?.control_cabinet_nr || item?.id || id;
  }

  function bumpControlCabinetRefresh() {
    controlCabinetRefreshKey += 1;
  }

  function bumpFieldDeviceRefresh() {
    fieldDeviceRefreshKey += 1;
  }

  function bumpSystemTypeRefresh() {
    systemTypeRefreshKey += 1;
  }

  function bumpProjectFacilityRefresh() {
    bumpControlCabinetRefresh();
    bumpFieldDeviceRefresh();
    bumpSystemTypeRefresh();
  }

  function clearProjectEventsConnection() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    if (projectEventsSource) {
      projectEventsSource.close();
      projectEventsSource = null;
    }
  }

  function queueProjectRefreshFromEvent() {
    if (pendingSseRefresh) return;
    pendingSseRefresh = true;

    setTimeout(async () => {
      pendingSseRefresh = false;
      await refreshProjectFacilityData({
        controlCabinets: true,
        spsControllers: true,
        fieldDevices: true,
        systemTypes: true
      });
    }, 200);
  }

  function connectProjectEvents() {
    if (!projectId) return;

    clearProjectEventsConnection();

    const source = new EventSource(`/api/v1/projects/${projectId}/events`);
    projectEventsSource = source;

    source.addEventListener('project.change', (event) => {
      try {
        const payload = JSON.parse((event as MessageEvent<string>).data) as { type?: string };
        if (payload.type === 'ready') return;
      } catch {
        // Ignore parse issues and still trigger refresh.
      }
      queueProjectRefreshFromEvent();
    });

    source.onerror = () => {
      source.close();
      if (projectEventsSource === source) {
        projectEventsSource = null;
      }
      if (!reconnectTimer) {
        reconnectTimer = setTimeout(() => {
          reconnectTimer = null;
          connectProjectEvents();
        }, 3000);
      }
    };
  }

  async function refreshProjectFacilityData(options?: {
    controlCabinets?: boolean;
    spsControllers?: boolean;
    fieldDevices?: boolean;
    systemTypes?: boolean;
  }) {
    const {
      controlCabinets = false,
      spsControllers = false,
      fieldDevices = false,
      systemTypes = false
    } = options ?? {};

    const tasks: Promise<void>[] = [];

    if (controlCabinets) {
      tasks.push(loadControlCabinets());
    }
    if (spsControllers) {
      tasks.push(loadSpsControllers());
    }

    await Promise.all(tasks);

    if (controlCabinets) {
      bumpControlCabinetRefresh();
    }
    if (fieldDevices) {
      bumpFieldDeviceRefresh();
    }
    if (systemTypes) {
      bumpSystemTypeRefresh();
    }
  }

  $effect(() => {
    if (controlCabinetTotalPages === 0) {
      controlCabinetPage = 1;
      return;
    }
    if (controlCabinetPage > controlCabinetTotalPages) {
      controlCabinetPage = controlCabinetTotalPages;
    }
  });

  $effect(() => {
    if (spsControllerTotalPages === 0) {
      spsControllerPage = 1;
      return;
    }
    if (spsControllerPage > spsControllerTotalPages) {
      spsControllerPage = spsControllerTotalPages;
    }
  });

  async function loadProject() {
    if (!projectId) return;
    loading = true;
    error = null;
    try {
      project = await getProject(projectId);
    } catch (err) {
      const message = err instanceof Error ? err.message : translate('projects.errors.load_failed');
      error = message;
      addToast(message, 'error');
    } finally {
      loading = false;
    }
  }

  async function loadControlCabinets() {
    if (!projectId) return;
    controlCabinetLoading = true;
    try {
      const linksRes = await listProjectControlCabinets(projectId, { page: 1, limit: 200 });
      controlCabinetLinks = linksRes.items;

      const cabinetIds = linksRes.items.map((l) => l.control_cabinet_id);
      controlCabinetOptions = mergeControlCabinetOptions(await fetchControlCabinetsByIds(cabinetIds));
      await ensureBuildingLabels(controlCabinetOptions);
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : translate('projects.control_cabinets.load_failed'),
        'error'
      );
    } finally {
      controlCabinetLoading = false;
    }
  }

  async function loadSpsControllers() {
    if (!projectId) return;
    spsControllerLoading = true;
    try {
      const linksRes = await listProjectSPSControllers(projectId, { page: 1, limit: 200 });
      spsControllerLinks = linksRes.items;

      const controllerIds = linksRes.items.map((l) => l.sps_controller_id);
      spsControllerOptions = await fetchSpsControllersByIds(controllerIds);

      const cabinetIds = spsControllerOptions.map((c) => c.control_cabinet_id);
      const existing = new Set(controlCabinetOptions.map((c) => c.id));
      const missing = uniqueIds(cabinetIds).filter((id) => !existing.has(id));
      if (missing.length > 0) {
        const fetched = await fetchControlCabinetsByIds(missing);
        controlCabinetOptions = mergeControlCabinetOptions([...controlCabinetOptions, ...fetched]);
        await ensureBuildingLabels(fetched);
      }
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : translate('projects.sps_controllers.load_failed'),
        'error'
      );
    } finally {
      spsControllerLoading = false;
    }
  }

  function handleControlCabinetCreate() {
    editingControlCabinet = undefined;
    showControlCabinetForm = true;
  }

  function handleControlCabinetEdit(item: ControlCabinet) {
    editingControlCabinet = item;
    showControlCabinetForm = true;
  }

  function handleControlCabinetCancel() {
    showControlCabinetForm = false;
    editingControlCabinet = undefined;
  }

  function handleControlCabinetSearch(text: string) {
    controlCabinetSearch = text;
    controlCabinetPage = 1;
  }

  function handleControlCabinetPageChange(page: number) {
    controlCabinetPage = page;
  }

  async function handleControlCabinetCreated(item: ControlCabinet) {
    if (!projectId) return;
    try {
      if (editingControlCabinet) {
        addToast(translate('projects.control_cabinets.updated'), 'success');
      } else {
        await addProjectControlCabinet(projectId, item.id);
        addToast(translate('projects.control_cabinets.created'), 'success');
      }
      showControlCabinetForm = false;
      editingControlCabinet = undefined;
      await refreshProjectFacilityData({
        controlCabinets: true,
        spsControllers: true,
        fieldDevices: true,
        systemTypes: true
      });
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : translate('projects.control_cabinets.link_failed'),
        'error'
      );
    }
  }

  async function removeControlCabinetByLink(linkId: string) {
    if (!projectId) return;
    const ok = await confirm({
      title: translate('projects.control_cabinets.remove_title'),
      message: translate('projects.control_cabinets.remove_message'),
      confirmText: translate('projects.control_cabinets.remove_confirm'),
      cancelText: translate('common.cancel'),
      variant: 'destructive'
    });
    if (!ok) return;
    try {
      await removeProjectControlCabinet(projectId, linkId);
      addToast(translate('projects.control_cabinets.removed'), 'success');
      await refreshProjectFacilityData({
        controlCabinets: true,
        spsControllers: true,
        fieldDevices: true,
        systemTypes: true
      });
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : translate('projects.control_cabinets.remove_failed'),
        'error'
      );
    }
  }

  async function handleControlCabinetRemove(item: ControlCabinet) {
    const link = controlCabinetLinkMap.get(item.id);
    if (!link) return;
    await removeControlCabinetByLink(link.id);
  }

  async function handleDuplicateControlCabinet(item: ControlCabinet) {
    if (!projectId) return;
    try {
      await copyProjectControlCabinet(projectId, item.id);
      addToast(translate('projects.control_cabinets.duplicated'), 'success');
      await refreshProjectFacilityData({
        controlCabinets: true,
        spsControllers: true,
        fieldDevices: true,
        systemTypes: true
      });
    } catch (err) {
      addToast(
        err instanceof Error
          ? err.message
          : translate('projects.control_cabinets.duplicate_failed'),
        'error'
      );
    }
  }

  async function handleDuplicateSpsController(item: SPSController) {
    if (!projectId) return;
    try {
      await copyProjectSPSController(projectId, item.id);
      addToast(translate('projects.sps_controllers.duplicated'), 'success');
      await refreshProjectFacilityData({
        spsControllers: true,
        fieldDevices: true,
        systemTypes: true
      });
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : translate('projects.sps_controllers.duplicate_failed'),
        'error'
      );
    }
  }
  async function handleControlCabinetCopy(value: string) {
    try {
      await navigator.clipboard.writeText(value);
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
    }
  }

  function handleSpsControllerEdit(item: SPSController) {
    editingSpsController = item;
    showSpsControllerForm = true;
  }

  function handleSpsControllerCreate() {
    editingSpsController = undefined;
    showSpsControllerForm = true;
  }

  function handleSpsControllerCancel() {
    showSpsControllerForm = false;
    editingSpsController = undefined;
  }

  function handleSpsControllerSearch(text: string) {
    spsControllerSearchText = text;
    spsControllerPage = 1;
  }

  function handleSpsControllerPageChange(page: number) {
    spsControllerPage = page;
  }

  async function handleSpsControllerSuccess(item: SPSController) {
    if (!projectId) return;
    try {
      if (!editingSpsController) {
        await addProjectSPSController(projectId, item.id);
        addToast(translate('projects.sps_controllers.created'), 'success');
      } else {
        addToast(translate('projects.sps_controllers.updated'), 'success');
      }
      showSpsControllerForm = false;
      editingSpsController = undefined;
      await refreshProjectFacilityData({
        spsControllers: true,
        fieldDevices: true,
        systemTypes: true
      });
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : translate('projects.sps_controllers.save_failed'),
        'error'
      );
    }
  }

  async function handleDeleteSpsController(item: SPSController) {
    const ok = await confirm({
      title: translate('projects.sps_controllers.delete_title'),
      message: translate('projects.sps_controllers.delete_message', { name: item.device_name }),
      confirmText: translate('common.delete'),
      cancelText: translate('common.cancel'),
      variant: 'destructive'
    });
    if (!ok) return;
    try {
      await spsControllerRepository.delete(item.id);
      addToast(translate('projects.sps_controllers.deleted'), 'success');
      await refreshProjectFacilityData({
        spsControllers: true,
        fieldDevices: true,
        systemTypes: true
      });
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : translate('projects.sps_controllers.delete_failed'),
        'error'
      );
    }
  }

  onMount(() => {
    loadProject();
    loadControlCabinets();
    loadSpsControllers();
    connectProjectEvents();
  });

  onDestroy(() => {
    clearProjectEventsConnection();
  });
</script>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <div class="flex items-start gap-3">
    <Button variant="outline" onclick={() => goto('/projects')}>
      <ArrowLeft class="mr-2 h-4 w-4" />
      {$t('common.back')}
    </Button>
    <div>
      <h1 class="text-3xl font-bold tracking-tight">{project?.name ?? $t('project.project')}</h1>
      <p class="mt-1 text-muted-foreground">{$t('projects.detail.description')}</p>
    </div>
    <div class="ml-auto">
      <Tooltip.Root>
        <Tooltip.Trigger>
          <Button variant="ghost" href={`/projects/${projectId}/settings`} size="icon">
            <Settings />
          </Button>
        </Tooltip.Trigger>

        <Tooltip.Content>
          {$t('projects.detail.settings')}
        </Tooltip.Content>
      </Tooltip.Root>
    </div>
  </div>

  {#if error}
    <div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
      <p class="font-medium">{$t('projects.errors.load_title')}</p>
      <p class="text-sm">{error}</p>
    </div>
  {/if}

  {#if loading}
    <div class="rounded-lg border bg-background p-6">
      <div class="grid gap-4 md:grid-cols-2">
        {#each Array(6) as _}
          <Skeleton class="h-6 w-full" />
        {/each}
      </div>
    </div>
  {:else if !project}
    <div class="rounded-lg border bg-background p-6 text-sm text-muted-foreground">
      {$t('projects.errors.not_found')}
    </div>
  {:else}
    <div class="grid gap-6">
      <div class="rounded-lg border bg-background p-6">
        <Collapsible.Root bind:open={controlCabinetOpen} class="group/collapsible">
          <div class="flex items-center gap-3">
            <Collapsible.Trigger class="rounded px-2 py-1 hover:bg-accent">
              <ChevronDown
                class="size-4 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-180"
              />
            </Collapsible.Trigger>
            <h2 class="text-lg font-semibold">{$t('projects.control_cabinets.title')}</h2>
          </div>

          <Collapsible.Content class="mt-">
            <ControlCabinetList
              state={controlCabinetListState}
              showForm={showControlCabinetForm}
              editingItem={editingControlCabinet}
              {projectId}
              searchPlaceholder={$t('projects.control_cabinets.search_placeholder')}
              emptyMessage={$t('projects.control_cabinets.empty')}
              cabinetColumnLabel={$t('projects.control_cabinets.table.control_cabinet')}
              buildingColumnLabel={$t('projects.control_cabinets.table.building')}
              newLabel={$t('projects.control_cabinets.new')}
              canCreate={canPerform('create', 'controlcabinet')}
              canDuplicate={canPerform('create', 'controlcabinet')}
              canUpdate={canPerform('update', 'controlcabinet')}
              canDelete={canPerform('delete', 'controlcabinet')}
              {getBuildingLabel}
              onCreate={handleControlCabinetCreate}
              onSearch={handleControlCabinetSearch}
              onPageChange={handleControlCabinetPageChange}
              onReload={loadControlCabinets}
              onFormSuccess={handleControlCabinetCreated}
              onFormCancel={handleControlCabinetCancel}
              onEdit={handleControlCabinetEdit}
              onDelete={handleControlCabinetRemove}
              onDuplicate={handleDuplicateControlCabinet}
              onCopy={handleControlCabinetCopy}
            />
          </Collapsible.Content>
        </Collapsible.Root>
      </div>

      <div class="rounded-lg border bg-background p-6">
        <Collapsible.Root bind:open={spsControllerOpen} class="group/collapsible">
          <div class="flex items-center gap-3">
            <Collapsible.Trigger class="rounded px-2 py-1 hover:bg-accent">
              <ChevronDown
                class="size-4 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-180"
              />
            </Collapsible.Trigger>
            <h2 class="text-lg font-semibold">{$t('projects.sps_controllers.title')}</h2>
          </div>
          <Collapsible.Content class="mt-4">
            <div class="flex flex-wrap items-center justify-end gap-2 mb-4">
              {#if !showSpsControllerForm}
                <Button onclick={handleSpsControllerCreate}>
                  <Plus class="mr-2 size-4" />
                  {$t('projects.sps_controllers.new')}
                </Button>
              {/if}
            </div>

            {#if showSpsControllerForm}
              <SPSControllerForm
                initialData={editingSpsController}
                {projectId}
                {controlCabinetRefreshKey}
                onSuccess={handleSpsControllerSuccess}
                onCancel={handleSpsControllerCancel}
              />
            {/if}

            <PaginatedList
              state={spsControllerListState}
              columns={[
                { key: 'device_name', label: $t('projects.sps_controllers.columns.device_name') },
                { key: 'ga_device', label: $t('projects.sps_controllers.columns.ga_device') },
                { key: 'ip_address', label: $t('projects.sps_controllers.columns.ip_address') },
                { key: 'cabinet', label: $t('projects.sps_controllers.columns.cabinet') },
                { key: 'created', label: $t('projects.sps_controllers.columns.created') },
                { key: 'actions', label: '', width: 'w-[100px]' }
              ]}
              searchPlaceholder={$t('projects.sps_controllers.search_placeholder')}
              emptyMessage={$t('projects.sps_controllers.empty')}
              onSearch={handleSpsControllerSearch}
              onPageChange={handleSpsControllerPageChange}
              onReload={loadSpsControllers}
            >
              {#snippet rowSnippet(controller: SPSController)}
                <Table.Cell class="font-medium">
                  <a href="/facility/sps-controllers/{controller.id}" class="hover:underline">
                    {controller.device_name}
                  </a>
                </Table.Cell>
                <Table.Cell>{controller.ga_device ?? '-'}</Table.Cell>
                <Table.Cell>
                  {#if controller.ip_address}
                    <code class="rounded bg-muted px-1.5 py-0.5 text-sm">
                      {controller.ip_address}
                    </code>
                  {:else}
                    -
                  {/if}
                </Table.Cell>
                <Table.Cell>{controlCabinetLabel(controller.control_cabinet_id)}</Table.Cell>
                <Table.Cell>
                  {new Date(controller.created_at).toLocaleDateString()}
                </Table.Cell>
                <Table.Cell class="text-right">
                  <DropdownMenu.Root>
                    <DropdownMenu.Trigger>
                      {#snippet child({ props })}
                        <Button variant="ghost" size="icon" {...props}>
                          <EllipsisIcon class="size-4" />
                        </Button>
                      {/snippet}
                    </DropdownMenu.Trigger>
                    <DropdownMenu.Content align="end" class="w-44">
                      {#if canPerform('create', 'spscontroller')}
                        <DropdownMenu.Item onclick={() => handleDuplicateSpsController(controller)}>
                          {$t('facility.duplicate')}
                        </DropdownMenu.Item>
                      {/if}
                      {#if canPerform('read', 'spscontroller')}
                        <DropdownMenu.Item
                          onclick={() => goto(`/facility/sps-controllers/${controller.id}`)}
                        >
                          {$t('common.view')}
                        </DropdownMenu.Item>
                      {/if}
                      {#if canPerform('update', 'spscontroller')}
                        <DropdownMenu.Item onclick={() => handleSpsControllerEdit(controller)}>
                          {$t('common.edit')}
                        </DropdownMenu.Item>
                      {/if}
                      {#if canPerform('delete', 'spscontroller')}
                        <DropdownMenu.Separator />
                        <DropdownMenu.Item
                          variant="destructive"
                          onclick={() => handleDeleteSpsController(controller)}
                        >
                          {$t('common.delete')}
                        </DropdownMenu.Item>
                      {/if}
                    </DropdownMenu.Content>
                  </DropdownMenu.Root>
                </Table.Cell>
              {/snippet}
            </PaginatedList>
          </Collapsible.Content>
        </Collapsible.Root>
      </div>

      <div class="rounded-lg border bg-background p-6">
        <div class="mb-4">
          <h2 class="text-lg font-semibold">{$t('projects.field_devices.title')}</h2>
        </div>
        <FieldDeviceListView
          {projectId}
          refreshKey={fieldDeviceRefreshKey}
          {systemTypeRefreshKey}
        />
      </div>
    </div>
  {/if}
</div>
