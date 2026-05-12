<script lang="ts">
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { cn } from '$lib/utils.js';
  import type { ProjectStatus } from '$lib/domain/project/index.js';

  type Props = {
    status: ProjectStatus;
    label?: string;
    class?: string;
  };

  let { status, label, class: className }: Props = $props();

  const statusClasses: Record<ProjectStatus, string> = {
    planned: 'border-border bg-muted text-muted-foreground',
    ongoing: 'border-info-border bg-info-muted text-info-muted-foreground',
    completed: 'border-success-border bg-success-muted text-success-muted-foreground'
  };
</script>

<Badge
  variant="outline"
  class={cn('capitalize shadow-none hover:bg-current/0', statusClasses[status], className)}
>
  {label ?? status}
</Badge>
