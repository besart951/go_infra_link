import { getProject } from '$lib/infrastructure/api/project.adapter.js';
import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';

export const projectDetailService = {
  getProject,
  listUsers(projectId: string) {
    return projectRepository.listUsers(projectId);
  }
};
