export interface CreateFieldDeviceExportRequest {
	project_ids?: string[];
	buildings_id?: string[];
	control_cabinet_id?: string[];
	sps_controller_id?: string[];
	force_async?: boolean;
}

export interface FieldDeviceExportJobResponse {
	job_id: string;
	status: 'queued' | 'processing' | 'completed' | 'failed';
	progress: number;
	message: string;
	output_type?: 'excel' | 'zip';
	file_name?: string;
	content_type?: string;
	download_url?: string;
	error?: string;
	created_at: string;
	updated_at: string;
}
