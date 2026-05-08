import { redirect } from '@sveltejs/kit';
import { hasUserPermission } from '$lib/utils/permissions.js';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ parent }) => {
  const { user } = await parent();

  if (!user) {
    throw redirect(302, '/login');
  }

  if (!hasUserPermission(user, 'notification.smtp.manage')) {
    throw redirect(302, '/');
  }

  return { user };
};
