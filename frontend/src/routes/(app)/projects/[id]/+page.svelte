<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { page } from '$app/stores';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Skeleton } from '$lib/components/ui/skeleton/index.js';
  import * as Collapsible from '$lib/components/ui/collapsible/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import { addToast } from '$lib/components/toast.svelte';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import UserAvatar from '$lib/components/user-avatar.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { t as translate } from '$lib/i18n/index.js';
  import ControlCabinetListView from '$lib/components/facility/control-cabinets/ControlCabinetListView.svelte';
  import SPSControllerListView from '$lib/components/facility/sps-controllers/SPSControllerListView.svelte';
  import FieldDeviceListView from '$lib/components/facility/field-device/FieldDeviceListView.svelte';
  import { projectDetailService } from '$lib/components/project/ProjectDetailService.js';
  import type { ControlCabinet, FieldDevice, SPSController } from '$lib/domain/facility/index.js';
  import type { Project } from '$lib/domain/project/index.js';
  import type { User } from '$lib/domain/user/index.js';
  import type { FieldDeviceRefreshRequest } from '$lib/components/facility/field-device/state/types.js';
  import type {
    EntityChangeEvent,
    EntityDeltaRequest,
    EntityRefreshRequest
  } from '$lib/components/facility/shared/entityRefresh.js';
  import { ProjectCollaborationState } from '$lib/services/projectCollaboration.svelte.js';
  import { ChevronDown, Settings, Wifi, WifiOff } from '@lucide/svelte';

  const t = createTranslator();
  const projectId = $derived($page.params.id ?? '');

  let project = $state<Project | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let projectUsers = $state<User[]>([]);

  let controlCabinetOpen = $state(true);
  let spsControllerOpen = $state(true);

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

  const collaboration = new ProjectCollaborationState({
    onEntityDelta: (message) => {
      if (message.actor_id && message.actor_id === currentUser?.id) return;

      switch (message.scope) {
        case 'control_cabinet':
          if (message.control_cabinets && message.control_cabinets.length > 0) {
            requestControlCabinetDelta(message.control_cabinets);
            requestSPSControllerCabinetLabelDelta(message.control_cabinets);
            bumpControlCabinetOptionsRefresh();
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
    if (!projectId) return;

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
  });
</script>

<ConfirmDialog />

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
  {:else}
    <div class="grid min-w-0 gap-6">
      <div class="min-w-0 rounded-lg border bg-card p-6">
        <Collapsible.Root bind:open={controlCabinetOpen} class="group/collapsible">
          <div class="flex items-center gap-3">
            <Collapsible.Trigger class="rounded-md px-2 py-1 hover:bg-accent">
              <ChevronDown
                class="size-4 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-180"
              />
            </Collapsible.Trigger>
            <h2 class="text-lg font-semibold">{$t('projects.control_cabinets.title')}</h2>
          </div>

          <Collapsible.Content class="mt-4">
            <ControlCabinetListView
              {projectId}
              refreshKey={controlCabinetViewRefreshKey}
              refreshRequest={controlCabinetRefreshRequest}
              deltaRequest={controlCabinetDeltaRequest}
              onChanged={handleControlCabinetsChanged}
            />
          </Collapsible.Content>
        </Collapsible.Root>
      </div>

      <div class="min-w-0 rounded-lg border bg-card p-6">
        <Collapsible.Root bind:open={spsControllerOpen} class="group/collapsible">
          <div class="flex items-center gap-3">
            <Collapsible.Trigger class="rounded-md px-2 py-1 hover:bg-accent">
              <ChevronDown
                class="size-4 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-180"
              />
            </Collapsible.Trigger>
            <h2 class="text-lg font-semibold">{$t('projects.sps_controllers.title')}</h2>
          </div>

          <Collapsible.Content class="mt-4">
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
          </Collapsible.Content>
        </Collapsible.Root>
      </div>

      <div class="min-w-0 rounded-lg border bg-card p-6">
        <div class="mb-4">
          <h2 class="text-lg font-semibold">{$t('projects.field_devices.title')}</h2>
        </div>
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
      </div>
    </div>
  {/if}
</div>
