/**
 * Domain entity representing an SPS Controller
 */
export interface SPSController {
	id: string;
	control_cabinet_id: string;
	ga_device?: string | null;
	device_name: string;
	device_description?: string | null;
	device_location?: string | null;
	ip_address?: string | null;
	subnet?: string | null;
	gateway?: string | null;
	vlan?: string | null;
	created_at: string;
	updated_at: string;
}
