<script lang="ts">
  import type { RegistrationProcess } from '$lib/infrastructure/api/userRepository.js';
  import { createTranslator } from '$lib/i18n/translator';
  import { cn } from '$lib/utils.js';
  import { RegistrationProcessStepperState } from './RegistrationProcessStepperState.svelte.js';
  import type { RegistrationStepVisualStatus } from './RegistrationProcessStepperState.svelte.js';

  interface Props {
    process?: RegistrationProcess | null;
  }

  let { process = null }: Props = $props();
  const t = createTranslator();
  const state = new RegistrationProcessStepperState(() => process);

  function gridClass(count: number): string {
    switch (count) {
      case 1:
        return 'grid-cols-1';
      case 2:
        return 'grid-cols-2';
      case 3:
        return 'grid-cols-3';
      case 5:
        return 'grid-cols-5';
      case 6:
        return 'grid-cols-6';
      default:
        return 'grid-cols-4';
    }
  }

  function statusColorClass(status: RegistrationStepVisualStatus): string {
    switch (status) {
      case 'complete':
        return 'bg-success';
      case 'active':
        return 'bg-primary';
      case 'failed':
        return 'bg-destructive';
      case 'skipped':
        return 'bg-muted';
      default:
        return 'bg-muted-foreground/25';
    }
  }

  function segmentClass(status: RegistrationStepVisualStatus, isActive: boolean): string {
    return cn(
      'h-2 min-w-0 rounded-xs ring-1 ring-border/70 shadow-inner',
      statusColorClass(status),
      isActive && 'ring-2 ring-ring/45 ring-offset-1 ring-offset-background'
    );
  }

  function stepClass(isActive: boolean): string {
    return cn(
      'flex min-w-0 items-center justify-center gap-1 text-muted-foreground sm:justify-start',
      isActive && 'font-semibold text-foreground'
    );
  }

  function dotClass(status: RegistrationStepVisualStatus): string {
    return cn('size-1.5 flex-none rounded-full', statusColorClass(status));
  }

  function statusLabel(status: string): string {
    switch (status) {
      case 'completed':
        return $t('user.registration_status_completed');
      case 'current':
        return $t('user.registration_status_current');
      case 'failed':
        return $t('user.registration_status_failed');
      case 'blocked':
        return $t('user.registration_status_blocked');
      case 'skipped':
        return $t('user.registration_status_skipped');
      default:
        return $t('user.registration_status_pending');
    }
  }
</script>

{#if state.hasProcess}
  <div
    class="flex w-56 max-w-full min-w-48 flex-col gap-1 sm:w-full sm:max-w-76 sm:min-w-54 sm:gap-[0.45rem]"
  >
    <div
      class="grid min-w-0 grid-cols-1 items-baseline gap-0.5 sm:grid-cols-[auto_minmax(0,1fr)] sm:gap-2"
    >
      <span
        class={cn(
          'text-xs leading-none font-bold whitespace-nowrap',
          state.isProblemState ? 'text-destructive' : 'text-foreground'
        )}
      >
        {$t('user.registration_step', {
          step: state.currentStepNumber,
          total: state.totalSteps
        })}
      </span>
      <span class="min-w-0 truncate text-[0.72rem] leading-none font-medium text-muted-foreground">
        {state.currentLabel}
      </span>
    </div>

    <div
      class={cn('grid w-full gap-1.5', gridClass(state.totalSteps))}
      role="progressbar"
      aria-valuemin="1"
      aria-valuemax={state.totalSteps}
      aria-valuenow={state.currentStepNumber}
      aria-label={$t('user.registration_progress_aria', {
        step: state.currentStepNumber,
        total: state.totalSteps,
        label: state.currentLabel
      })}
    >
      {#each state.stepViews as step (step.key)}
        <span
          class={segmentClass(step.visualStatus, step.isActive)}
          title={`${step.label}: ${statusLabel(step.status)}`}
        ></span>
      {/each}
    </div>

    <ol class={cn('grid gap-1.5', gridClass(state.totalSteps))}>
      {#each state.stepViews as step (step.key)}
        <li
          class={stepClass(step.isActive)}
          title={`${step.label}: ${statusLabel(step.status)}`}
          aria-current={step.isActive ? 'step' : undefined}
        >
          <span class={dotClass(step.visualStatus)}></span>
          <span class="hidden min-w-0 truncate text-[0.65rem] leading-none sm:inline">
            {step.shortLabel}
          </span>
        </li>
      {/each}
    </ol>
  </div>
{:else}
  <span class="text-sm text-muted-foreground">-</span>
{/if}
