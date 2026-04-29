/// <reference types="vitest" />

const mockGoto = vi.hoisted(() => vi.fn());

vi.mock('$app/navigation', () => ({
  goto: mockGoto
}));

vi.mock('$lib/i18n/index.js', () => ({
  t: (key: string) => key
}));

import { api, buildHttpErrorRoute, getHttpErrorPath, HandledApiException } from './client.js';

describe('api client HTTP error navigation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.history.replaceState({}, '', '/projects/project-1');
  });

  it('maps backend middleware errors to app error pages', () => {
    expect(getHttpErrorPath(403)).toBe('/errors/403');
    expect(getHttpErrorPath(404)).toBe('/errors/404');
    expect(getHttpErrorPath(500)).toBeNull();
    expect(buildHttpErrorRoute(403, '/projects/project-1')).toBe(
      '/errors/403?from=%2Fprojects%2Fproject-1'
    );
  });

  it('navigates 403 responses to the forbidden page with replaceState', async () => {
    const customFetch = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ error: 'authorization_failed', message: 'Forbidden' }), {
        status: 403,
        statusText: 'Forbidden',
        headers: { 'Content-Type': 'application/json' }
      })
    );

    await expect(api('/projects/project-1', { customFetch })).rejects.toBeInstanceOf(
      HandledApiException
    );

    expect(mockGoto).toHaveBeenCalledWith('/errors/403?from=%2Fprojects%2Fproject-1', {
      replaceState: true
    });
  });

  it('navigates 404 responses to the not-found page with replaceState', async () => {
    const customFetch = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ error: 'not_found', message: 'Not Found' }), {
        status: 404,
        statusText: 'Not Found',
        headers: { 'Content-Type': 'application/json' }
      })
    );

    await expect(api('/projects/missing', { customFetch })).rejects.toBeInstanceOf(
      HandledApiException
    );

    expect(mockGoto).toHaveBeenCalledWith('/errors/404?from=%2Fprojects%2Fproject-1', {
      replaceState: true
    });
  });
});
