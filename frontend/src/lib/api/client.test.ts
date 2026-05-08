/// <reference types="vitest" />

const mockGoto = vi.hoisted(() => vi.fn());

vi.mock('$app/navigation', () => ({
  goto: mockGoto
}));

vi.mock('$lib/i18n/index.js', () => ({
  t: (key: string) => key
}));

import {
  api,
  ApiException,
  buildHttpErrorRoute,
  getFieldError,
  getFieldErrors,
  getHttpErrorPath,
  HandledApiException
} from './client.js';

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
    expect(
      buildHttpErrorRoute(
        403,
        '/projects/project-1',
        'Das Projekt befindet sich in der Phase "SIA 51".'
      )
    ).toBe(
      '/errors/403?from=%2Fprojects%2Fproject-1&message=Das+Projekt+befindet+sich+in+der+Phase+%22SIA+51%22.'
    );
  });

  it('navigates 403 responses to the forbidden page with the backend message', async () => {
    const backendMessage =
      'Das Projekt befindet sich in der Phase "SIA 51". In dieser Phase haben Sie keine Berechtigung.';
    const customFetch = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ error: 'authorization_failed', message: backendMessage }), {
        status: 403,
        statusText: 'Forbidden',
        headers: { 'Content-Type': 'application/json' }
      })
    );

    await expect(api('/projects/project-1', { customFetch })).rejects.toBeInstanceOf(
      HandledApiException
    );

    expect(mockGoto).toHaveBeenCalledWith(
      '/errors/403?from=%2Fprojects%2Fproject-1&message=Das+Projekt+befindet+sich+in+der+Phase+%22SIA+51%22.+In+dieser+Phase+haben+Sie+keine+Berechtigung.',
      {
        replaceState: true
      }
    );
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

    expect(mockGoto).toHaveBeenCalledWith(
      '/errors/404?from=%2Fprojects%2Fproject-1&message=errors.not_found',
      {
        replaceState: true
      }
    );
  });

  it('lets callers handle recoverable 404 responses without page navigation', async () => {
    const customFetch = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ error: 'not_found', message: 'Not Found' }), {
        status: 404,
        statusText: 'Not Found',
        headers: { 'Content-Type': 'application/json' }
      })
    );

    let caught: unknown;
    try {
      await api('/admin/notifications/smtp', {
        customFetch,
        skipHttpErrorNavigation: true
      });
    } catch (error) {
      caught = error;
    }

    expect(caught).toBeInstanceOf(ApiException);
    expect(caught).not.toBeInstanceOf(HandledApiException);
    expect(caught).toMatchObject({ status: 404, error: 'not_found' });
    expect(mockGoto).not.toHaveBeenCalled();
  });

  it('retries safe GET requests after a transient network failure', async () => {
    const customFetch = vi
      .fn()
      .mockRejectedValueOnce(new TypeError('Failed to fetch'))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ ok: true }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' }
        })
      );

    await expect(
      api<{ ok: boolean }>('/health', { customFetch, retryDelayMs: 0 })
    ).resolves.toEqual({
      ok: true
    });

    expect(customFetch).toHaveBeenCalledTimes(2);
  });

  it('does not retry unsafe POST requests after a network failure', async () => {
    const customFetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'));

    await expect(
      api('/teams', {
        customFetch,
        method: 'POST',
        body: JSON.stringify({ name: 'Team' }),
        retryDelayMs: 0
      })
    ).rejects.toMatchObject({ status: 0, error: 'network_error' });

    expect(customFetch).toHaveBeenCalledTimes(1);
  });

  it('retries safe GET requests on transient backend availability responses', async () => {
    const customFetch = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ error: 'unavailable', message: 'Service Unavailable' }), {
          status: 503,
          statusText: 'Service Unavailable',
          headers: { 'Content-Type': 'application/json' }
        })
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ ok: true }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' }
        })
      );

    await expect(
      api<{ ok: boolean }>('/dashboard', { customFetch, retryDelayMs: 0 })
    ).resolves.toEqual({
      ok: true
    });

    expect(customFetch).toHaveBeenCalledTimes(2);
  });

  it('maps unified field_errors responses onto form field paths', async () => {
    const customFetch = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          code: 'validation_error',
          error: 'validation_error',
          message: 'validation failed',
          field_errors: [
            {
              path: 'field_devices[0].bacnet_objects[0].text_fix',
              code: 'required',
              message: 'is required'
            }
          ],
          fields: {
            'field_devices[0].bacnet_objects[0].text_fix': 'is required'
          },
          request_id: 'req-1'
        }),
        {
          status: 400,
          statusText: 'Bad Request',
          headers: { 'Content-Type': 'application/json' }
        }
      )
    );

    let caught: unknown;
    try {
      await api('/facility/field-devices/multi-create', {
        customFetch,
        skipHttpErrorNavigation: true,
        method: 'POST',
        body: JSON.stringify({})
      });
    } catch (error) {
      caught = error;
    }

    expect(caught).toBeInstanceOf(ApiException);
    expect(getFieldErrors(caught)).toEqual({
      'field_devices[0].bacnet_objects[0].text_fix': 'validation.required'
    });
  });

  it('resolves nested field errors through domain aliases without leaking nested errors upward', () => {
    const errors = {
      'data.sps_controller.control_cabinet_id': 'cabinet required',
      'data.field_devices[0].bacnet_objects[0].alarm_type_id': 'alarm type required'
    };

    expect(getFieldError(errors, 'control_cabinet_id', ['spscontroller'])).toBe('cabinet required');
    expect(getFieldError(errors, 'alarm_type_id', ['fielddevice.bacnetobjects'])).toBe(
      'alarm type required'
    );
    expect(getFieldError(errors, 'alarm_type_id', ['fielddevice'])).toBeUndefined();
  });
});
