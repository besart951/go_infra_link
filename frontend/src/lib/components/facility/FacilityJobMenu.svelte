<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import type { FacilityJob } from '$lib/domain/facility/facility-job.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { facilityJobState } from '$lib/state/facilityJobState.svelte.js';
  import BanIcon from '@lucide/svelte/icons/ban';
  import CheckCircle2Icon from '@lucide/svelte/icons/circle-check-big';
  import CircleXIcon from '@lucide/svelte/icons/circle-x';
  import DownloadIcon from '@lucide/svelte/icons/download';
  import ListChecksIcon from '@lucide/svelte/icons/list-checks';
  import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
  import XIcon from '@lucide/svelte/icons/x';

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

  function isActive(job: FacilityJob): boolean {
    return job.status === 'queued' || job.status === 'running';
  }

  function isTerminal(job: FacilityJob): boolean {
    return job.status === 'completed' || job.status === 'failed' || job.status === 'cancelled';
  }

  function downloadURL(job: FacilityJob): string {
    return job.result?.download_url ?? '';
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
        title={$t('facility.jobs.open')}
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

  <Popover.Content class="w-[min(25rem,calc(100vw-2rem))] p-0" align="end">
    <div class="flex items-center justify-between border-b px-3 py-2.5">
      <h2 class="text-sm font-semibold">{$t('facility.jobs.title')}</h2>
      {#if activeCount > 0}
        <span class="text-xs text-muted-foreground tabular-nums">{activeCount}</span>
      {/if}
    </div>
    {#if jobs.length === 0}
      <p class="px-4 py-8 text-center text-sm text-muted-foreground">{$t('facility.jobs.empty')}</p>
    {:else}
      <div class="max-h-[min(32rem,calc(100vh-8rem))] divide-y overflow-y-auto">
        {#each jobs as job (job.jobId)}
          <section class="px-3 py-2.5">
            <div class="flex items-start gap-2.5">
              {#if isActive(job)}
                <LoaderCircleIcon class="mt-0.5 size-4 shrink-0 animate-spin text-primary" />
              {:else if job.status === 'failed'}
                <CircleXIcon class="mt-0.5 size-4 shrink-0 text-destructive" />
              {:else if job.status === 'cancelled'}
                <BanIcon class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
              {:else}
                <CheckCircle2Icon class="mt-0.5 size-4 shrink-0 text-primary" />
              {/if}

              <div class="min-w-0 flex-1">
                <div class="flex items-start justify-between gap-2">
                  <div class="min-w-0">
                    <p class="truncate text-sm font-medium">{typeLabel(job.type)}</p>
                    <p class="truncate text-xs text-muted-foreground">{stageLabel(job.stage)}</p>
                  </div>
                  <div class="flex shrink-0 items-center gap-0.5">
                    {#if job.result?.download_url}
                      <Button
                        size="icon-sm"
                        variant="ghost"
                        href={downloadURL(job)}
                        aria-label={$t('facility.jobs.download')}
                        title={$t('facility.jobs.download')}><DownloadIcon class="size-4" /></Button
                      >
                    {/if}
                    {#if job.status === 'failed' && job.retryable}
                      <Button
                        size="icon-sm"
                        variant="ghost"
                        aria-label={$t('facility.jobs.retry')}
                        title={$t('facility.jobs.retry')}
                        onclick={() => facilityJobState.retry(job.jobId)}
                        ><RefreshCwIcon class="size-4" /></Button
                      >
                    {/if}
                    {#if isActive(job)}
                      <Button
                        size="icon-sm"
                        variant="ghost"
                        aria-label={$t('facility.jobs.cancel')}
                        title={$t('facility.jobs.cancel')}
                        onclick={() => facilityJobState.cancel(job.jobId)}
                        ><BanIcon class="size-4" /></Button
                      >
                    {/if}
                    {#if isTerminal(job)}
                      <Button
                        size="icon-sm"
                        variant="ghost"
                        class="text-muted-foreground hover:text-foreground"
                        aria-label={$t('facility.jobs.remove')}
                        title={$t('facility.jobs.remove')}
                        onclick={() => facilityJobState.remove(job.jobId)}
                        ><XIcon class="size-4" /></Button
                      >
                    {/if}
                  </div>
                </div>

                {#if isActive(job)}
                  <div class="mt-2 h-1 overflow-hidden rounded-full bg-muted">
                    <div
                      class="h-full rounded-full bg-primary transition-[width] duration-300"
                      style:width={`${job.progress}%`}
                    ></div>
                  </div>
                {/if}

                <div class="mt-1.5 flex flex-wrap gap-x-2 text-xs text-muted-foreground">
                  {#if isActive(job)}<span class="tabular-nums">{job.progress}%</span>{/if}
                  {#if job.total !== undefined}<span>{job.processed} / {job.total}</span>{/if}
                  <span>{$t('facility.jobs.attempts', { count: job.attempts })}</span>
                  {#if job.failureCount > 0}<span
                      >{$t('facility.jobs.failed_count', { count: job.failureCount })}</span
                    >{/if}
                </div>
                {#if job.status === 'failed' && job.error}
                  <p class="mt-1.5 text-xs wrap-break-word text-destructive">{job.error}</p>
                {/if}
                {#if facilityJobState.connectionInterrupted && isActive(job)}
                  <p class="mt-1.5 text-xs text-muted-foreground">
                    {$t('facility.copy_progress.connection_interrupted')}
                  </p>
                {/if}
              </div>
            </div>
          </section>
        {/each}
      </div>
    {/if}
  </Popover.Content>
</Popover.Root>
