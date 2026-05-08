import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

type SessionResponse = {
  authenticated?: boolean;
};

export const load: PageLoad = async ({ fetch }) => {
  let isAuthenticated = false;

  try {
    const response = await fetch('/api/v1/auth/session', {
      credentials: 'include',
      headers: {
        Accept: 'application/json'
      }
    });

    if (response.ok) {
      const session = (await response.json()) as SessionResponse;
      isAuthenticated = session.authenticated === true;
    }
  } catch {
    isAuthenticated = false;
  }

  if (isAuthenticated) {
    throw redirect(302, '/');
  }
};
