<script lang="ts">
  import type { ActivityScope } from '$lib/activity/contract.js';
  import { t as translate } from '$lib/i18n/index.js';
  import PermissionGate from '$lib/components/auth/PermissionGate.svelte';
  import HistoryTimelineDialog from '$lib/components/history/HistoryTimelineDialog.svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import HistoryIcon from '@lucide/svelte/icons/history';

  interface Props {
    title: string;
    scope: ActivityScope;
    projectId?: string;
    controlCabinetId?: string;
    compact?: boolean;
    onRestored?: () => void | Promise<void>;
  }

  let { title, scope, projectId, controlCabinetId, compact = true, onRestored }: Props = $props();

  let open = $state(false);
</script>

<PermissionGate permission="timeline.read">
  <Tooltip.Root>
    <Tooltip.Trigger>
      {#snippet child({ props })}
        <Button
          {...props}
          variant="outline"
          size={compact ? 'icon' : 'sm'}
          aria-label={translate('history.activity.open')}
          onclick={() => (open = true)}
        >
          <HistoryIcon class="size-4" />
          {#if !compact}{translate('history.open')}{/if}
        </Button>
      {/snippet}
    </Tooltip.Trigger>
    <Tooltip.Content>{translate('history.activity.open')}</Tooltip.Content>
  </Tooltip.Root>

  <HistoryTimelineDialog
    bind:open
    {title}
    scopeType={scope.scopeType}
    scopeId={scope.scopeId}
    entityTable={scope.entity?.table}
    entityId={scope.entity?.id}
    fields={scope.fields}
    {projectId}
    {controlCabinetId}
    {onRestored}
  />
</PermissionGate>
