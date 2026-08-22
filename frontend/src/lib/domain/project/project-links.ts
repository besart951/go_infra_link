import type { Pagination } from '../utils/index.js';

export interface ProjectControlCabinetLink {
  id: string;
  version: number;
  project_id: string;
  control_cabinet_id: string;
  created_at: string;
  updated_at: string;
}

export interface ProjectSPSControllerLink {
  id: string;
  version: number;
  project_id: string;
  sps_controller_id: string;
  created_at: string;
  updated_at: string;
}

export interface ProjectFieldDeviceLink {
  id: string;
  version: number;
  project_id: string;
  field_device_id: string;
  created_at: string;
  updated_at: string;
}

export interface ProjectLinkDeleteCommand {
  project_id: string;
  link_id: string;
  base_version: number;
}

export interface ProjectFieldDeviceMultiCreateResponse {
  success_field_device_ids: string[];
  association_errors: string[];
}

export interface ProjectControlCabinetListResponse extends Pagination<ProjectControlCabinetLink> {}
export interface ProjectSPSControllerListResponse extends Pagination<ProjectSPSControllerLink> {}
export interface ProjectFieldDeviceListResponse extends Pagination<ProjectFieldDeviceLink> {}
