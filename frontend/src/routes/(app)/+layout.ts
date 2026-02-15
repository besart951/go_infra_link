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
        try {
            const userRes = await api<User>('/auth/me', { customFetch });
            user = userRes;
        } catch (e) {
            // 401 or network error; handled below if backend is unavailable.
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
