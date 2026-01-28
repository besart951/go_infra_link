import type { Pagination } from '../utils/index.js';

// StateText
export interface StateText {
	id: string;
	ref_number: number;
	state_text1?: string;
	state_text2?: string;
	state_text3?: string;
	state_text4?: string;
	state_text5?: string;
	state_text6?: string;
	state_text7?: string;
	state_text8?: string;
	state_text9?: string;
	state_text10?: string;
	state_text11?: string;
	state_text12?: string;
	state_text13?: string;
	state_text14?: string;
	state_text15?: string;
	state_text16?: string;
	created_at: string;
	updated_at: string;
}

export interface CreateStateTextRequest {
	ref_number: number;
	state_text1?: string;
	state_text2?: string;
	state_text3?: string;
	state_text4?: string;
	state_text5?: string;
	state_text6?: string;
	state_text7?: string;
	state_text8?: string;
	state_text9?: string;
	state_text10?: string;
	state_text11?: string;
	state_text12?: string;
	state_text13?: string;
	state_text14?: string;
	state_text15?: string;
	state_text16?: string;
}

export interface UpdateStateTextRequest {
	ref_number?: number;
	state_text1?: string;
	state_text2?: string;
	state_text3?: string;
	state_text4?: string;
	state_text5?: string;
	state_text6?: string;
	state_text7?: string;
	state_text8?: string;
	state_text9?: string;
	state_text10?: string;
	state_text11?: string;
	state_text12?: string;
	state_text13?: string;
	state_text14?: string;
	state_text15?: string;
	state_text16?: string;
}

export interface StateTextListParams {
	page?: number;
	limit?: number;
	search?: string;
}

export interface StateTextListResponse extends Pagination<StateText> {
	total_pages: number;
}

// NotificationClass
export interface NotificationClass {
	id: string;
	event_category: string;
	nc: number;
	object_description: string;
	internal_description: string;
	meaning: string;
	ack_required_not_normal: boolean;
	ack_required_error: boolean;
	ack_required_normal: boolean;
	norm_not_normal: number;
	norm_error: number;
	norm_normal: number;
	created_at: string;
	updated_at: string;
}

export interface CreateNotificationClassRequest {
	event_category: string;
	nc: number;
	object_description: string;
	internal_description: string;
	meaning: string;
	ack_required_not_normal?: boolean;
	ack_required_error?: boolean;
	ack_required_normal?: boolean;
	norm_not_normal?: number;
	norm_error?: number;
	norm_normal?: number;
}

export interface UpdateNotificationClassRequest {
	event_category?: string;
	nc?: number;
	object_description?: string;
	internal_description?: string;
	meaning?: string;
	ack_required_not_normal?: boolean;
	ack_required_error?: boolean;
	ack_required_normal?: boolean;
	norm_not_normal?: number;
	norm_error?: number;
	norm_normal?: number;
}

export interface NotificationClassListParams {
	page?: number;
	limit?: number;
	search?: string;
}

export interface NotificationClassListResponse extends Pagination<NotificationClass> {
	total_pages: number;
}

// AlarmDefinition
export interface AlarmDefinition {
	id: string;
	name: string;
	alarm_note?: string;
	created_at: string;
	updated_at: string;
}

export interface CreateAlarmDefinitionRequest {
	name: string;
	alarm_note?: string;
}

export interface UpdateAlarmDefinitionRequest {
	name?: string;
	alarm_note?: string;
}

export interface AlarmDefinitionListParams {
	page?: number;
	limit?: number;
	search?: string;
}

export interface AlarmDefinitionListResponse extends Pagination<AlarmDefinition> {
	total_pages: number;
}
