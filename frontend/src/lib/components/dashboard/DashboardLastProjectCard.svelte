<script lang="ts">
  import * as Card from '$lib/components/ui/card/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import type { DashboardProject } from '$lib/domain/dashboard/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { projectStatusVariant } from './project-status.js';

  type Props = {
    project?: DashboardProject;
  };

  let { project }: Props = $props();
  const t = createTranslator();

  function formatDate(value: string): string {
    return new Date(value).toLocaleString();
  }
</script>

<Card.Root>
  <Card.Header>
    <Card.Title>{$t('dashboard.last_project_title')}</Card.Title>
    <Card.Description>{$t('dashboard.last_project_desc')}</Card.Description>
  </Card.Header>
  <Card.Content>
    {#if project}
      <div class="space-y-3">
        <a href="/projects/{project.id}" class="text-base font-semibold hover:underline">
          {project.name}
        </a>
        <div class="flex flex-wrap gap-2">
          <Badge variant={projectStatusVariant(project.status)}
            >{$t(`messages.${project.status}`)}</Badge
          >
          <Badge variant="outline">{$t('dashboard.phase_label')}: {project.phase}</Badge>
        </div>
        <p class="text-sm text-muted-foreground">
          {$t('dashboard.updated_label')}: {formatDate(project.updated_at)}
        </p>
        <a class="text-sm text-primary hover:underline" href="/projects/{project.id}">
          {$t('dashboard.open_project')}
        </a>
      </div>
    {:else}
      <p class="text-sm text-muted-foreground">{$t('dashboard.no_project')}</p>
    {/if}
  </Card.Content>
</Card.Root>
