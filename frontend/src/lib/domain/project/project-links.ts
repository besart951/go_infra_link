import type { Pagination } from '../utils/index.js';

export interface ProjectControlCabinetLink {
	id: string;
	project_id: string;
	control_cabinet_id: string;
	created_at: string;
	updated_at: string;
}

export interface ProjectSPSControllerLink {
	id: string;
	project_id: string;
	sps_controller_id: string;
	created_at: string;
	updated_at: string;
}

export interface ProjectFieldDeviceLink {
	id: string;
	project_id: string;
	field_device_id: string;
	created_at: string;
	updated_at: string;
}

export interface ProjectControlCabinetListResponse extends Pagination<ProjectControlCabinetLink> {}
export interface ProjectSPSControllerListResponse extends Pagination<ProjectSPSControllerLink> {}
export interface ProjectFieldDeviceListResponse extends Pagination<ProjectFieldDeviceLink> {}
