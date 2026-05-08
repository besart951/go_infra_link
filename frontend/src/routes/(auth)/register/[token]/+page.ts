import type { PageLoad } from './$types';
import type { PublicRegistrationResponse } from '$lib/api/registrations.js';

type RegistrationLoadData = {
  token: string;
  registration: PublicRegistrationResponse | null;
  error: string | null;
};

export const load: PageLoad = async ({ fetch, params }): Promise<RegistrationLoadData> => {
  const token = params.token ?? '';
  try {
    const response = await fetch(`/api/v1/auth/registrations/${encodeURIComponent(token)}`, {
      credentials: 'include',
      headers: { Accept: 'application/json' }
    });

    if (!response.ok) {
      const error =
        response.status >= 500
          ? 'registration_unavailable'
          : response.status === 410
            ? 'registration_expired'
            : 'registration_invalid';
      return {
        token,
        registration: null,
        error
      };
    }

    return {
      token,
      registration: (await response.json()) as PublicRegistrationResponse,
      error: null
    };
  } catch {
    return {
      token,
      registration: null,
      error: 'registration_unavailable'
    };
  }
};
