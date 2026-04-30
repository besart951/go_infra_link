import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';

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
    const linkId = await this.resolveFieldDeviceLinkId(projectId, deviceId);
    if (!linkId) {
      throw new Error(missingLinkMessage);
    }

    await projectRepository.removeFieldDevice(projectId, linkId);
  }

  async removeFieldDevices(
    projectId: string,
    ids: string[]
  ): Promise<ProjectFieldDeviceRemovalResult> {
    const linkIdsByDeviceId = await this.loadFieldDeviceLinkIds(projectId);

    const results = await Promise.all(
      ids.map(async (id) => {
        const linkId = linkIdsByDeviceId.get(id);
        if (!linkId) {
          return { id, success: false };
        }

        try {
          await projectRepository.removeFieldDevice(projectId, linkId);
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

  async resolveFieldDeviceLinkId(projectId: string, deviceId: string): Promise<string | undefined> {
    const linkIdsByDeviceId = await this.loadFieldDeviceLinkIds(projectId);
    return linkIdsByDeviceId.get(deviceId);
  }

  async loadFieldDeviceLinkIds(projectId: string): Promise<Map<string, string>> {
    const links = await projectRepository.listFieldDevices(projectId, {
      page: 1,
      limit: 1000
    });
    return new Map(links.items.map((item) => [item.field_device_id, item.id]));
  }
}
