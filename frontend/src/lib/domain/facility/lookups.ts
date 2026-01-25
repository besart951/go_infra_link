import type { Pagination } from '../utils/index.js';

// StateText
export interface StateText {
	id: string;
	ref_number: number;
	state_text1?: string;
	created_at: string;
	updated_at: string;
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
	meaning: string;
	created_at: string;
	updated_at: string;
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

export interface AlarmDefinitionListParams {
	page?: number;
	limit?: number;
	search?: string;
}

export interface AlarmDefinitionListResponse extends Pagination<AlarmDefinition> {
	total_pages: number;
}
