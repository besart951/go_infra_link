import { ApiException, api } from '$lib/api/client';
import type { LayoutLoad } from './$types';
import type { User } from '$lib/domain/user';
import type { Team } from '$lib/domain/team';
import type { Project, ProjectListResponse } from '$lib/domain/project';
import { hasUserPermission } from '$lib/utils/permissions.js';
import { redirect } from '@sveltejs/kit';
import { canAccessProtectedRoute, forbiddenRouteRedirect } from '$lib/navigation/routeGuards.js';

// Disable SSR for this layout and children
export const ssr = false;

function isNetworkUnavailable(error: unknown): boolean {
  return error instanceof ApiException && error.status === 0 && error.error === 'network_error';
}

export const load: LayoutLoad = async ({ fetch, url }) => {
  let backendAvailable = true;
  let user: User | null = null;
  let teams: Team[] = [];
  let projects: Project[] = [];

  const customFetch = fetch;

  const hasPermission = (permission: string) => hasUserPermission(user, permission);

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
        const teamPromise =
          user.can_access_user_directory || !hasPermission('team.read')
            ? Promise.resolve([] as Team[])
            : api<Team[]>('/teams', { customFetch, skipHttpErrorNavigation: true });
        const projectPromise = hasPermission('project.listAll')
          ? api<ProjectListResponse>('/projects?page=1&limit=10', {
              customFetch,
              skipHttpErrorNavigation: true
            }).then((response) => response.items ?? [])
          : Promise.resolve([] as Project[]);
        const [t, p] = await Promise.all([teamPromise, projectPromise]);
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

  if (user && !canAccessProtectedRoute(user, url.pathname)) {
    throw redirect(302, forbiddenRouteRedirect(url));
  }

  return {
    backendAvailable,
    user,
    teams,
    projects
  };
};
