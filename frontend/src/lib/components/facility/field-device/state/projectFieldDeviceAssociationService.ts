import { fetchAllPages } from '$lib/components/facility/shared/paginatedListFetcher.js';
import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
import type { ProjectFieldDeviceLink } from '$lib/domain/project/index.js';

export interface ProjectFieldDeviceRemovalResult {
  results: Array<{ id: string; success: boolean }>;
  total_count: number;
  success_count: number;
  failure_count: number;
}

export class ProjectFieldDeviceAssociationService {
  async removeFieldDevice(
    projectId: string,
    deviceId: string,
    missingLinkMessage: string
  ): Promise<void> {
    const link = await this.resolveFieldDeviceLink(projectId, deviceId);
    if (!link) {
      throw new Error(missingLinkMessage);
    }

    await projectRepository.removeFieldDevice({
      project_id: projectId,
      link_id: link.id,
      base_version: link.version
    });
  }

  async removeFieldDevices(
    projectId: string,
    ids: string[]
  ): Promise<ProjectFieldDeviceRemovalResult> {
    const linksByDeviceId = await this.loadFieldDeviceLinks(projectId);

    const results = await Promise.all(
      ids.map(async (id) => {
        const link = linksByDeviceId.get(id);
        if (!link) {
          return { id, success: false };
        }

        try {
          await projectRepository.removeFieldDevice({
            project_id: projectId,
            link_id: link.id,
            base_version: link.version
          });
          return { id, success: true };
        } catch {
          return { id, success: false };
        }
      })
    );

    const successCount = results.filter((item) => item.success).length;
    return {
      results,
      total_count: ids.length,
      success_count: successCount,
      failure_count: ids.length - successCount
    };
  }

  async resolveFieldDeviceLink(
    projectId: string,
    deviceId: string
  ): Promise<ProjectFieldDeviceLink | undefined> {
    const linksByDeviceId = await this.loadFieldDeviceLinks(projectId);
    return linksByDeviceId.get(deviceId);
  }

  async loadFieldDeviceLinks(projectId: string): Promise<Map<string, ProjectFieldDeviceLink>> {
    const links = await fetchAllPages((page, pageSize) =>
      projectRepository.listFieldDevices(projectId, { page, limit: pageSize })
    );
    return new Map(links.map((item) => [item.field_device_id, item]));
  }
}
