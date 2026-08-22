import { describe, expect, it } from 'vitest';

import {
  normalizeFacilityJob,
  reconcileFacilityJob,
  sameFacilityJobProgress
} from './facilityJobModel.js';

const baseJob = {
  jobId: 'job-1',
  kind: 'field_device' as const,
  status: 'running' as const,
  progress: 25,
  stage: 'processing_items'
};

describe('facility job reconciliation', () => {
  it('fills compatibility defaults in one place', () => {
    expect(normalizeFacilityJob(baseJob)).toMatchObject({
      type: 'copy',
      class: 'mutation',
      attempts: 0,
      processed: 0,
      successCount: 0,
      failureCount: 0,
      retryable: false
    });
  });

  it('replaces a job immutably without duplicating it', () => {
    const current = normalizeFacilityJob(baseJob);
    const currentJobs = [current];
    const jobs = reconcileFacilityJob(currentJobs, { ...baseJob, progress: 75 });

    expect(jobs).toHaveLength(1);
    expect(jobs[0]?.progress).toBe(75);
    expect(jobs).not.toBe(currentJobs);
  });

  it('detects equivalent progress hints', () => {
    const current = normalizeFacilityJob(baseJob);
    expect(sameFacilityJobProgress(current, { ...current })).toBe(true);
    expect(sameFacilityJobProgress(current, { ...current, processed: 2 })).toBe(false);
  });
});
