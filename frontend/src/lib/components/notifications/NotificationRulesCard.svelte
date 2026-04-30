<script lang="ts">
  import { getErrorMessage } from '$lib/api/client.js';
  import { getTeam, listTeams, type Team } from '$lib/api/teams.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import StaticCombobox from '$lib/components/ui/combobox/StaticCombobox.svelte';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Switch } from '$lib/components/ui/switch/index.js';
  import type {
    NotificationRule,
    NotificationRuleRecipientType,
    UpsertNotificationRuleRequest
  } from '$lib/domain/notification/index.js';
  import type {
    ControlCabinet,
    FieldDevice,
    ObjectData,
    SPSController
  } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import {
    getControlCabinet,
    getControlCabinets,
    getFieldDevice,
    getObjectData,
    getSPSController,
    getSPSControllers
  } from '$lib/infrastructure/api/facility.adapter.js';
  import { notificationRuleRepository } from '$lib/infrastructure/api/notificationRuleRepository.js';
  import { getProject, listProjects } from '$lib/infrastructure/api/project.adapter.js';
  import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
  import { listRoles } from '$lib/infrastructure/api/role.adapter.js';
  import type { Role } from '$lib/domain/role/index.js';
  import type { Project } from '$lib/domain/project/index.js';
  import PlusIcon from '@lucide/svelte/icons/plus';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';
  import { onMount } from 'svelte';

  type SelectOption = {
    id: string;
    label: string;
  };

  type NotificationResourceType =
    | ''
    | 'project'
    | 'project_user'
    | 'control_cabinet'
    | 'sps_controller'
    | 'field_device'
    | 'object_data';

  type EventDefinition = {
    id: string;
    labelKey: string;
    resourceType: NotificationResourceType;
  };

  const t = createTranslator();

  const eventDefinitions: EventDefinition[] = [
    {
      id: 'project.updated',
      labelKey: 'notifications.rules.events.project_updated',
      resourceType: 'project'
    },
    {
      id: 'project.deleted',
      labelKey: 'notifications.rules.events.project_deleted',
      resourceType: 'project'
    },
    {
      id: 'project.phase.changed',
      labelKey: 'notifications.rules.events.project_phase_changed',
      resourceType: 'project'
    },
    {
      id: 'project.user.invited',
      labelKey: 'notifications.rules.events.project_user_invited',
      resourceType: 'project_user'
    },
    {
      id: 'project.user.removed',
      labelKey: 'notifications.rules.events.project_user_removed',
      resourceType: 'project_user'
    },
    {
      id: 'project.control_cabinet.created',
      labelKey: 'notifications.rules.events.control_cabinet_created',
      resourceType: 'control_cabinet'
    },
    {
      id: 'project.control_cabinet.updated',
      labelKey: 'notifications.rules.events.control_cabinet_updated',
      resourceType: 'control_cabinet'
    },
    {
      id: 'project.control_cabinet.deleted',
      labelKey: 'notifications.rules.events.control_cabinet_deleted',
      resourceType: 'control_cabinet'
    },
    {
      id: 'project.sps_controller.created',
      labelKey: 'notifications.rules.events.sps_controller_created',
      resourceType: 'sps_controller'
    },
    {
      id: 'project.sps_controller.updated',
      labelKey: 'notifications.rules.events.sps_controller_updated',
      resourceType: 'sps_controller'
    },
    {
      id: 'project.sps_controller.deleted',
      labelKey: 'notifications.rules.events.sps_controller_deleted',
      resourceType: 'sps_controller'
    },
    {
      id: 'project.sps_controller.ip_address.changed',
      labelKey: 'notifications.rules.events.sps_controller_ip_changed',
      resourceType: 'sps_controller'
    },
    {
      id: 'project.field_device.created',
      labelKey: 'notifications.rules.events.field_device_created',
      resourceType: 'field_device'
    },
    {
      id: 'project.field_device.updated',
      labelKey: 'notifications.rules.events.field_device_updated',
      resourceType: 'field_device'
    },
    {
      id: 'project.field_device.deleted',
      labelKey: 'notifications.rules.events.field_device_deleted',
      resourceType: 'field_device'
    },
    {
      id: 'project.field_device.multi_created',
      labelKey: 'notifications.rules.events.field_device_multi_created',
      resourceType: 'field_device'
    },
    {
      id: 'project.object_data.created',
      labelKey: 'notifications.rules.events.object_data_created',
      resourceType: 'object_data'
    },
    {
      id: 'project.object_data.deleted',
      labelKey: 'notifications.rules.events.object_data_deleted',
      resourceType: 'object_data'
    }
  ];

  const resourceTypeIds: NotificationResourceType[] = [
    '',
    'project',
    'project_user',
    'control_cabinet',
    'sps_controller',
    'field_device',
    'object_data'
  ];

  const recipientTypeOptions: SelectOption[] = [
    { id: 'project_users', label: 'Projektbenutzer' },
    { id: 'project_role', label: 'Projektrolle' },
    { id: 'team', label: 'Team' },
    { id: 'users', label: 'Einzelne Benutzer' }
  ];

  const fallbackRoleOptions: SelectOption[] = [
    { id: 'superadmin', label: 'Superadmin' },
    { id: 'admin_fzag', label: 'Admin FZAG' },
    { id: 'fzag', label: 'FZAG' },
    { id: 'admin_planer', label: 'Admin Planer' },
    { id: 'planer', label: 'Planer' },
    { id: 'admin_entrepreneur', label: 'Admin Unternehmer' },
    { id: 'entrepreneur', label: 'Unternehmer' }
  ];

  let rules = $state<NotificationRule[]>([]);
  let roleOptions = $state<SelectOption[]>(fallbackRoleOptions);
  let isLoading = $state(true);
  let isSubmitting = $state(false);
  let error = $state<string | null>(null);
  let enabled = $state(true);
  let name = $state('');
  let eventKey = $state('project.phase.changed');
  let projectID = $state('');
  let resourceType = $state<NotificationResourceType>('project');
  let resourceID = $state('');
  let recipientType = $state<NotificationRuleRecipientType>('project_users');
  let recipientUserIDs = $state('');
  let recipientTeamID = $state('');
  let recipientRole = $state('');

  const eventOptions = $derived(
    eventDefinitions.map((event) => ({
      id: event.id,
      label: `${$t(event.labelKey)} · ${event.id}`
    }))
  );
  const resourceTypeOptions = $derived(
    resourceTypeIds.map((id) => ({
      id,
      label: id
        ? $t(`notifications.rules.resource_types.${id}`)
        : $t('notifications.rules.resource_types.all')
    }))
  );
  const translatedRecipientTypeOptions = $derived(
    recipientTypeOptions.map((option) => ({
      ...option,
      label: $t(`notifications.rules.recipient_${option.id}`)
    }))
  );
  const resourceRefreshKey = $derived(`${projectID}|${resourceType}`);
  const resourceDisabled = $derived(
    !resourceType || resourceType === 'project_user' || (resourceType !== 'project' && !projectID)
  );

  async function loadRules() {
    isLoading = true;
    error = null;
    try {
      const result = await notificationRuleRepository.list();
      rules = result.items;
    } catch (err) {
      error = getErrorMessage(err);
    } finally {
      isLoading = false;
    }
  }

  async function loadRoles() {
    try {
      const roles = await listRoles();
      roleOptions = roles
        .slice()
        .sort((left, right) => right.level - left.level)
        .map((role: Role) => ({
          id: role.name,
          label: role.display_name || role.name
        }));
    } catch {
      roleOptions = fallbackRoleOptions;
    }
  }

  function buildPayload(): UpsertNotificationRuleRequest {
    return {
      name: name.trim(),
      enabled,
      event_key: eventKey.trim(),
      project_id: projectID.trim() || null,
      resource_type: resourceType.trim(),
      resource_id: resourceID.trim() || null,
      recipient_type: recipientType,
      recipient_user_ids:
        recipientType === 'users'
          ? recipientUserIDs
              .split(',')
              .map((value) => value.trim())
              .filter(Boolean)
          : [],
      recipient_team_id: recipientType === 'team' ? recipientTeamID.trim() || null : null,
      recipient_role: recipientType === 'project_role' ? recipientRole.trim() : ''
    };
  }

  function resetForm() {
    enabled = true;
    name = '';
    eventKey = 'project.phase.changed';
    projectID = '';
    resourceType = 'project';
    resourceID = '';
    recipientType = 'project_users';
    recipientUserIDs = '';
    recipientTeamID = '';
    recipientRole = '';
  }

  function handleEventKeyChange(value: string) {
    eventKey = value;
    const event = eventDefinitions.find((item) => item.id === value);
    if (event) {
      resourceType = event.resourceType;
      resourceID = '';
    }
  }

  function handleResourceTypeChange(value: string) {
    resourceType = value as NotificationResourceType;
    resourceID = '';
  }

  async function fetchProjects(search: string): Promise<Project[]> {
    const result = await listProjects({ search, limit: 20 });
    return result.items || [];
  }

  async function fetchTeams(search: string): Promise<Team[]> {
    const result = await listTeams({ page: 1, limit: 20, search });
    return result.items || [];
  }

  async function fetchResources(search: string): Promise<SelectOption[]> {
    const normalizedSearch = search.trim().toLowerCase();
    if (resourceType === 'project') {
      const projects = await fetchProjects(search);
      return projects.map(projectOption);
    }
    if (!projectID || resourceType === '' || resourceType === 'project_user') {
      return [];
    }
    switch (resourceType) {
      case 'control_cabinet':
        return filterResourceOptions(await fetchProjectControlCabinets(), normalizedSearch);
      case 'sps_controller':
        return filterResourceOptions(await fetchProjectSPSControllers(), normalizedSearch);
      case 'field_device':
        return filterResourceOptions(await fetchProjectFieldDevices(), normalizedSearch);
      case 'object_data':
        return (
          await projectRepository.listObjectData(projectID, { page: 1, limit: 20, search })
        ).items.map(objectDataOption);
      default:
        return [];
    }
  }

  async function fetchResourceById(id: string): Promise<SelectOption | null> {
    if (!id) return null;
    switch (resourceType) {
      case 'project':
        return projectOption(await getProject(id));
      case 'control_cabinet':
        return controlCabinetOption(await getControlCabinet(id));
      case 'sps_controller':
        return spsControllerOption(await getSPSController(id));
      case 'field_device':
        return fieldDeviceOption(await getFieldDevice(id));
      case 'object_data':
        return objectDataOption(await getObjectData(id));
      default:
        return null;
    }
  }

  async function fetchProjectControlCabinets(): Promise<SelectOption[]> {
    const links = await projectRepository.listControlCabinets(projectID, { page: 1, limit: 100 });
    const ids = links.items.map((link) => link.control_cabinet_id);
    if (ids.length === 0) return [];
    const response = await getControlCabinets(ids);
    return response.items.map(controlCabinetOption);
  }

  async function fetchProjectSPSControllers(): Promise<SelectOption[]> {
    const links = await projectRepository.listSPSControllers(projectID, { page: 1, limit: 100 });
    const ids = links.items.map((link) => link.sps_controller_id);
    if (ids.length === 0) return [];
    const response = await getSPSControllers(ids);
    return response.items.map(spsControllerOption);
  }

  async function fetchProjectFieldDevices(): Promise<SelectOption[]> {
    const links = await projectRepository.listFieldDevices(projectID, { page: 1, limit: 50 });
    const devices = await Promise.all(
      links.items.map((link) => getFieldDevice(link.field_device_id).catch(() => null))
    );
    return devices.filter((item): item is FieldDevice => item !== null).map(fieldDeviceOption);
  }

  function filterResourceOptions(items: SelectOption[], search: string): SelectOption[] {
    if (!search) return items;
    return items.filter((item) => item.label.toLowerCase().includes(search));
  }

  function projectOption(project: Project): SelectOption {
    return { id: project.id, label: project.name || project.id };
  }

  function controlCabinetOption(item: ControlCabinet): SelectOption {
    return { id: item.id, label: item.control_cabinet_nr || item.id };
  }

  function spsControllerOption(item: SPSController): SelectOption {
    const parts = [item.ga_device, item.device_name, item.ip_address].filter(Boolean);
    return { id: item.id, label: parts.length ? parts.join(' · ') : item.id };
  }

  function fieldDeviceOption(item: FieldDevice): SelectOption {
    const parts = [item.bmk, item.text_fix, item.description, item.apparat_nr].filter(Boolean);
    return { id: item.id, label: parts.length ? parts.join(' · ') : item.id };
  }

  function objectDataOption(item: ObjectData): SelectOption {
    const parts = [item.description, item.version].filter(Boolean);
    return { id: item.id, label: parts.length ? parts.join(' · ') : item.id };
  }

  function displayRole(role: string): string {
    return roleOptions.find((option) => option.id === role)?.label || role;
  }

  async function createRule() {
    isSubmitting = true;
    error = null;
    try {
      await notificationRuleRepository.create(buildPayload());
      resetForm();
      await loadRules();
    } catch (err) {
      error = getErrorMessage(err);
    } finally {
      isSubmitting = false;
    }
  }

  async function deleteRule(rule: NotificationRule) {
    isSubmitting = true;
    error = null;
    try {
      await notificationRuleRepository.delete(rule.id);
      await loadRules();
    } catch (err) {
      error = getErrorMessage(err);
    } finally {
      isSubmitting = false;
    }
  }

  onMount(() => {
    loadRules();
    loadRoles();
  });
