/// <reference types="vitest" />

import { fieldErrorsFromApiDetails, parseApiErrorResponse } from './errorResponse.js';

describe('API error response helpers', () => {
  it('parses unified backend error responses', async () => {
    const response = new Response(
      JSON.stringify({
        error: 'validation_error',
        code: 'validation_error',
        message: 'validation failed',
        localized_key: 'errors.validation_error',
        field_errors: [{ path: 'name', code: 'required', message: 'is required' }],
        request_id: 'req-1'
      }),
      { status: 400, statusText: 'Bad Request' }
    );

    await expect(parseApiErrorResponse(response)).resolves.toMatchObject({
      error: 'validation_error',
      code: 'validation_error',
      message: 'validation failed',
      request_id: 'req-1',
      status: 400
    });
  });

  it('maps field_errors arrays to localized field maps', () => {
    expect(
      fieldErrorsFromApiDetails([
        {
          path: 'field_devices[0].bacnet_objects[0].text_fix',
          code: 'required',
          message: 'is required'
        }
      ])
    ).toEqual({
      'field_devices[0].bacnet_objects[0].text_fix': 'Text Fix ist erforderlich.'
    });
  });
});
