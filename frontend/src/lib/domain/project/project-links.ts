import type { Pagination } from '../utils/index.js';
import type { BulkOperationResultItem } from '../facility/field-device.js';

export interface ProjectControlCabinetLink {
  id: string;
  project_id: string;
  control_cabinet_id: string;
  revision: number;
  created_at: string;
  updated_at: string;
}

export interface ProjectSPSControllerLink {
  id: string;
  project_id: string;
  sps_controller_id: string;
  revision: number;
  created_at: string;
  updated_at: string;
}

export interface ProjectFieldDeviceLink {
  id: string;
  project_id: string;
  field_device_id: string;
  revision: number;
  created_at: string;
  updated_at: string;
}

export interface ProjectFieldDeviceMultiCreateResponse {
  success_field_device_ids: string[];
  association_errors: string[];
  results: BulkOperationResultItem[];
}

export interface ProjectControlCabinetListResponse extends Pagination<ProjectControlCabinetLink> {}
export interface ProjectSPSControllerListResponse extends Pagination<ProjectSPSControllerLink> {}
export interface ProjectFieldDeviceListResponse extends Pagination<ProjectFieldDeviceLink> {}
