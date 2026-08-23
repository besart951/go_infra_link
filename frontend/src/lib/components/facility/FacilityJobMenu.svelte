<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { facilityJobState } from '$lib/state/facilityJobState.svelte.js';
  import BanIcon from '@lucide/svelte/icons/ban';
  import DownloadIcon from '@lucide/svelte/icons/download';
  import ListChecksIcon from '@lucide/svelte/icons/list-checks';
  import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';

  const t = createTranslator();

  const stageKey: Record<string, string> = {
    queued: 'facility.copy_progress.queued',
    preparing: 'facility.copy_progress.preparing',
    snapshotting: 'facility.copy_progress.preparing',
    generating: 'facility.copy_progress.copying_field_devices',
    packaging: 'facility.copy_progress.finalizing',
    copying_root: 'facility.copy_progress.copying_root',
    copying_controllers: 'facility.copy_progress.copying_controllers',
    copying_system_types: 'facility.copy_progress.copying_system_types',
    copying_field_devices: 'facility.copy_progress.copying_field_devices',
    finalizing: 'facility.copy_progress.finalizing',
    completed: 'facility.copy_progress.completed',
    failed: 'facility.copy_progress.failed',
    cancelled: 'facility.copy_progress.cancelled'
  };

  const jobs = $derived(
    facilityJobState.jobs.filter(
      (job) =>
        job.status === 'queued' ||
        job.status === 'running' ||
        job.status === 'cancelled' ||
        job.status === 'failed' ||
        (job.type === 'export' && job.status === 'completed')
    )
  );
  const activeCount = $derived(facilityJobState.activeJobs.length);

  function stageLabel(stage: string): string {
    return $t(stageKey[stage] ?? 'facility.copy_progress.preparing');
  }

  function typeLabel(type: string): string {
    return $t(`facility.jobs.${type}`);
  }
</script>

<Popover.Root>
  <Popover.Trigger>
    {#snippet child({ props })}
      <Button
        {...props}
        variant="ghost"
        size="icon"
        class="relative"
        aria-label={$t('facility.jobs.open')}
      >
        <ListChecksIcon class="size-5" />
        {#if activeCount > 0}
          <span
            class="absolute -top-0.5 -right-0.5 flex min-w-4 items-center justify-center rounded-full bg-primary px-1 text-[10px] leading-4 font-semibold text-primary-foreground"
            >{activeCount}</span
          >
        {/if}
      </Button>
    {/snippet}
  </Popover.Trigger>

  <Popover.Content class="w-[min(28rem,calc(100vw-2rem))] p-0" align="end">
    <div class="border-b px-3 py-2">
      <h2 class="text-sm font-semibold">{$t('facility.jobs.title')}</h2>
    </div>
    {#if jobs.length === 0}
      <p class="px-4 py-8 text-center text-sm text-muted-foreground">{$t('facility.jobs.empty')}</p>
    {:else}
      <div class="max-h-[min(32rem,calc(100vh-8rem))] overflow-y-auto">
        {#each jobs as job (job.jobId)}
          <section class="border-b p-3 last:border-b-0">
            <div class="flex items-start gap-2">
              {#if job.status === 'queued' || job.status === 'running'}
                <LoaderCircleIcon class="mt-0.5 size-4 shrink-0 animate-spin text-primary" />
              {/if}
              <div class="min-w-0 flex-1">
                <div class="flex items-center justify-between gap-2">
                  <p class="truncate text-sm font-medium">
                    {typeLabel(job.type)} · {stageLabel(job.stage)}
                  </p>
                  <span class="shrink-0 text-sm text-muted-foreground tabular-nums"
                    >{job.progress}%</span
                  >
                </div>
                <div class="mt-2 h-1.5 overflow-hidden rounded-full bg-muted">
                  <div
                    class="h-full rounded-full bg-primary transition-[width] duration-300"
                    style:width={`${job.progress}%`}
                  ></div>
                </div>
                <div class="mt-1 flex flex-wrap gap-x-2 text-xs text-muted-foreground">
                  {#if job.total !== undefined}<span>{job.processed} / {job.total}</span>{/if}
                  <span>{$t('facility.jobs.attempts', { count: job.attempts })}</span>
                  {#if job.failureCount > 0}<span
                      >{$t('facility.jobs.failed_count', { count: job.failureCount })}</span
                    >{/if}
                </div>
                {#if job.error}<p class="mt-1 text-xs text-destructive">{job.error}</p>{/if}
                {#if facilityJobState.connectionInterrupted && (job.status === 'queued' || job.status === 'running')}
                  <p class="mt-1 text-xs text-muted-foreground">
                    {$t('facility.copy_progress.connection_interrupted')}
                  </p>
                {/if}
                <div class="mt-2 flex gap-2">
                  {#if job.result?.download_url}
                    <Button size="sm" variant="outline" href={job.result.download_url}
                      ><DownloadIcon class="size-4" />{$t('facility.jobs.download')}</Button
                    >
                  {/if}
                  {#if job.status === 'failed' && job.retryable}
                    <Button
                      size="sm"
                      variant="outline"
                      onclick={() => facilityJobState.retry(job.jobId)}
                      ><RefreshCwIcon class="size-4" />{$t('facility.jobs.retry')}</Button
                    >
                  {/if}
                  {#if job.status === 'queued' || job.status === 'running'}
                    <Button
                      size="sm"
                      variant="outline"
                      onclick={() => facilityJobState.cancel(job.jobId)}
                      ><BanIcon class="size-4" />{$t('facility.jobs.cancel')}</Button
                    >
                  {/if}
                </div>
              </div>
            </div>
          </section>
        {/each}
      </div>
    {/if}
  </Popover.Content>
</Popover.Root>
