import { ApiException, api } from '$lib/api/client';
import type { LayoutLoad } from './$types';
import type { User } from '$lib/domain/user';
import type { Team } from '$lib/domain/team';
import type { Project } from '$lib/domain/project';

// Disable SSR for this layout and children
export const ssr = false;

function isNetworkUnavailable(error: unknown): boolean {
  return error instanceof ApiException && error.status === 0 && error.error === 'network_error';
}

export const load: LayoutLoad = async ({ fetch }) => {
  let backendAvailable = true;
  let user: User | null = null;
  let teams: Team[] = [];
  let projects: Project[] = [];

  const customFetch = fetch;

  const hasPermission = (permission: string) => Boolean(user?.permissions?.includes(permission));

  try {
    try {
      const userRes = await api<User>('/auth/me', { customFetch });
      user = userRes;
    } catch (e) {
      if (isNetworkUnavailable(e)) {
        backendAvailable = false;
      }
      // 401 or missing session remains an unauthenticated user and is handled by the layout.
    }

    if (user) {
      try {
        const teamPromise = user.can_access_user_directory || !hasPermission('team.read')
          ? Promise.resolve([] as Team[])
          : api<Team[]>('/teams', { customFetch, skipHttpErrorNavigation: true });
        const projectPromise = hasPermission('project.listAll')
          ? api<Project[]>('/projects', { customFetch, skipHttpErrorNavigation: true })
          : Promise.resolve([] as Project[]);
        const [t, p] = await Promise.all([
          teamPromise,
          projectPromise
        ]);
        teams = t;
        projects = p;
      } catch (e) {
        if (isNetworkUnavailable(e)) {
          backendAvailable = false;
        }
        console.error('Failed to load user data', e);
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
