import type { VersionedDeleteCommand } from '$lib/domain/ports/crudRepository.js';
import type { ProjectLinkDeleteCommand } from '$lib/domain/project/project-links.js';

export function versionedDeletePath(path: string, command: VersionedDeleteCommand): string {
  const query = new URLSearchParams({ base_version: String(command.base_version) });
  return `${path}/${encodeURIComponent(command.id)}?${query.toString()}`;
}

export function versionedProjectLinkDeletePath(
  resource: 'control-cabinets' | 'sps-controllers' | 'field-devices',
  command: ProjectLinkDeleteCommand
): string {
  return versionedDeletePath(`/projects/${encodeURIComponent(command.project_id)}/${resource}`, {
    id: command.link_id,
    base_version: command.base_version
  });
}