</script>

<Card.Root>
  <Card.Header class="gap-2">
    <div class="flex items-start justify-between gap-3">
      <div>
        <Card.Title>{$t('notifications.rules.title')}</Card.Title>
        <Card.Description>{$t('notifications.rules.description')}</Card.Description>
      </div>
      <Button
        variant="outline"
        size="icon-sm"
        onclick={loadRules}
        disabled={isLoading || isSubmitting}
      >
        <RefreshCwIcon class={`size-4${isLoading ? ' animate-spin' : ''}`} />
      </Button>
    </div>
  </Card.Header>

  <Card.Content class="space-y-5">
    {#if error}
      <div
        class="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive"
      >
        {error}
      </div>
    {/if}

    <div class="grid gap-3 rounded-md border p-3 lg:grid-cols-2">
      <div class="space-y-2">
        <Label for="notification_rule_name">{$t('notifications.rules.name')}</Label>
        <Input id="notification_rule_name" bind:value={name} />
      </div>
      <div class="space-y-2">
        <Label for="notification_rule_event">{$t('notifications.rules.event_key')}</Label>
        <StaticCombobox
          id="notification_rule_event"
          items={eventOptions}
          bind:value={eventKey}
          labelKey="label"
          width="w-full"
          searchPlaceholder={$t('notifications.rules.event_search')}
          emptyText={$t('notifications.rules.no_events')}
          onValueChange={handleEventKeyChange}
        />
      </div>
      <div class="space-y-2">
        <Label for="notification_rule_project">{$t('notifications.rules.project_id')}</Label>
        <AsyncCombobox
          id="notification_rule_project"
          bind:value={projectID}
          fetcher={fetchProjects}
          fetchById={getProject}
          labelKey="name"
          clearable
          clearText={$t('notifications.rules.clear_selection')}
          placeholder={$t('notifications.rules.project_placeholder')}
          searchPlaceholder={$t('notifications.rules.project_search')}
          emptyText={$t('notifications.rules.no_projects')}
          width="w-full"
          popupWidth="w-[min(32rem,calc(100vw-2rem))]"
        />
      </div>
      <div class="space-y-2">
        <Label for="notification_rule_resource_type"
          >{$t('notifications.rules.resource_type')}</Label
        >
        <StaticCombobox
          id="notification_rule_resource_type"
          items={resourceTypeOptions}
          bind:value={resourceType}
          labelKey="label"
          width="w-full"
          clearable
          clearText={$t('notifications.rules.clear_selection')}
          searchPlaceholder={$t('notifications.rules.resource_type_search')}
          emptyText={$t('notifications.rules.no_resource_types')}
          onValueChange={handleResourceTypeChange}
        />
      </div>
      <div class="space-y-2">
        <Label for="notification_rule_resource">{$t('notifications.rules.resource_id')}</Label>
        <AsyncCombobox
          id="notification_rule_resource"
          bind:value={resourceID}
          fetcher={fetchResources}
          fetchById={fetchResourceById}
          labelKey="label"
          disabled={resourceDisabled}
          clearable
          clearText={$t('notifications.rules.clear_selection')}
          placeholder={$t(
            resourceType === 'project_user'
              ? 'notifications.rules.resource_not_available'
              : resourceDisabled
                ? 'notifications.rules.resource_requires_project'
                : 'notifications.rules.resource_placeholder'
          )}
          searchPlaceholder={$t('notifications.rules.resource_search')}
          emptyText={$t('notifications.rules.no_resources')}
          width="w-full"
          popupWidth="w-[min(32rem,calc(100vw-2rem))]"
          refreshKey={resourceRefreshKey}
        />
      </div>
      <div class="space-y-2">
        <Label for="notification_rule_recipient_type"
          >{$t('notifications.rules.recipient_type')}</Label
        >
        <StaticCombobox
          id="notification_rule_recipient_type"
          items={translatedRecipientTypeOptions}
          bind:value={recipientType}
          labelKey="label"
          width="w-full"
          searchPlaceholder={$t('notifications.rules.recipient_type_search')}
          emptyText={$t('notifications.rules.no_recipient_types')}
        />
      </div>
      <div class="flex items-end gap-2">
        <Switch id="notification_rule_enabled" bind:checked={enabled} />
        <Label for="notification_rule_enabled">{$t('notifications.rules.enabled')}</Label>
      </div>

      {#if recipientType === 'users'}
        <div class="space-y-2 lg:col-span-2">
          <Label for="notification_rule_users">{$t('notifications.rules.user_ids')}</Label>
          <Input id="notification_rule_users" bind:value={recipientUserIDs} />
        </div>
      {:else if recipientType === 'team'}
        <div class="space-y-2 lg:col-span-2">
          <Label for="notification_rule_team">{$t('notifications.rules.team_id')}</Label>
          <AsyncCombobox
            id="notification_rule_team"
            bind:value={recipientTeamID}
            fetcher={fetchTeams}
            fetchById={getTeam}
            labelKey="name"
            clearable
            clearText={$t('notifications.rules.clear_selection')}
            placeholder={$t('notifications.rules.team_placeholder')}
            searchPlaceholder={$t('notifications.rules.team_search')}
            emptyText={$t('notifications.rules.no_teams')}
            width="w-full"
            popupWidth="w-[min(32rem,calc(100vw-2rem))]"
          />
        </div>
      {:else if recipientType === 'project_role'}
        <div class="space-y-2 lg:col-span-2">
          <Label for="notification_rule_role">{$t('notifications.rules.role')}</Label>
          <StaticCombobox
            id="notification_rule_role"
            items={roleOptions}
            bind:value={recipientRole}
            labelKey="label"
            width="w-full"
            searchPlaceholder={$t('notifications.rules.role_search')}
            emptyText={$t('notifications.rules.no_roles')}
          />
          <p class="text-xs text-muted-foreground">{$t('notifications.rules.role_hint')}</p>
        </div>
      {/if}

      <div class="lg:col-span-2">
        <Button onclick={createRule} disabled={isSubmitting || !name.trim() || !eventKey.trim()}>
          <PlusIcon class="size-4" />
          {$t('notifications.rules.create')}
        </Button>
      </div>
    </div>

    <div class="space-y-2">
      {#if isLoading && rules.length === 0}
        <div class="py-4 text-sm text-muted-foreground">{$t('common.loading')}</div>
      {:else if rules.length === 0}
        <div class="py-4 text-sm text-muted-foreground">{$t('notifications.rules.empty')}</div>
      {:else}
        {#each rules as rule (rule.id)}
          <div class="flex items-start justify-between gap-3 rounded-md border px-3 py-2">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <span class="font-medium">{rule.name}</span>
                <span class="rounded-full bg-muted px-2 py-0.5 text-xs">{rule.event_key}</span>
              </div>
              <p class="mt-1 text-xs text-muted-foreground">
                {$t(`notifications.rules.recipient_${rule.recipient_type}`)}
                {#if rule.recipient_role}
                  · {displayRole(rule.recipient_role)}
                {/if}
                {#if rule.project_id}
                  · {$t('notifications.rules.project_id')}: {rule.project_id}
                {/if}
                {#if rule.resource_type}
                  · {$t('notifications.rules.resource_type')}: {$t(
                    `notifications.rules.resource_types.${rule.resource_type}`
                  )}
                {/if}
              </p>
            </div>
            <Button
              variant="ghost"
              size="icon-sm"
              onclick={() => deleteRule(rule)}
              disabled={isSubmitting}
            >
              <Trash2Icon class="size-4" />
            </Button>
          </div>
        {/each}
      {/if}
    </div>
  </Card.Content>
</Card.Root>
