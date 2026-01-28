import type { LayoutServerLoad } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.ts';
import type { User } from '$lib/domain/user/index.js';
import type { Team } from '$lib/domain/team/index.js';
import type { Project } from '$lib/domain/project/index.js';

export const load: LayoutServerLoad = async ({ fetch, cookies }) => {
	let backendAvailable = true;
	let user: User | null = null;
	let teams: Team[] = [];
	let projects: Project[] = [];

	const accessToken = cookies.get('access_token');
	const headers: Record<string, string> = {};
	if (accessToken) {
		headers['Cookie'] = `access_token=${accessToken}`;
	}

	try {
		// Check backend health
		const healthRes = await fetch(`${getBackendUrl()}/health`, { method: 'GET' });
		backendAvailable = healthRes.ok;

		if (backendAvailable && accessToken) {
			// Fetch current user
			const userRes = await fetch(`${getBackendUrl()}/api/v1/auth/me`, { headers });
			if (userRes.ok) {
				user = await userRes.json();
			}

			// Fetch teams
			const teamsRes = await fetch(`${getBackendUrl()}/api/v1/teams`, { headers });
			if (teamsRes.ok) {
				const teamsData = await teamsRes.json();
				teams = Array.isArray(teamsData.teams)
					? teamsData.teams
					: Array.isArray(teamsData)
						? teamsData
						: [];
			}

			// Fetch projects (user can only see permitted projects)
			const projectsRes = await fetch(`${getBackendUrl()}/api/v1/projects`, { headers });
			if (projectsRes.ok) {
				const projectsData = await projectsRes.json();
				projects = Array.isArray(projectsData.projects)
					? projectsData.projects
					: Array.isArray(projectsData)
						? projectsData
						: [];
			}
		}
	} catch {
		backendAvailable = false;
	}

	return { backendAvailable, user, teams, projects };
};
