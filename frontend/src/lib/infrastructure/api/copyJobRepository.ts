import { api } from '$lib/api/client.js';
import type { components } from '$lib/api/generated/schema.js';
import { toCopyJob, type CopyJob } from '$lib/domain/facility/copy-job.js';

type CopyJobResponse =
  components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.CopyJobResponse'];

export const copyJobRepository = {
  async get(id: string, signal?: AbortSignal): Promise<CopyJob> {
    const data = await api<CopyJobResponse>(`/facility/copy-jobs/${id}`, {
      signal,
      skipHttpErrorNavigation: true
    });
    return toCopyJob(data);
  }
};
