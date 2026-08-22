import type { FacilityJob } from '$lib/domain/facility/facility-job.js';

export type FacilityJobInput = Pick<
  FacilityJob,
  'jobId' | 'kind' | 'status' | 'progress' | 'stage'
> &
  Partial<Omit<FacilityJob, 'jobId' | 'kind' | 'status' | 'progress' | 'stage'>>;

export function reconcileFacilityJob(jobs: FacilityJob[], input: FacilityJobInput): FacilityJob[] {
  const normalized = normalizeFacilityJob(input);
  return [normalized, ...jobs.filter((job) => job.jobId !== normalized.jobId)];
}

export function normalizeFacilityJob(job: FacilityJobInput): FacilityJob {
  return {
    ...job,
    type: job.type ?? 'copy',
    class: job.class ?? (job.type === 'export' ? 'export' : 'mutation'),
    attempts: job.attempts ?? 0,
    processed: job.processed ?? 0,
    successCount: job.successCount ?? 0,
    failureCount: job.failureCount ?? 0,
    retryable: job.retryable ?? false
  };
}

export function sameFacilityJobProgress(current: FacilityJob, next: FacilityJob): boolean {
  return progressFields.every((field) => current[field] === next[field]);
}

const progressFields = [
  'updatedAt',
  'kind',
  'type',
  'class',
  'status',
  'progress',
  'stage',
  'processed',
  'total',
  'successCount',
  'failureCount',
  'error'
] as const satisfies readonly (keyof FacilityJob)[];
