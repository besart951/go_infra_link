<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { facilityJobState } from '$lib/state/facilityJobState.svelte.js';
  import DownloadIcon from '@lucide/svelte/icons/download';
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
    failed: 'facility.copy_progress.failed'
  };

  const visibleJobs = $derived(
    facilityJobState.jobs
      .filter(
        (job) =>
          job.status === 'queued' ||
          job.status === 'running' ||
          (job.type === 'export' && job.status === 'completed') ||
          (job.status === 'failed' && job.retryable)
      )
      .slice(0, 5)
  );

  function stageLabel(stage: string): string {
    return $t(stageKey[stage] ?? 'facility.copy_progress.preparing');
  }
</script>

{#if visibleJobs.length > 0}
  <div
    class="fixed right-4 bottom-4 z-50 flex w-[min(26rem,calc(100vw-2rem))] flex-col gap-2"
    aria-live="polite"
  >
    {#each visibleJobs as job (job.jobId)}
      <div class="rounded-lg border bg-card p-4 shadow-lg">
        <div class="flex items-start gap-3">
          {#if job.status === 'queued' || job.status === 'running'}
            <LoaderCircleIcon class="mt-0.5 size-5 shrink-0 animate-spin text-primary" />
          {/if}
          <div class="min-w-0 flex-1">
            <div class="flex items-center justify-between gap-3">
              <p class="truncate text-sm font-medium">{stageLabel(job.stage)}</p>
              <span class="text-sm text-muted-foreground tabular-nums">{job.progress}%</span>
            </div>
            <div
              class="mt-2 h-2 overflow-hidden rounded-full bg-muted"
              role="progressbar"
              aria-label={$t('facility.copy_progress.aria')}
              aria-valuemin="0"
              aria-valuemax="100"
              aria-valuenow={job.progress}
            >
              <div
                class="h-full rounded-full bg-primary transition-[width] duration-300"
                style:width={`${job.progress}%`}
              ></div>
            </div>
            {#if job.total !== undefined}
              <p class="mt-1 text-xs text-muted-foreground">
                {job.processed} / {job.total}
                {#if job.failureCount > 0}
                  · {job.failureCount} fehlgeschlagen{/if}
              </p>
            {/if}
            {#if job.error}
              <p class="mt-2 text-xs text-destructive">{job.error}</p>
            {/if}
            {#if facilityJobState.connectionInterrupted && (job.status === 'queued' || job.status === 'running')}
              <p class="mt-2 text-xs text-muted-foreground">
                {$t('facility.copy_progress.connection_interrupted')}
              </p>
            {/if}
            <div class="mt-2 flex gap-2">
              {#if job.result?.download_url}
                <Button size="sm" variant="outline" href={job.result.download_url}>
                  <DownloadIcon class="size-4" />
                  Download
                </Button>
              {/if}
              {#if job.status === 'failed' && job.retryable}
                <Button
                  size="sm"
                  variant="outline"
                  onclick={() => facilityJobState.retry(job.jobId)}
                >
                  <RefreshCwIcon class="size-4" />
                  Retry
                </Button>
              {/if}
            </div>
          </div>
        </div>
      </div>
    {/each}
  </div>
{/if}
