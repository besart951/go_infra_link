<script lang="ts">
  import { Badge } from '$lib/components/ui/badge/index.js';
  import Check from '@lucide/svelte/icons/check';
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';
  import PencilLine from '@lucide/svelte/icons/pencil-line';
  import Radio from '@lucide/svelte/icons/radio';
  import { cn } from '$lib/utils.js';

  export type DetailSaveStatus = 'saved' | 'editing' | 'saving' | 'updated' | 'conflict';

  let { status = 'saved' }: { status?: DetailSaveStatus } = $props();

  const state = $derived.by(() => {
    switch (status) {
      case 'editing':
        return {
          label: 'Lokale Bearbeitung',
          icon: PencilLine,
          className: 'border-warning-border bg-warning-muted text-warning-muted-foreground'
        };
      case 'saving':
        return {
          label: 'Speichert',
          icon: LoaderCircle,
          className: 'border-info-border bg-info-muted text-info-muted-foreground animate-pulse'
        };
      case 'updated':
        return {
          label: 'Live aktualisiert',
          icon: Radio,
          className:
            'border-success-border bg-success-muted text-success-muted-foreground animate-pulse'
        };
      case 'conflict':
        return {
          label: 'Konflikt prüfen',
          icon: CircleAlert,
          className: 'border-destructive/60 bg-destructive/20 text-destructive'
        };
      default:
        return {
          label: 'Gespeichert',
          icon: Check,
          className: 'border-border bg-muted/50 text-muted-foreground'
        };
    }
  });
  const StatusIcon = $derived(state.icon);
</script>

<Badge variant="outline" class={cn('gap-1.5 font-medium', state.className)} aria-live="polite">
  <StatusIcon class={cn('size-3.5', status === 'saving' && 'animate-spin')} />
  {state.label}
</Badge>
