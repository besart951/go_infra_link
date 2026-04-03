import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ parent }) => {
  const { user } = await parent();

  if (!user) {
    throw redirect(302, '/login');
  }

  if (user.role !== 'superadmin') {
    throw redirect(302, '/');
  }

  return { user };
};
