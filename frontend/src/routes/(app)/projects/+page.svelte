<script lang="ts">
  import ModuleCardGrid, {
    type ModuleCardItem
  } from '$lib/components/navigation/ModuleCardGrid.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import CalendarCheckIcon from '@lucide/svelte/icons/calendar-check';
  import FolderKanbanIcon from '@lucide/svelte/icons/folder-kanban';

  const t = createTranslator();

  const projectCards = $derived.by<ModuleCardItem[]>(() =>
    [
      {
        title: $t('hub.projects.list_title'),
        description: $t('hub.projects.list_desc'),
        href: '/projects/list',
        icon: FolderKanbanIcon,
        tone: 'project',
        hasAccess: true
      },
      {
        title: $t('phase.phases'),
        description: $t('hub.projects.phases_desc'),
        href: '/projects/phases',
        icon: CalendarCheckIcon,
        tone: 'project',
        hasAccess: canPerform('read', 'phase')
      }
    ].filter((item) => item.hasAccess)
  );
</script>

<svelte:head>
  <title>{$t('navigation.projects')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <EntityListHeader
    title={$t('hub.projects.title')}
    description={$t('hub.projects.description')}
    backHref="/"
    backLabel={$t('hub.back_to_dashboard')}
  />

  <ModuleCardGrid items={projectCards} emptyMessage={$t('hub.no_access')} />
</div>
