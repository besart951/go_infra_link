import { historyFieldLabel, historyTableLabel } from './historyLabels.js';

export interface HistoryEntityFilterOption {
  id: string;
  label: string;
  fields: string[];
}

const ENTITY_FIELDS: Array<Omit<HistoryEntityFilterOption, 'label'>> = [
  {
    id: 'buildings',
    fields: ['iws_code', 'building_group']
  },
  {
    id: 'control_cabinets',
    fields: ['control_cabinet_nr', 'building_id']
  },
  {
    id: 'sps_controllers',
    fields: [
      'ga_device',
      'device_name',
      'device_description',
      'device_location',
      'ip_address',
      'subnet',
      'gateway',
      'vlan',
      'control_cabinet_id'
    ]
  },
  {
    id: 'sps_controller_system_types',
    fields: ['sps_controller_id', 'system_type_id', 'number', 'document_name']
  },
  {
    id: 'field_devices',
    fields: [
      'bmk',
      'description',
      'text_fix',
      'apparat_nr',
      'sps_controller_system_type_id',
      'system_part_id',
      'specification_id',
      'apparat_id'
    ]
  },
  {
    id: 'specifications',
    fields: [
      'specification_supplier',
      'specification_brand',
      'specification_type',
      'additional_info_motor_valve',
      'additional_info_size',
      'additional_information_installation_location',
      'electrical_connection_ph',
      'electrical_connection_acdc',
      'electrical_connection_amperage',
      'electrical_connection_power',
      'electrical_connection_rotation'
    ]
  },
  {
    id: 'bacnet_objects',
    fields: [
      'text_fix',
      'description',
      'gms_visible',
      'optional',
      'text_individual',
      'software_type',
      'software_number',
      'hardware_type',
      'hardware_quantity',
      'software_reference_id',
      'state_text_id',
      'notification_class_id',
      'alarm_type_id'
    ]
  },
  {
    id: 'bacnet_object_alarm_values',
    fields: [
      'alarm_type_field_id',
      'value_number',
      'value_integer',
      'value_boolean',
      'value_string',
      'value_json',
      'unit_id',
      'source'
    ]
  },
  {
    id: 'object_data',
    fields: ['description', 'version', 'is_active', 'project_id']
  },
  {
    id: 'system_types',
    fields: ['number_min', 'number_max', 'name']
  },
  {
    id: 'system_parts',
    fields: ['short_name', 'name', 'description']
  },
  {
    id: 'apparats',
    fields: ['short_name', 'name', 'description']
  },
  {
    id: 'state_texts',
    fields: [
      'ref_number',
      'state_text1',
      'state_text2',
      'state_text3',
      'state_text4',
      'state_text5',
      'state_text6',
      'state_text7',
      'state_text8',
      'state_text9',
      'state_text10',
      'state_text11',
      'state_text12',
      'state_text13',
      'state_text14',
      'state_text15',
      'state_text16'
    ]
  },
  {
    id: 'notification_classes',
    fields: [
      'event_category',
      'nc',
      'object_description',
      'internal_description',
      'meaning',
      'ack_required_not_normal',
      'ack_required_error',
      'ack_required_normal',
      'norm_not_normal',
      'norm_error',
      'norm_normal'
    ]
  },
  {
    id: 'alarm_definitions',
    fields: ['name', 'alarm_note', 'alarm_type_id', 'scope']
  },
  {
    id: 'alarm_fields',
    fields: ['key', 'label', 'data_type', 'default_unit_code']
  },
  {
    id: 'alarm_types',
    fields: ['code', 'name']
  },
  {
    id: 'alarm_type_fields',
    fields: [
      'alarm_type_id',
      'alarm_field_id',
      'display_order',
      'is_required',
      'is_user_editable',
      'default_value_json',
      'validation_json',
      'default_unit_id',
      'ui_group'
    ]
  },
  {
    id: 'units',
    fields: ['code', 'symbol', 'name']
  },
  {
    id: 'projects',
    fields: ['name', 'description']
  },
  {
    id: 'project_control_cabinets',
    fields: ['project_id', 'control_cabinet_id']
  },
  {
    id: 'project_sps_controllers',
    fields: ['project_id', 'sps_controller_id']
  },
  {
    id: 'project_field_devices',
    fields: ['project_id', 'field_device_id']
  }
];

export function historyEntityFilterOptions(): HistoryEntityFilterOption[] {
  return ENTITY_FIELDS.map((option) => ({
    ...option,
    label: historyTableLabel(option.id)
  })).sort((a, b) => a.label.localeCompare(b.label, 'de-CH'));
}

export function historyFieldFilterOptions(
  entityTable?: string
): Array<{ id: string; label: string }> {
  const fields = entityTable
    ? (ENTITY_FIELDS.find((option) => option.id === entityTable)?.fields ?? [])
    : [...new Set(ENTITY_FIELDS.flatMap((option) => option.fields))];

  return fields
    .map((field) => ({ id: field, label: historyFieldLabel(field) }))
    .sort((a, b) => a.label.localeCompare(b.label, 'de-CH'));
}
