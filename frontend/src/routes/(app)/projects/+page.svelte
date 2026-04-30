<script lang="ts">
  import ModuleCardGrid, {
    type ModuleCardItem
  } from '$lib/components/navigation/ModuleCardGrid.svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
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
  <header class="flex flex-col gap-4 border-b pb-5 sm:flex-row sm:items-end sm:justify-between">
    <div class="min-w-0 space-y-1">
      <h1 class="text-2xl font-semibold tracking-tight sm:text-3xl">
        {$t('hub.projects.title')}
      </h1>
      <p class="max-w-3xl text-sm leading-6 text-muted-foreground">
        {$t('hub.projects.description')}
      </p>
    </div>
    <Button variant="outline" href="/" class="w-full sm:w-auto">
      <ArrowLeftIcon class="size-4" />
      {$t('hub.back_to_dashboard')}
    </Button>
  </header>

  <ModuleCardGrid items={projectCards} emptyMessage={$t('hub.no_access')} />
</div>
