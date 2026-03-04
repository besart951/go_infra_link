import type { DashboardProjectStatus } from '$lib/domain/dashboard/index.js';

export function projectStatusVariant(status: DashboardProjectStatus):
  | 'secondary'
  | 'success'
  | 'default' {
  switch (status) {
    case 'ongoing':
      return 'default';
    case 'completed':
      return 'success';
    default:
      return 'secondary';
  }
}
