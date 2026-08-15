<script lang="ts">
  import { ActivityTimelineController } from '$lib/activity/activityTimelineController.svelte.js';
  import { invalidateActivityCache } from '$lib/activity/activityCache.js';
  import { subscribeToProjectActivity } from '$lib/activity/activityLiveUpdates.js';
  import {
    globalActivityDataSource,
    projectActivityDataSource
  } from '$lib/activity/historyActivityDataSource.js';
  import ActivityFeed from '$lib/components/activity/ActivityFeed.svelte';
  import ActivityLiveNotice from '$lib/components/activity/ActivityLiveNotice.svelte';
  import { addToast } from '$lib/components/toast.svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Dialog from '$lib/components/ui/dialog/index.js';
  import { getErrorMessage } from '$lib/api/client.js';
  import type { ChangeEvent, HistoryTimelineParams } from '$lib/domain/history.js';
  import { t as translate } from '$lib/i18n/index.js';
  import { historyRepository } from '$lib/infrastructure/api/historyRepository.js';
  import { can } from '$lib/utils/permissions.js';
  import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
  import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';

  interface Props {
    open?: boolean;
    title?: string;
    scopeType?: string;
    scopeId?: string;
    entityTable?: string;
    entityId?: string;
    fields?: string[];
    projectId?: string;
    liveVersion?: number;
    controlCabinetId?: string;
    onRestored?: () => void | Promise<void>;
  }

  let {
    open = $bindable(false),
    title,
    scopeType,
    scopeId,
    entityTable,
    entityId,
    fields = [],
    projectId,
    liveVersion = 0,
    controlCabinetId,
    onRestored
  }: Props = $props();

  const limit = 25;
  let timeline = $state<ActivityTimelineController | null>(null);
  let restoringEventId = $state<string | null>(null);
  let observedLiveVersion = 0;
  const canReadTimeline = $derived(can('timeline.read'));
  const canRestoreTimeline = $derived(can('timeline.restore'));
  const params = $derived<HistoryTimelineParams>({
    scopeType,
    scopeId,
    entityTable,
    entityId,
    fields,
    limit
  });

  $effect(() => {
    if (!open || !canReadTimeline) {
      if (!canReadTimeline) open = false;
      return;
    }

    const controller = new ActivityTimelineController(
      projectId ? projectActivityDataSource(projectId) : globalActivityDataSource
    );
    timeline = controller;
    void controller.load(params);
    return () => {
      controller.dispose();
      if (timeline === controller) timeline = null;
    };
  });

  $effect(() => {
    if (!projectId || !open) return;

    return subscribeToProjectActivity(projectId, () => {
      timeline?.markLiveChange();
    });
  });

  $effect(() => {
    const controller = timeline;
    const namespace = projectId ? `history:project:${projectId}` : 'history:global';
    if (liveVersion > observedLiveVersion) {
      observedLiveVersion = liveVersion;
      invalidateActivityCache(namespace);
      controller?.markLiveChange();
      return;
    }
    if (!controller) {
      observedLiveVersion = liveVersion;
      return;
    }
  });

  async function reload(force = true): Promise<void> {
    await timeline?.load(params, { force });
  }

  async function restoreEvent(event: ChangeEvent): Promise<void> {
    if (!canRestoreTimeline) return;
    restoringEventId = event.id;
    try {
      const result =
        projectId && controlCabinetId
          ? await historyRepository.restoreProjectControlCabinet(
              projectId,
              controlCabinetId,
              event.id
            )
          : await historyRepository.restoreEvent(
              event.id,
              event.action === 'delete' ? 'before' : 'after'
            );
      const changedCount = result.restored_count + result.deleted_count;
      addToast(translate('history.timeline.undo_success', { count: changedCount }), 'success');
      await onRestored?.();
      await reload();
    } catch (error) {
      addToast(getErrorMessage(error), 'error');
    } finally {
      restoringEventId = null;
    }
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-h-[88vh] overflow-hidden p-0 sm:max-w-4xl" showCloseButton={false}>
    <Dialog.Header class="border-b px-6 py-5 pr-12">
      <Dialog.Title>{title ?? translate('history.open')}</Dialog.Title>
      <Dialog.Description>
        {translate('history.description')}
      </Dialog.Description>
    </Dialog.Header>

    <div class="max-h-[calc(88vh-10rem)] space-y-4 overflow-y-auto px-6 py-5">
      <ActivityLiveNotice
        count={timeline?.pendingLiveChanges ?? 0}
        loading={timeline?.loading ?? false}
        onShow={() => void reload()}
      />
      <ActivityFeed
        events={timeline?.events ?? []}
        loading={timeline?.loading ?? false}
        error={timeline?.error}
        emptyText={translate('history.timeline.empty')}
        canRestore={canRestoreTimeline}
        {restoringEventId}
        onRestore={restoreEvent}
      />

      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="text-sm text-muted-foreground">
          {timeline
            ? translate('history.activity.loaded_count', {
                loaded: timeline.events.length,
                total: timeline.total
              })
            : ''}
        </div>
        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            size="icon-sm"
            disabled={!timeline || timeline.loading || timeline.loadingMore}
            onclick={() => void reload()}
            aria-label={translate('common.refresh')}
          >
            <RefreshCwIcon class="size-4" />
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={!timeline || timeline.loading || timeline.loadingMore || timeline.page <= 1}
            onclick={() =>
              void timeline?.load({ ...params, page: timeline.page - 2 }, { force: true })}
          >
            <ChevronLeftIcon class="size-4" />
            {translate('common.back')}
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={!timeline || timeline.loading || timeline.loadingMore || !timeline.hasMore}
            onclick={() => void timeline?.load(params, { append: true })}
          >
            {translate('common.next')}
            <ChevronRightIcon class="size-4" />
          </Button>
        </div>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>
