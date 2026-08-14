/* eslint-disable */
// This file is generated from backend/docs/swagger.json. Do not edit manually.

import createClient from 'openapi-fetch';
import { apiFetch, assertApiSuccess } from '../client.js';
import type { paths } from './schema.js';

const baseUrl = typeof window === 'undefined' ? 'http://localhost' : window.location.origin;

export const apiClient = createClient<paths>({
  baseUrl,
  fetch: apiFetch
});

apiClient.use({
  onResponse: async ({ response }) => assertApiSuccess(response)
});
