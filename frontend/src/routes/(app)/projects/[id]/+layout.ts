import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types.js';
import { projectDetailService } from '$lib/components/project/ProjectDetailService.js';
import { canProject } from '$lib/utils/permissions.js';
import { forbiddenRouteRedirect } from '$lib/navigation/routeGuards.js';

export const load: LayoutLoad = async ({ params, url }) => {
  // Re-runs this capability check when the user moves between pages in the
  // same project. The service serves a 30-minute, realtime-refreshed cache.
  void url.pathname;

  const projectId = params.id;
  if (!projectId) {
    throw redirect(302, forbiddenRouteRedirect(url));
  }

  try {
    const { project, capabilities: projectCapabilities } =
      await projectDetailService.loadProjectContext(projectId);

    if (url.pathname.endsWith('/settings') && !canProject(projectCapabilities, 'project.update')) {
      throw redirect(302, forbiddenRouteRedirect(url));
    }

    return { project, projectCapabilities };
  } catch (error) {
    const route = projectDetailService.projectLoadErrorRoute(error, `${url.pathname}${url.search}`);
    if (route) {
      throw redirect(302, route);
    }
    throw error;
  }
};
