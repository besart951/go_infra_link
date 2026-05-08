<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { page } from '$app/stores';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Skeleton } from '$lib/components/ui/skeleton/index.js';
  import * as Tabs from '$lib/components/ui/tabs/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import { addToast } from '$lib/components/toast.svelte';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import UserAvatar from '$lib/components/user-avatar.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { t as translate } from '$lib/i18n/index.js';
  import HistoryTimelineDialog from '$lib/components/history/HistoryTimelineDialog.svelte';
  import { projectDetailService } from '$lib/components/project/ProjectDetailService.js';
  import { FieldDeviceLookupService } from '$lib/components/facility/field-device/state/fieldDeviceLookupService.js';
  import type { ControlCabinet, FieldDevice, SPSController } from '$lib/domain/facility/index.js';
  import type { Project } from '$lib/domain/project/index.js';
  import type { User } from '$lib/domain/user/index.js';
  import type { FieldDeviceRefreshRequest } from '$lib/components/facility/field-device/state/types.js';
  import type {
    EntityChangeEvent,
    EntityDeltaRequest,
    EntityRefreshRequest
  } from '$lib/components/facility/shared/entityRefresh.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { ProjectCollaborationState } from '$lib/services/projectCollaboration.svelte.js';
  import { Cpu, History, PanelsTopLeft, Settings, Server, Wifi, WifiOff } from '@lucide/svelte';

  type FacilityTabId = 'control-cabinets' | 'sps-controllers' | 'field-devices';

  const t = createTranslator();
  const projectId = $derived($page.params.id ?? '');
  const canOpenProjectSettings = $derived(canPerform('update', 'project'));
  const canReadProjectControlCabinets = $derived(canPerform('read', 'project.controlcabinet'));
  const canReadProjectSPSControllers = $derived(canPerform('read', 'project.spscontroller'));
  const canReadProjectFieldDevices = $derived(canPerform('read', 'project.fielddevice'));
  const canReadTimeline = $derived(canPerform('read', 'timeline'));
  const availableFacilityTabs = $derived.by((): FacilityTabId[] => {
    const tabs: FacilityTabId[] = [];
    if (canReadProjectControlCabinets) tabs.push('control-cabinets');
    if (canReadProjectSPSControllers) tabs.push('sps-controllers');
    if (canReadProjectFieldDevices) tabs.push('field-devices');
    return tabs;
  });

  let project = $state<Project | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let projectUsers = $state<User[]>([]);

  let activeFacilityTab = $state<FacilityTabId | ''>('control-cabinets');
  let projectHistoryOpen = $state(false);

  let controlCabinetViewRefreshKey = $state(0);
  let controlCabinetOptionsRefreshKey = $state(0);
  let controlCabinetRefreshRequest = $state<EntityRefreshRequest | undefined>(undefined);
  let controlCabinetDeltaRequest = $state<EntityDeltaRequest<ControlCabinet> | undefined>(
    undefined
  );
  let spsControllerRefreshKey = $state(0);
  let spsControllerRefreshRequest = $state<EntityRefreshRequest | undefined>(undefined);
  let spsControllerDeltaRequest = $state<EntityDeltaRequest<SPSController> | undefined>(undefined);
  let spsControllerCabinetLabelRefreshRequest = $state<EntityRefreshRequest | undefined>(undefined);
  let spsControllerCabinetLabelDeltaRequest = $state<
    EntityDeltaRequest<ControlCabinet> | undefined
  >(undefined);
  let fieldDeviceRefreshKey = $state(0);
  let fieldDeviceRefreshRequest = $state<FieldDeviceRefreshRequest | undefined>(undefined);
  let systemTypeRefreshKey = $state(0);
  let fieldDeviceMultiCreateFormOpen = $state(false);
  let entityRefreshRequestVersion = 0;
  let entityDeltaRequestVersion = 0;
  let fieldDeviceRefreshRequestVersion = 0;

  type ControlCabinetListViewModule =
    typeof import('$lib/components/facility/control-cabinets/ControlCabinetListView.svelte');
  type SPSControllerListViewModule =
    typeof import('$lib/components/facility/sps-controllers/SPSControllerListView.svelte');
  type FieldDeviceListViewModule =
    typeof import('$lib/components/facility/field-device/FieldDeviceListView.svelte');

  let controlCabinetListViewModule = $state<ControlCabinetListViewModule | null>(null);
  let spsControllerListViewModule = $state<SPSControllerListViewModule | null>(null);
  let fieldDeviceListViewModule = $state<FieldDeviceListViewModule | null>(null);

  let controlCabinetListViewLoad: Promise<ControlCabinetListViewModule> | null = null;
  let spsControllerListViewLoad: Promise<SPSControllerListViewModule> | null = null;
  let fieldDeviceListViewLoad: Promise<FieldDeviceListViewModule> | null = null;

  const collaboration = new ProjectCollaborationState({
    onEntityDelta: (message) => {
      if (message.actor_id && message.actor_id === currentUser?.id) return;

      switch (message.scope) {
        case 'control_cabinet':
          if (message.control_cabinets && message.control_cabinets.length > 0) {
            requestControlCabinetDelta(message.control_cabinets);
            requestSPSControllerCabinetLabelDelta(message.control_cabinets);
            bumpControlCabinetOptionsRefresh();
            bumpSPSControllerRefresh();
            bumpFieldDeviceRefresh();
            bumpSystemTypeRefresh();
          }
          break;
        case 'sps_controller':
          if (message.sps_controllers && message.sps_controllers.length > 0) {
            requestSPSControllerDelta(message.sps_controllers);
            requestFieldDeviceSPSControllerDelta(message.sps_controllers);
            bumpSystemTypeRefresh();
          }
          break;
        case 'field_device':
          if (message.field_devices && message.field_devices.length > 0) {
            requestFieldDeviceDelta(message.field_devices);
          }
          break;
      }
    },
    onRefreshRequest: (message) => {
      if (message.actor_id && message.actor_id === currentUser?.id) return;

      switch (message.scope) {
        case 'field_device':
          if (message.device_ids && message.device_ids.length > 0) {
            requestFieldDeviceRefresh(message.device_ids);
            break;
          }

          bumpFieldDeviceRefresh();
          break;
        case 'control_cabinet':
          if (message.entity_ids && message.entity_ids.length > 0) {
            requestControlCabinetRefresh(message.entity_ids);
            requestSPSControllerCabinetLabelRefresh(message.entity_ids);
            bumpControlCabinetOptionsRefresh();
            break;
          }

          bumpControlCabinetViewRefresh();
          bumpControlCabinetOptionsRefresh();
          bumpSPSControllerRefresh();
          break;
        case 'sps_controller':
          if (message.entity_ids && message.entity_ids.length > 0) {
            requestSPSControllerRefresh(message.entity_ids);
            requestFieldDeviceSPSControllerRefresh(message.entity_ids);
            bumpSystemTypeRefresh();
            break;
          }

          bumpSPSControllerRefresh();
          bumpFieldDeviceRefresh();
          bumpSystemTypeRefresh();
          break;
        case 'project':
          refreshProjectFacilityViews();
          void loadProject();
          break;
        case 'project_users':
          void loadProjectUsers();
          break;
      }
    },
    onReconnect: () => {
      refreshProjectFacilityViews();
      void loadProject();
      void loadProjectUsers();
    }
  });

  const currentUser = $derived(($page.data.user as User | null) ?? null);
  const usersById = $derived.by(() => {
    const users = new Map<string, User>();
    for (const user of projectUsers) {
      users.set(user.id, user);
    }
    if (currentUser) {
      users.set(currentUser.id, currentUser);
    }
    return users;
  });

  const onlineCollaborators = $derived.by(() =>
    collaboration.onlineUsers.map((presence) => ({
      presence,
      user: usersById.get(presence.user_id)
    }))
  );

  const fieldDeviceEditorsByDevice = $derived.by(() =>
    collaboration.buildFieldDeviceEditorsByDevice(usersById, currentUser?.id)
  );

  function preloadControlCabinetTabView(): void {
    if (controlCabinetListViewModule || controlCabinetListViewLoad) return;

    controlCabinetListViewLoad =
      import('$lib/components/facility/control-cabinets/ControlCabinetListView.svelte')
        .then((module) => {
          controlCabinetListViewModule = module;
          return module;
        })
        .catch((error) => {
          console.error('Failed to load control cabinet list view:', error);
          throw error;
        })
        .finally(() => {
          controlCabinetListViewLoad = null;
        });
  }

  function preloadSPSControllerTabView(): void {
    if (spsControllerListViewModule || spsControllerListViewLoad) return;

    spsControllerListViewLoad =
      import('$lib/components/facility/sps-controllers/SPSControllerListView.svelte')
        .then((module) => {
          spsControllerListViewModule = module;
          return module;
        })
        .catch((error) => {
          console.error('Failed to load SPS controller list view:', error);
          throw error;
        })
        .finally(() => {
          spsControllerListViewLoad = null;
        });
  }

  function preloadFieldDeviceTabView(): void {
    if (fieldDeviceListViewModule || fieldDeviceListViewLoad) return;

    fieldDeviceListViewLoad =
      import('$lib/components/facility/field-device/FieldDeviceListView.svelte')
        .then((module) => {
          fieldDeviceListViewModule = module;
          return module;
        })
        .catch((error) => {
          console.error('Failed to load field device list view:', error);
          throw error;
        })
        .finally(() => {
          fieldDeviceListViewLoad = null;
        });
  }

  $effect(() => {
    if (availableFacilityTabs.length === 0) return;
    if (!availableFacilityTabs.includes(activeFacilityTab as FacilityTabId)) {
      activeFacilityTab = availableFacilityTabs[0] ?? '';
    }
  });

  $effect(() => {
    if (!availableFacilityTabs.includes(activeFacilityTab as FacilityTabId)) return;

    if (activeFacilityTab === 'control-cabinets') {
      preloadControlCabinetTabView();
    } else if (activeFacilityTab === 'sps-controllers') {
      preloadSPSControllerTabView();
    } else if (activeFacilityTab === 'field-devices') {
      preloadFieldDeviceTabView();
    }
  });

  function bumpControlCabinetViewRefresh(): void {
    controlCabinetViewRefreshKey += 1;
  }

  function bumpControlCabinetOptionsRefresh(): void {
    controlCabinetOptionsRefreshKey += 1;
  }

  function bumpSPSControllerRefresh(): void {
    spsControllerRefreshKey += 1;
  }

  function bumpFieldDeviceRefresh(): void {
    fieldDeviceRefreshKey += 1;
  }

  function requestFieldDeviceRefresh(deviceIds: string[]): void {
    fieldDeviceRefreshRequestVersion += 1;
    fieldDeviceRefreshRequest = {
      key: fieldDeviceRefreshRequestVersion,
      deviceIds: [...deviceIds]
    };
  }

  function requestFieldDeviceSPSControllerRefresh(spsControllerIds: string[]): void {
    fieldDeviceRefreshRequestVersion += 1;
    fieldDeviceRefreshRequest = {
      key: fieldDeviceRefreshRequestVersion,
      spsControllerIds: [...spsControllerIds]
    };
  }

  function nextEntityRefreshRequest(entityIds: string[]): EntityRefreshRequest {
    entityRefreshRequestVersion += 1;
    return {
      key: entityRefreshRequestVersion,
      entityIds: [...entityIds]
    };
  }

  function nextEntityDeltaRequest<T>(items: T[]): EntityDeltaRequest<T> {
    entityDeltaRequestVersion += 1;
    return {
      key: entityDeltaRequestVersion,
      items: [...items]
    };
  }

  function requestControlCabinetRefresh(entityIds: string[]): void {
    controlCabinetRefreshRequest = nextEntityRefreshRequest(entityIds);
  }

  function requestControlCabinetDelta(items: ControlCabinet[]): void {
    controlCabinetDeltaRequest = nextEntityDeltaRequest(items);
  }

  function requestSPSControllerRefresh(entityIds: string[]): void {
    spsControllerRefreshRequest = nextEntityRefreshRequest(entityIds);
  }

  function requestSPSControllerDelta(items: SPSController[]): void {
    spsControllerDeltaRequest = nextEntityDeltaRequest(items);
  }

  function requestSPSControllerCabinetLabelRefresh(entityIds: string[]): void {
    spsControllerCabinetLabelRefreshRequest = nextEntityRefreshRequest(entityIds);
  }

  function requestSPSControllerCabinetLabelDelta(items: ControlCabinet[]): void {
    spsControllerCabinetLabelDeltaRequest = nextEntityDeltaRequest(items);
  }

  function requestFieldDeviceDelta(devices: FieldDevice[]): void {
    fieldDeviceRefreshRequestVersion += 1;
    fieldDeviceRefreshRequest = {
      key: fieldDeviceRefreshRequestVersion,
      devices: [...devices]
    };
  }

  function requestFieldDeviceSPSControllerDelta(controllers: SPSController[]): void {
    fieldDeviceRefreshRequestVersion += 1;
    fieldDeviceRefreshRequest = {
      key: fieldDeviceRefreshRequestVersion,
      spsControllers: [...controllers]
    };
  }

  function bumpSystemTypeRefresh(): void {
    if (!fieldDeviceMultiCreateFormOpen) {
      return;
    }

    systemTypeRefreshKey += 1;
  }

  function refreshProjectFacilityViews(): void {
    bumpControlCabinetViewRefresh();
    bumpControlCabinetOptionsRefresh();
    bumpSPSControllerRefresh();
    bumpFieldDeviceRefresh();
    bumpSystemTypeRefresh();
  }

  function handleControlCabinetsChanged(event?: EntityChangeEvent<ControlCabinet>): void {
    bumpControlCabinetOptionsRefresh();
    bumpSPSControllerRefresh();
    bumpFieldDeviceRefresh();
    bumpSystemTypeRefresh();

    if (event?.items && event.items.length > 0) {
      requestSPSControllerCabinetLabelDelta(event.items);
      return;
    }

    if (event?.entityIds && event.entityIds.length > 0) {
      requestSPSControllerCabinetLabelRefresh(event.entityIds);
    }
  }

  function handleSPSControllersChanged(event?: EntityChangeEvent<SPSController>): void {
    if (event?.items && event.items.length > 0) {
      requestFieldDeviceSPSControllerDelta(event.items);
    } else if (event?.entityIds && event.entityIds.length > 0) {
      requestFieldDeviceSPSControllerRefresh(event.entityIds);
    } else {
      bumpFieldDeviceRefresh();
    }

    bumpSystemTypeRefresh();
  }

  async function loadProject(): Promise<void> {
    if (!projectId) return;

    loading = true;
    error = null;

    try {
      project = await projectDetailService.getProject(projectId);
    } catch (loadError) {
      const message =
        loadError instanceof Error ? loadError.message : translate('projects.errors.load_failed');
      error = message;
      addToast(message, 'error');
    } finally {
      loading = false;
    }
  }

  async function loadProjectUsers(): Promise<void> {
    if (!projectId || !canOpenProjectSettings) {
      projectUsers = [];
      return;
    }

    try {
      const response = await projectDetailService.listUsers(projectId);
      projectUsers = response.items;
    } catch (loadError) {
      console.error('Failed to load project users', loadError);
    }
  }

  onMount(() => {
    void loadProject();
    void loadProjectUsers();
    collaboration.connect(projectId);
  });

  onDestroy(() => {
    collaboration.disconnect();
    FieldDeviceLookupService.resetAllCachedLookups();
  });
