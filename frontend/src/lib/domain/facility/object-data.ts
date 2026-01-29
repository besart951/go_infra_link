import type { Pagination } from '../utils/index.js';
import type { BacnetObject } from './bacnet-object.js';
import type { BacnetObjectInput } from './field-device.js';

export interface ObjectData {
	id: string;
	description: string;
	version: string;
	is_active: boolean;
	project_id?: string;
	bacnet_objects?: BacnetObject[];
	created_at: string;
	updated_at: string;
}

export interface CreateObjectDataRequest {
	description: string;
	version: string;
	is_active?: boolean;
	project_id?: string;
	bacnet_objects?: BacnetObjectInput[];
}

export interface UpdateObjectDataRequest {
	description?: string;
	version?: string;
	is_active?: boolean;
	project_id?: string;
}

export interface ObjectDataListParams {
	page?: number;
	limit?: number;
	search?: string;
}

export interface ObjectDataListResponse extends Pagination<ObjectData> {
	total_pages: number;
}
