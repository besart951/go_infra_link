import type { components } from '$lib/api/generated/schema.js';

type GeneratedCopyJob =
  components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CopyJobResponse'];

export type CopyJobKind = NonNullable<GeneratedCopyJob['kind']>;
export type CopyJobStatus = NonNullable<GeneratedCopyJob['status']>;

export interface CopyJob {
  jobId: string;
  kind: CopyJobKind;
  status: CopyJobStatus;
  progress: number;
  stage: string;
  error?: string;
}

export function toCopyJob(response: GeneratedCopyJob): CopyJob {
  if (
    !response.job_id ||
    !response.kind ||
    !response.status ||
    response.progress === undefined ||
    !response.stage
  ) {
    throw new Error('Invalid copy job response');
  }

  return {
    jobId: response.job_id,
    kind: response.kind,
    status: toCopyJobStatus(response.status),
    progress: Math.min(100, Math.max(0, response.progress)),
    stage: response.stage,
    ...(response.error ? { error: response.error } : {})
  };
}

function toCopyJobStatus(value: string): CopyJobStatus {
  if (value === 'queued' || value === 'running' || value === 'completed' || value === 'failed') {
    return value;
  }
  throw new Error(`Unsupported copy job status: ${value}`);
}
