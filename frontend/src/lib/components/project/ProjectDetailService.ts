import { getProject } from '$lib/infrastructure/api/project.adapter.js';
import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
import { ApiException, buildHttpErrorRoute } from '$lib/api/client.js';
import { ProjectContextCache, type ProjectContext } from '$lib/services/projectContextCache.js';

function projectLoadErrorRoute(error: unknown, fromPath: string): string | undefined {
  if (!(error instanceof ApiException)) return undefined;
  return buildHttpErrorRoute(error.status, fromPath, error.message) ?? undefined;
}

const projectContextCache = new ProjectContextCache(async (projectId) => {
  const [project, capabilities] = await Promise.all([
    getProject(projectId),
    projectRepository.getCapabilities(projectId)
  ]);
  return { project, capabilities };
});

export const projectDetailService = {
  getProject(projectId: string) {
    return projectContextCache.load(projectId).then((context) => context.project);
  },
  getCapabilities(projectId: string) {
    return projectContextCache.load(projectId).then((context) => context.capabilities);
  },
  loadProjectContext(projectId: string): Promise<ProjectContext> {
    return projectContextCache.load(projectId);
  },
  refreshProjectContext(projectId: string): Promise<ProjectContext> {
    return projectContextCache.refresh(projectId);
  },
  clearProjectContextCache(): void {
    projectContextCache.clear();
  },
  projectLoadErrorRoute,
  listUsers(projectId: string) {
    return projectRepository.listUsers(projectId);
  }
};
