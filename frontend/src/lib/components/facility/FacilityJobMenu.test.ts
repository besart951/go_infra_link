import { fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { FacilityJob } from '$lib/domain/facility/facility-job.js';
import { facilityJobState } from '$lib/state/facilityJobState.svelte.js';
import FacilityJobMenuTestHarness from './FacilityJobMenuTestHarness.svelte';

const runningJob: FacilityJob = {
  jobId: '00000000-0000-4000-8000-000000000001',
  kind: 'sps_controller',
  type: 'copy',
  class: 'mutation',
  status: 'running',
  progress: 42,
  stage: 'copying_root',
  attempts: 1,
  processed: 1,
  total: 3,
  successCount: 1,
  failureCount: 0,
  retryable: true
};

const failedJob: FacilityJob = {
  ...runningJob,
  jobId: '00000000-0000-4000-8000-000000000002',
  status: 'failed',
  progress: 100,
  stage: 'failed',
  error: 'ERROR: invalid input syntax for type json (SQLSTATE 22P02)',
  failureCount: 1
};

describe('FacilityJobMenu', () => {
  beforeEach(() => {
    facilityJobState.jobs = [];
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('shows active-job badge, progress and cancel action', async () => {
    facilityJobState.jobs = [runningJob];
    const cancel = vi.spyOn(facilityJobState, 'cancel').mockResolvedValue();
    render(FacilityJobMenuTestHarness);

    const trigger = screen.getByLabelText('Jobs öffnen');
    expect(trigger).toHaveTextContent('1');
    await fireEvent.click(trigger);

    expect(await screen.findByText('42%')).toBeInTheDocument();
    await fireEvent.click(screen.getByLabelText('Abbrechen'));
    expect(cancel).toHaveBeenCalledWith(runningJob.jobId);
  });

  it('shows failed error with icon-only retry and remove actions', async () => {
    facilityJobState.jobs = [failedJob];
    const retry = vi.spyOn(facilityJobState, 'retry').mockResolvedValue();
    const remove = vi.spyOn(facilityJobState, 'remove').mockResolvedValue();
    render(FacilityJobMenuTestHarness);

    await fireEvent.click(screen.getByLabelText('Jobs öffnen'));

    expect(await screen.findByText(failedJob.error!)).toBeInTheDocument();
    await fireEvent.click(screen.getByLabelText('Erneut versuchen'));
    await fireEvent.click(screen.getByLabelText('Job entfernen'));
    expect(retry).toHaveBeenCalledWith(failedJob.jobId);
    expect(remove).toHaveBeenCalledWith(failedJob.jobId);
  });
});
