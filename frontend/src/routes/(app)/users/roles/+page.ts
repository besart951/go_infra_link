import { redirect } from '@sveltejs/kit';
import { canAccessRoleDirectory } from '$lib/navigation/userAccess.js';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ parent, url }) => {
  const { user } = await parent();

  if (!user) {
    throw redirect(302, '/login');
  }

  if (!canAccessRoleDirectory(user)) {
    const from = encodeURIComponent(`${url.pathname}${url.search}`);
    throw redirect(302, `/errors/403?from=${from}`);
  }

  return { user };
};
