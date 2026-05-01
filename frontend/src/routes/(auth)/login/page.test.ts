import { load } from './+page.js';

function createLoadEvent(responseOrError: Response | Error) {
  const fetch = vi.fn().mockImplementation(async () => {
    if (responseOrError instanceof Error) {
      throw responseOrError;
    }

    return responseOrError;
  });

  return {
    event: { fetch } as unknown as Parameters<typeof load>[0],
    fetch
  };
}

describe('/login load', () => {
  it('redirects authenticated users to the dashboard', async () => {
    const { event, fetch } = createLoadEvent(new Response('{}', { status: 200 }));

    await expect(load(event)).rejects.toMatchObject({
      status: 302,
      location: '/'
    });
    expect(fetch).toHaveBeenCalledWith('/api/v1/auth/me', {
      credentials: 'include',
      headers: {
        Accept: 'application/json'
      }
    });
  });

  it('keeps unauthenticated users on the login page', async () => {
    const { event } = createLoadEvent(new Response('{}', { status: 401 }));

    await expect(load(event)).resolves.toBeUndefined();
  });

  it('keeps the login page reachable when the auth check cannot complete', async () => {
    const { event } = createLoadEvent(new Error('network unavailable'));

    await expect(load(event)).resolves.toBeUndefined();
  });
});
