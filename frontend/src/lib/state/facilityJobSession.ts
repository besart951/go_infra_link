const storageKey = 'facility-copy-job';

export interface PersistedFacilityJob {
  jobId: string;
  ownerId?: string;
}

export function readFacilityJobHint(): PersistedFacilityJob | null {
  if (typeof sessionStorage === 'undefined') return null;
  try {
    return parseFacilityJobHint(JSON.parse(sessionStorage.getItem(storageKey) ?? 'null'));
  } catch {
    return null;
  }
}

export function persistFacilityJobHint(job: PersistedFacilityJob): void {
  if (typeof sessionStorage === 'undefined') return;
  sessionStorage.setItem(storageKey, JSON.stringify(job));
}

export function clearFacilityJobHint(): void {
  if (typeof sessionStorage === 'undefined') return;
  sessionStorage.removeItem(storageKey);
}

function parseFacilityJobHint(value: unknown): PersistedFacilityJob | null {
  if (
    !value ||
    typeof value !== 'object' ||
    !('jobId' in value) ||
    typeof value.jobId !== 'string'
  ) {
    return null;
  }
  return {
    jobId: value.jobId,
    ...('ownerId' in value && typeof value.ownerId === 'string' ? { ownerId: value.ownerId } : {})
  };
}
