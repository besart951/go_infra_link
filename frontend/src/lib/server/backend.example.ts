/**
 * Server-side utility to get the backend URL
 * Copy this file to backend.ts and configure as needed
 */

export function getBackendUrl(): string {
	return process.env.BACKEND_URL || 'http://localhost:8080';
}
