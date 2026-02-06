/**
 * Server-side utility to get the backend URL
 * Copy this file to backend.ts and configure as needed
 */

import { env } from '$env/dynamic/private';

export function getBackendUrl(): string {
	return env.BACKEND_URL || 'http://localhost:8080';
}
