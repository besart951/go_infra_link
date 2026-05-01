import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => {
  let isAuthenticated = false;

  try {
    const response = await fetch('/api/v1/auth/me', {
      credentials: 'include',
      headers: {
        Accept: 'application/json'
      }
    });

    isAuthenticated = response.ok;
  } catch {
    isAuthenticated = false;
  }

  if (isAuthenticated) {
    throw redirect(302, '/');
  }
};
