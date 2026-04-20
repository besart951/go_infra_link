<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Skeleton } from '$lib/components/ui/skeleton/index.js';
  import * as Collapsible from '$lib/components/ui/collapsible/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import { addToast } from '$lib/components/toast.svelte';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import UserAvatar from '$lib/components/user-avatar.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { t as translate } from '$lib/i18n/index.js';
  import ControlCabinetListView from '$lib/components/facility/control-cabinets/ControlCabinetListView.svelte';
  import SPSControllerListView from '$lib/components/facility/sps-controllers/SPSControllerListView.svelte';
  import FieldDeviceListView from '$lib/components/facility/field-device/FieldDeviceListView.svelte';
  import { getProject } from '$lib/infrastructure/api/project.adapter.js';
  import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
  import type { Project } from '$lib/domain/project/index.js';
  import type { User } from '$lib/domain/user/index.js';
  import { ProjectCollaborationState } from '$lib/services/projectCollaboration.svelte.js';
  import { ArrowLeft, ChevronDown, Settings, Wifi, WifiOff } from '@lucide/svelte';

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
  let spsControllerRefreshKey = $state(0);
  let fieldDeviceRefreshKey = $state(0);
  let systemTypeRefreshKey = $state(0);

  let projectEventsSource: EventSource | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let pendingSseRefresh = $state(false);

  const collaboration = new ProjectCollaborationState({
    onRefreshRequest: (message) => {
      if (message.actor_id && message.actor_id === currentUser?.id) return;

      switch (message.scope) {
        case 'field_device':
          bumpFieldDeviceRefresh();
          break;
        case 'control_cabinet':
          bumpControlCabinetViewRefresh();
          bumpControlCabinetOptionsRefresh();
          bumpSPSControllerRefresh();
          bumpFieldDeviceRefresh();
          bumpSystemTypeRefresh();
          break;
        case 'sps_controller':
          bumpSPSControllerRefresh();
          bumpFieldDeviceRefresh();
          bumpSystemTypeRefresh();
          break;
      }
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

  function bumpSystemTypeRefresh(): void {
    systemTypeRefreshKey += 1;
  }

  function refreshProjectFacilityViews(): void {
    bumpControlCabinetViewRefresh();
    bumpControlCabinetOptionsRefresh();
    bumpSPSControllerRefresh();
    bumpFieldDeviceRefresh();
    bumpSystemTypeRefresh();
  }

  function handleControlCabinetsChanged(): void {
    bumpControlCabinetOptionsRefresh();
    bumpSPSControllerRefresh();
    bumpFieldDeviceRefresh();
    bumpSystemTypeRefresh();
  }

  function handleSPSControllersChanged(): void {
    bumpFieldDeviceRefresh();
    bumpSystemTypeRefresh();
  }

  function clearProjectEventsConnection(): void {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }

    if (projectEventsSource) {
      projectEventsSource.close();
      projectEventsSource = null;
    }
  }

  function queueProjectRefreshFromEvent(): void {
    if (pendingSseRefresh) return;
    pendingSseRefresh = true;

    setTimeout(() => {
      pendingSseRefresh = false;
      refreshProjectFacilityViews();
    }, 200);
  }

  function connectProjectEvents(): void {
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

  async function loadProject(): Promise<void> {
    if (!projectId) return;

    loading = true;
    error = null;

    try {
      project = await getProject(projectId);
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
      const response = await projectRepository.listUsers(projectId);
      projectUsers = response.items;
    } catch (loadError) {
      console.error('Failed to load project users', loadError);
    }
  }

  onMount(() => {
    void loadProject();
    void loadProjectUsers();
    connectProjectEvents();
    collaboration.connect(projectId);
  });

  onDestroy(() => {
    clearProjectEventsConnection();
    collaboration.disconnect();
  });
</script>

<ConfirmDialog />

<div class="flex min-w-0 flex-col gap-6 overflow-x-hidden">
  <div class="flex min-w-0 items-start gap-3">
    <Button variant="outline" onclick={() => goto('/projects')}>
      <ArrowLeft class="mr-2 h-4 w-4" />
      {$t('common.back')}
    </Button>
    <div class="min-w-0">
      <h1 class="text-3xl font-bold tracking-tight">{project?.name ?? $t('project.project')}</h1>
      <p class="mt-1 text-muted-foreground">{$t('projects.detail.description')}</p>
    </div>
    <div class="ml-auto">
      <div class="flex items-center gap-3">
        <div
          class="flex items-center gap-2 rounded-full border bg-card px-3 py-1.5 text-sm text-muted-foreground"
        >
          {#if collaboration.socketStatus === 'connected'}
            <Wifi class="h-4 w-4 text-emerald-600" />
          {:else}
            <WifiOff class="h-4 w-4 text-amber-600" />
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
    </div>
  </div>

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
            <Collapsible.Trigger class="rounded px-2 py-1 hover:bg-accent">
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
              onChanged={handleControlCabinetsChanged}
            />
          </Collapsible.Content>
        </Collapsible.Root>
      </div>

      <div class="min-w-0 rounded-lg border bg-card p-6">
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
            <SPSControllerListView
              {projectId}
              refreshKey={spsControllerRefreshKey}
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
          refreshKey={fieldDeviceRefreshKey}
          {systemTypeRefreshKey}
          sharedFieldDeviceEditors={fieldDeviceEditorsByDevice}
          onSharedFieldDeviceStateChange={(state) =>
            collaboration.publishFieldDeviceDraftState(state)}
          onFieldDevicesSaved={(deviceIds) => collaboration.requestFieldDeviceRefresh(deviceIds)}
        />
      </div>
    </div>
  {/if}
</div>
