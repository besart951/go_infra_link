/**
 * Server-side utility to get the backend URL
 * Copy this file to backend.ts and configure as needed
 */

import { env } from '$env/dynamic/private';

export function getBackendUrl(): string {
  const backendPort = env.BACKEND_PORT ?? '8080';
  return env.BACKEND_URL ?? `http://localhost:${backendPort}`;
}