</script>

<ConfirmDialog />
<HistoryTimelineDialog
  bind:open={projectHistoryOpen}
  title={$t('history.project_title')}
  {projectId}
  onRestored={async () => {
    refreshProjectFacilityViews();
    await loadProject();
  }}
/>

<div class="flex min-w-0 flex-col gap-6 overflow-x-hidden">
  <EntityListHeader
    title={project?.name ?? $t('project.project')}
    description={$t('projects.detail.description')}
    backHref="/projects"
    backLabel={$t('common.back')}
  >
    <div class="flex items-center gap-3">
      <div
        class="flex items-center gap-2 rounded-full border bg-card px-3 py-1.5 text-sm text-muted-foreground"
      >
        {#if collaboration.socketStatus === 'connected'}
          <Wifi class="h-4 w-4 text-success" />
        {:else}
          <WifiOff class="h-4 w-4 text-warning" />
        {/if}
        <span>{onlineCollaborators.length}</span>
        <div class="flex -space-x-2">
          {#each onlineCollaborators.slice(0, 4) as collaborator}
            {#if collaborator.user}
              <Tooltip.Root>
                <Tooltip.Trigger>
                  <UserAvatar
                    firstName={collaborator.user.first_name}
                    lastName={collaborator.user.last_name}
                    class="h-7 w-7 border-2 border-background"
                  />
                </Tooltip.Trigger>
                <Tooltip.Content>
                  {collaborator.user.first_name}
                  {collaborator.user.last_name}
                </Tooltip.Content>
              </Tooltip.Root>
            {/if}
          {/each}
        </div>
      </div>

      {#if canReadTimeline}
        <Tooltip.Root>
          <Tooltip.Trigger>
            <Button
              variant="ghost"
              size="icon"
              aria-label={$t('history.open')}
              onclick={() => (projectHistoryOpen = true)}
            >
              <History />
            </Button>
          </Tooltip.Trigger>

          <Tooltip.Content>
            {$t('history.open')}
          </Tooltip.Content>
        </Tooltip.Root>
      {/if}

      {#if canOpenProjectSettings}
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
      {/if}
    </div>
  </EntityListHeader>

  {#if error}
    <div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
      <p class="font-medium">{$t('projects.errors.load_title')}</p>
      <p class="text-sm">{error}</p>
    </div>
  {/if}

  {#if loading}
    <div class="min-w-0 rounded-lg border bg-card p-6">
      <div class="grid gap-4 md:grid-cols-2">
        {#each Array(6) as _}
          <Skeleton class="h-6 w-full" />
        {/each}
      </div>
    </div>
  {:else if !project}
    <div class="min-w-0 rounded-lg border bg-card p-6 text-sm text-muted-foreground">
      {$t('projects.errors.not_found')}
    </div>
  {:else if availableFacilityTabs.length === 0}
    <div class="min-w-0 rounded-lg border bg-card p-6 text-sm text-muted-foreground">
      {$t('errors.forbidden')}
    </div>
  {:else}
    <Tabs.Root bind:value={activeFacilityTab} class="min-w-0">
      <Tabs.List class="w-full justify-start overflow-x-auto sm:w-fit">
        {#if canReadProjectControlCabinets}
          <Tabs.Trigger value="control-cabinets" class="gap-2">
            <PanelsTopLeft class="size-4" />
            {$t('projects.control_cabinets.title')}
          </Tabs.Trigger>
        {/if}
        {#if canReadProjectSPSControllers}
          <Tabs.Trigger value="sps-controllers" class="gap-2">
            <Cpu class="size-4" />
            {$t('projects.sps_controllers.title')}
          </Tabs.Trigger>
        {/if}
        {#if canReadProjectFieldDevices}
          <Tabs.Trigger value="field-devices" class="gap-2">
            <Server class="size-4" />
            {$t('projects.field_devices.title')}
          </Tabs.Trigger>
        {/if}
      </Tabs.List>

      {#if canReadProjectControlCabinets}
        <Tabs.Content value="control-cabinets" class="mt-4 min-w-0">
          <div class="min-w-0 rounded-lg border bg-card p-6">
            {#if controlCabinetListViewModule}
              {@const ControlCabinetListView = controlCabinetListViewModule.default}
              <ControlCabinetListView
                {projectId}
                refreshKey={controlCabinetViewRefreshKey}
                refreshRequest={controlCabinetRefreshRequest}
                deltaRequest={controlCabinetDeltaRequest}
                onChanged={handleControlCabinetsChanged}
              />
            {:else}
              <Skeleton class="h-6 w-full" />
            {/if}
          </div>
        </Tabs.Content>
      {/if}

      {#if canReadProjectSPSControllers}
        <Tabs.Content value="sps-controllers" class="mt-4 min-w-0">
          <div class="min-w-0 rounded-lg border bg-card p-6">
            {#if spsControllerListViewModule}
              {@const SPSControllerListView = spsControllerListViewModule.default}
              <SPSControllerListView
                {projectId}
                refreshKey={spsControllerRefreshKey}
                refreshRequest={spsControllerRefreshRequest}
                deltaRequest={spsControllerDeltaRequest}
                controlCabinetLabelRefreshRequest={spsControllerCabinetLabelRefreshRequest}
                controlCabinetLabelDeltaRequest={spsControllerCabinetLabelDeltaRequest}
                controlCabinetRefreshKey={controlCabinetOptionsRefreshKey}
                onChanged={handleSPSControllersChanged}
              />
            {:else}
              <Skeleton class="h-6 w-full" />
            {/if}
          </div>
        </Tabs.Content>
      {/if}

      {#if canReadProjectFieldDevices}
        <Tabs.Content value="field-devices" class="mt-4 min-w-0">
          <div class="min-w-0 rounded-lg border bg-card p-6">
            {#if fieldDeviceListViewModule}
              {@const FieldDeviceListView = fieldDeviceListViewModule.default}
              <FieldDeviceListView
                {projectId}
                pageSize={100}
                refreshKey={fieldDeviceRefreshKey}
                refreshRequest={fieldDeviceRefreshRequest}
                {systemTypeRefreshKey}
                onMultiCreateFormVisibilityChange={(open) => {
                  fieldDeviceMultiCreateFormOpen = open;
                }}
                sharedFieldDeviceEditors={fieldDeviceEditorsByDevice}
                onSharedFieldDeviceStateChange={(state) =>
                  collaboration.publishFieldDeviceDraftState(state)}
                onFieldDevicesSaved={(devices) => collaboration.publishFieldDeviceDelta(devices)}
              />
            {:else}
              <Skeleton class="h-6 w-full" />
            {/if}
          </div>
        </Tabs.Content>
      {/if}
    </Tabs.Root>
  {/if}
</div>
