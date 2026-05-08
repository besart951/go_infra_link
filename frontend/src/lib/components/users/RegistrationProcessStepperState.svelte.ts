import type {
  RegistrationProcess,
  RegistrationProcessStep
} from '$lib/infrastructure/api/userRepository.js';

type ProcessReader = () => RegistrationProcess | null | undefined;

export type RegistrationStepVisualStatus = 'complete' | 'active' | 'pending' | 'failed' | 'skipped';

export interface RegistrationStepView {
  key: string;
  label: string;
  shortLabel: string;
  status: RegistrationProcessStep['status'];
  visualStatus: RegistrationStepVisualStatus;
  isActive: boolean;
  timestamp?: string | null;
}

export class RegistrationProcessStepperState {
  constructor(private readonly readProcess: ProcessReader) {}

  readonly process = $derived.by(() => this.readProcess() ?? null);

  readonly stepViews = $derived.by<RegistrationStepView[]>(() => {
    const steps = this.process?.steps ?? [];
    const activeIndex = activeStepIndex(steps);

    return steps.map((step, index) => ({
      key: step.key,
      label: step.label,
      shortLabel: shortLabel(step),
      status: step.status,
      visualStatus: visualStatus(step.status),
      isActive: index === activeIndex,
      timestamp: step.timestamp
    }));
  });

  readonly totalSteps = $derived(this.stepViews.length);
  readonly hasProcess = $derived(this.process !== null && this.totalSteps > 0);

  readonly activeIndex = $derived.by(() => {
    if (this.stepViews.length === 0) return 0;
    const explicit = this.stepViews.findIndex((step) => step.isActive);
    return explicit >= 0 ? explicit : 0;
  });

  readonly currentStepNumber = $derived(this.hasProcess ? this.activeIndex + 1 : 0);
  readonly currentStep = $derived(this.stepViews[this.activeIndex] ?? null);
  readonly currentLabel = $derived(this.currentStep?.label ?? '');
  readonly currentStatus = $derived(this.currentStep?.status ?? 'pending');
  readonly isProblemState = $derived(
    this.currentStatus === 'failed' ||
      this.currentStatus === 'blocked' ||
      this.process?.status === 'email_failed' ||
      this.process?.status === 'expired'
  );
}

function activeStepIndex(steps: RegistrationProcessStep[]): number {
  const explicit = steps.findIndex(
    (step) => step.status === 'current' || step.status === 'failed' || step.status === 'blocked'
  );
  if (explicit >= 0) return explicit;

  for (let index = steps.length - 1; index >= 0; index -= 1) {
    if (steps[index].status === 'completed') return index;
  }

  return 0;
}

function visualStatus(status: RegistrationProcessStep['status']): RegistrationStepVisualStatus {
  switch (status) {
    case 'completed':
      return 'complete';
    case 'current':
      return 'active';
    case 'failed':
    case 'blocked':
      return 'failed';
    case 'skipped':
      return 'skipped';
    default:
      return 'pending';
  }
}

function shortLabel(step: RegistrationProcessStep): string {
  switch (step.key) {
    case 'email_sent':
      return 'E-Mail';
    case 'first_login':
      return 'Login';
    default:
      return step.label;
  }
}
