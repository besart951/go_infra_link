export type FacilityJobKind =
  | 'control_cabinet'
  | 'sps_controller'
  | 'sps_controller_system_type'
  | 'field_device'
  | 'object_data';
export type FacilityJobStatus = 'queued' | 'running' | 'completed' | 'failed';
export type FacilityJobType = 'copy' | 'export' | 'bulk' | 'delete' | 'restore';
export type FacilityJobClass = 'mutation' | 'export';

export interface FacilityJobResult {
  download_url?: string;
  file_name?: string;
  output_type?: 'excel' | 'zip';
  content_type?: string;
  size?: number;
  expires_at?: string;
  resource_id?: string;
  [key: string]: unknown;
}

export interface FacilityJob {
  jobId: string;
  kind: FacilityJobKind;
  type: FacilityJobType;
  class: FacilityJobClass;
  status: FacilityJobStatus;
  progress: number;
  stage: string;
  error?: string;
  attempts: number;
  processed: number;
  total?: number;
  successCount: number;
  failureCount: number;
  retryable: boolean;
  result?: FacilityJobResult;
  createdAt?: string;
  updatedAt?: string;
  completedAt?: string;
}

interface FacilityJobWire {
  job_id?: string;
  kind?: string;
  type?: string;
  class?: string;
  status?: string;
  progress?: number;
  stage?: string;
  error?: string;
  attempts?: number;
  processed?: number;
  total?: number;
  success_count?: number;
  failure_count?: number;
  retryable?: boolean;
  result?: FacilityJobResult;
  created_at?: string;
  updated_at?: string;
  completed_at?: string;
}

export function toFacilityJob(response: FacilityJobWire): FacilityJob {
  if (!response.job_id || !isKind(response.kind) || !isStatus(response.status) || !response.stage) {
    throw new Error('Invalid facility job response');
  }
  return {
    jobId: response.job_id,
    kind: response.kind,
    type: isType(response.type) ? response.type : 'copy',
    class: response.class === 'export' ? 'export' : 'mutation',
    status: response.status,
    progress: Math.min(100, Math.max(0, response.progress ?? 0)),
    stage: response.stage,
    attempts: response.attempts ?? 0,
    processed: response.processed ?? 0,
    ...(response.total !== undefined ? { total: response.total } : {}),
    successCount: response.success_count ?? 0,
    failureCount: response.failure_count ?? 0,
    retryable: response.retryable ?? false,
    ...(response.error ? { error: response.error } : {}),
    ...(response.result ? { result: response.result } : {}),
    ...(response.created_at ? { createdAt: response.created_at } : {}),
    ...(response.updated_at ? { updatedAt: response.updated_at } : {}),
    ...(response.completed_at ? { completedAt: response.completed_at } : {})
  };
}

function isKind(value: unknown): value is FacilityJobKind {
  return (
    value === 'control_cabinet' ||
    value === 'sps_controller' ||
    value === 'sps_controller_system_type' ||
    value === 'field_device' ||
    value === 'object_data'
  );
}

function isStatus(value: unknown): value is FacilityJobStatus {
  return value === 'queued' || value === 'running' || value === 'completed' || value === 'failed';
}

function isType(value: unknown): value is FacilityJobType {
  return (
    value === 'copy' ||
    value === 'export' ||
    value === 'bulk' ||
    value === 'delete' ||
    value === 'restore'
  );
}

export type CopyJob = FacilityJob;
export type CopyJobKind = FacilityJobKind;
export type CopyJobStatus = FacilityJobStatus;
export const toCopyJob = toFacilityJob;
