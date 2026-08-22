/**
 * AlarmType domain types
 * Mirrors backend: internal/domain/facility/alarm_type.go
 */

export interface Unit {
  id: string;
  version: number;
  code: string;
  symbol: string;
  name: string;
}

export interface CreateUnitRequest {
  code: string;
  symbol: string;
  name: string;
}

export interface UpdateUnitRequest {
  base_version: number;
  code?: string;
  symbol?: string;
  name?: string;
}

export interface AlarmField {
  id: string;
  version: number;
  key: string;
  label: string;
  data_type:
    | 'number'
    | 'integer'
    | 'boolean'
    | 'string'
    | 'enum'
    | 'duration'
    | 'state_map'
    | 'json';
  default_unit_code?: string;
}

export interface CreateAlarmFieldRequest {
  key: string;
  label: string;
  data_type: AlarmField['data_type'];
  default_unit_code?: string;
}

export interface UpdateAlarmFieldRequest {
  base_version: number;
  key?: string;
  label?: string;
  data_type?: AlarmField['data_type'];
  default_unit_code?: string;
}

export interface AlarmTypeField {
  id: string;
  version: number;
  alarm_type_id: string;
  alarm_field_id: string;
  alarm_field?: AlarmField;
  display_order: number;
  is_required: boolean;
  is_user_editable: boolean;
  default_value_json?: string;
  validation_json?: string;
  default_unit_id?: string;
  default_unit?: Unit;
  ui_group?: string;
  created_at: string;
  updated_at: string;
}

export interface AlarmType {
  id: string;
  version: number;
  code: string;
  name: string;
  fields?: AlarmTypeField[];
  created_at: string;
  updated_at: string;
}

export interface CreateAlarmTypeRequest {
  code: string;
  name: string;
}

export interface UpdateAlarmTypeRequest {
  base_version: number;
  name?: string;
}

export interface CreateAlarmTypeFieldRequest {
  alarm_field_id: string;
  display_order?: number;
  is_required?: boolean;
  is_user_editable?: boolean;
  default_value_json?: string;
  validation_json?: string;
  default_unit_id?: string;
  ui_group?: string;
}

export interface UpdateAlarmTypeFieldRequest {
  base_version: number;
  display_order?: number;
  is_required?: boolean;
  is_user_editable?: boolean;
  default_value_json?: string;
  validation_json?: string;
  default_unit_id?: string;
  ui_group?: string;
}

export interface AlarmTypeListResponse {
  items: AlarmType[];
  total: number;
  page: number;
  total_pages: number;
}

export interface AlarmValueDraft {
  alarm_type_field_id: string;
  value_number?: number;
  value_integer?: number;
  value_boolean?: boolean;
  value_string?: string;
  value_json?: string;
  unit_id?: string;
  source?: string;
}

export interface AlarmValue {
  id: string;
  bacnet_object_id: string;
  alarm_type_field_id: string;
  value_number?: number;
  value_integer?: number;
  value_boolean?: boolean;
  value_string?: string;
  value_json?: string;
  unit_id?: string;
  source: string;
  created_at: string;
  updated_at: string;
}

export interface AlarmValuesResponse {
  items: AlarmValue[];
}
