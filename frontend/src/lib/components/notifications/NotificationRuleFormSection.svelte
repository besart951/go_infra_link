<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import StaticCombobox from '$lib/components/ui/combobox/StaticCombobox.svelte';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Switch } from '$lib/components/ui/switch/index.js';
  import type { NotificationRuleRecipientType } from '$lib/domain/notification/index.js';
  import type { Project } from '$lib/domain/project/index.js';
  import type { Team } from '$lib/api/teams.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import PlusIcon from '@lucide/svelte/icons/plus';

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

  interface Props {
    name: string;
    eventKey: string;
    projectID: string;
    resourceType: NotificationResourceType;
    resourceID: string;
    recipientType: NotificationRuleRecipientType;
    recipientUserIDs: string;
    recipientTeamID: string;
    recipientRole: string;
    enabled: boolean;
    isSubmitting: boolean;
    eventOptions: SelectOption[];
    resourceTypeOptions: SelectOption[];
    recipientTypeOptions: SelectOption[];
    roleOptions: SelectOption[];
    resourceDisabled: boolean;
    resourceRefreshKey: string;
    fetchProjects: (search: string) => Promise<Project[]>;
    fetchProjectById: (id: string) => Promise<Project>;
    fetchResources: (search: string) => Promise<SelectOption[]>;
    fetchResourceById: (id: string) => Promise<SelectOption | null>;
    fetchTeams: (search: string) => Promise<Team[]>;
    fetchTeamById: (id: string) => Promise<Team>;
    onEventKeyChange: (value: string) => void;
    onResourceTypeChange: (value: string) => void;
    onCreateRule: () => void | Promise<void>;
  }

  let {
    name = $bindable(),
    eventKey = $bindable(),
    projectID = $bindable(),
    resourceType = $bindable(),
    resourceID = $bindable(),
    recipientType = $bindable(),
    recipientUserIDs = $bindable(),
    recipientTeamID = $bindable(),
    recipientRole = $bindable(),
    enabled = $bindable(),
    isSubmitting,
    eventOptions,
    resourceTypeOptions,
    recipientTypeOptions,
    roleOptions,
    resourceDisabled,
    resourceRefreshKey,
    fetchProjects,
    fetchProjectById,
    fetchResources,
    fetchResourceById,
    fetchTeams,
    fetchTeamById,
    onEventKeyChange,
    onResourceTypeChange,
    onCreateRule
  }: Props = $props();

  const t = createTranslator();
</script>

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
      onValueChange={onEventKeyChange}
    />
  </div>
  <div class="space-y-2">
    <Label for="notification_rule_project">{$t('notifications.rules.project_id')}</Label>
    <AsyncCombobox
      id="notification_rule_project"
      bind:value={projectID}
      fetcher={fetchProjects}
      fetchById={fetchProjectById}
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
    <Label for="notification_rule_resource_type">{$t('notifications.rules.resource_type')}</Label>
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
      onValueChange={onResourceTypeChange}
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
    <Label for="notification_rule_recipient_type">{$t('notifications.rules.recipient_type')}</Label>
    <StaticCombobox
      id="notification_rule_recipient_type"
      items={recipientTypeOptions}
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
        fetchById={fetchTeamById}
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
    <Button onclick={onCreateRule} disabled={isSubmitting || !name.trim() || !eventKey.trim()}>
      <PlusIcon class="size-4" />
      {$t('notifications.rules.create')}
    </Button>
  </div>
</div>
