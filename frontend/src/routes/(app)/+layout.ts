import { api } from '$lib/api/client';
import type { LayoutLoad } from './$types';
import type { User } from '$lib/domain/user';
import type { Team } from '$lib/domain/team';
import type { Project } from '$lib/domain/project';

// Disable SSR for this layout and children
export const ssr = false;

export const load: LayoutLoad = async ({ fetch }) => {
	let backendAvailable = true;
	let user: User | null = null;
	let teams: Team[] = [];
	let projects: Project[] = [];

	const customFetch = fetch;

	try {
        // Parallel checks for health and potentially user if cookie exists
        // Since we don't know if cookie exists easily in client-side load without checking document.cookie
        // (which is synchronous), we can just try fetching 'me'.

        // But we want to check health first to set backendAvailable flag.
		const healthRes = await customFetch('/health'); // Relative to origin, processed by Caddy
        // Wait, Caddy proxies /api/* to go-infra.
        // If /health is on go-infra root, it needs to be accessed via /api/health or similar unless Caddy proxies /health?
        // Checking Caddyfile:
        // handle /api/* { ... }
        // handle /swagger/* { ... }
        // handle { ... frontend ... }
        
        // So fetching `/health` hitting Caddy will go to Frontend -> index.html.
        // Backend health check must be exposed under /api or we must update Caddy.
        // The backend routes.go logs `health=http://localhost:8080/health`.
        // So backend has /health.
        // But Caddy only proxies /api/*.
        
        // This is a preexisting config issue! 
        // If +layout.server.ts was using `${getBackendUrl()}/health` (http://go-infra:8080/health), it worked serverside.
        // Client-side, we must route via Caddy.
        // I should ADD /health to Caddyfile proxy, OR ask backend to move health to /api/health.
        
        // Modification to Caddyfile is easier.
        // Or I can skip health check and just try /api/v1/auth/me.
        
        try {
            const userRes = await api<User>('/auth/me', { customFetch });
            user = userRes;
        } catch (e) {
             // 401 or network error
             // If network error (backend down), api throws error?
             // checking api/client.ts:
             // catch { return { error: 'network_error', ... } } -> throws ApiException
        }

        if (user) {
            try {
                const [t, p] = await Promise.all([
                    api<Team[]>('/teams', { customFetch }),
                    api<Project[]>('/projects', { customFetch })
                ]);
                teams = t;
                projects = p;
            } catch (e) {
                console.error("Failed to load user data", e);
            }
        }
		
	} catch (e) {
        // If /auth/me failed with network error, backend might be down.
		backendAvailable = false; 
	}

	return {
		backendAvailable,
		user,
		teams,
		projects
	};
};
