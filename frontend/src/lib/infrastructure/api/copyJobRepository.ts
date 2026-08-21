import { api } from '$lib/api/client.js';
import { toFacilityJob, type FacilityJob } from '$lib/domain/facility/copy-job.js';

interface FacilityJobListWire {
  items?: Parameters<typeof toFacilityJob>[0][];
  next_cursor?: string;
  previous_cursor?: string;
}

export interface FacilityJobPage {
  items: FacilityJob[];
  nextCursor?: string;
  previousCursor?: string;
}

export const facilityJobRepository = {
  async list(cursor?: string, signal?: AbortSignal): Promise<FacilityJobPage> {
    const params = new URLSearchParams({ limit: '50' });
    if (cursor) params.set('cursor', cursor);
    const data = await api<FacilityJobListWire>(`/facility/jobs?${params}`, {
      signal,
      skipHttpErrorNavigation: true
    });
    return {
      items: (data.items ?? []).map(toFacilityJob),
      ...(data.next_cursor ? { nextCursor: data.next_cursor } : {}),
      ...(data.previous_cursor ? { previousCursor: data.previous_cursor } : {})
    };
  },

  async get(id: string, signal?: AbortSignal): Promise<FacilityJob> {
    const data = await api<Parameters<typeof toFacilityJob>[0]>(`/facility/jobs/${id}`, {
      signal,
      skipHttpErrorNavigation: true
    });
    return toFacilityJob(data);
  },

  async retry(id: string, signal?: AbortSignal): Promise<FacilityJob> {
    const data = await api<Parameters<typeof toFacilityJob>[0]>(`/facility/jobs/${id}/retry`, {
      method: 'POST',
      signal,
      skipHttpErrorNavigation: true
    });
    return toFacilityJob(data);
  }
};

export const copyJobRepository = {
  async list(signal?: AbortSignal): Promise<FacilityJob[]> {
    return (await facilityJobRepository.list(undefined, signal)).items;
  },
  get: facilityJobRepository.get,
  retry: facilityJobRepository.retry
};
