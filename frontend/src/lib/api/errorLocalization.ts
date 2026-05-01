import { t } from '$lib/i18n/index.js';

const FIELD_LABEL_KEYS: Record<string, string> = {
  apparat: 'facility.apparat',
  apparat_id: 'facility.apparat',
  apparat_nr: 'field_device.table.apparat_nr',
  bacnet_objects: 'facility.bacnet_object',
  bacnetobject: 'facility.bacnet_object',
  bmk: 'field_device.table.bmk',
  building_group: 'facility.building_group',
  building_id: 'facility.building',
  controlcabinet: 'facility.control_cabinet',
  control_cabinet_id: 'facility.control_cabinet',
  control_cabinet_nr: 'facility.forms.control_cabinet.number_label',
  description: 'common.description',
  device_name: 'facility.device_name',
  channel: 'notifications.preferences.channel_title',
  code: 'notifications.preferences.email.code_label',
  event_key: 'notifications.rules.event_key',
  fielddevice: 'facility.field_device',
  frequency: 'notifications.preferences.frequency_title',
  from_email: 'notifications.form.from_email',
  ga_device: 'facility.ga_device',
  gateway: 'facility.forms.sps_controller.gateway_label',
  host: 'notifications.form.host',
  ip_address: 'facility.ip_address',
  iws_code: 'facility.iws_code',
  name: 'common.name',
  notification_email: 'notifications.preferences.email.label',
  objectdata: 'facility.object_data',
  object_data_id: 'facility.object_data',
  password: 'notifications.form.password',
  phase_id: 'projects.settings.phase',
  port: 'notifications.form.port',
  project_id: 'notifications.rules.project_id',
  reply_to: 'notifications.form.reply_to',
  recipient_role: 'notifications.rules.role',
  recipient_team_id: 'notifications.rules.team_id',
  recipient_type: 'notifications.rules.recipient_type',
  recipient_user_ids: 'notifications.rules.user_ids',
  resource_id: 'notifications.rules.resource_id',
  resource_type: 'notifications.rules.resource_type',
  spscontroller: 'facility.sps_controller',
  specification: 'facility.specifications',
  subnet: 'facility.forms.sps_controller.subnet_label',
  system_part: 'facility.system_part',
  system_part_id: 'facility.system_part',
  system_types: 'facility.forms.sps_controller.system_types_title',
  text_fix: 'field_device.table.text_fix',
  to: 'notifications.test.to',
  username: 'notifications.form.username',
  vlan: 'facility.forms.sps_controller.vlan_label'
};

const SCOPE_LABEL_KEYS: Record<string, string> = {
  building: 'facility.building',
  control_cabinet: 'facility.control_cabinet',
  'control cabinet': 'facility.control_cabinet',
  ip_address: 'facility.ip_address',
  vlan: 'facility.forms.sps_controller.vlan_label'
};

const DIRECT_MESSAGE_KEYS: Record<string, string> = {
  'Bad Request': 'errors.bad_request',
  Conflict: 'errors.conflict',
  'Failed to load notification count': 'notifications.errors.notification_count_load_failed',
  'Failed to load notification preference': 'notifications.errors.preference_load_failed',
  'Failed to load notifications': 'notifications.errors.notifications_load_failed',
  'Failed to load notification rules': 'notifications.errors.rules_load_failed',
  'Failed to load SMTP settings': 'notifications.errors.smtp_settings_load_failed',
  'Failed to mark notification read': 'notifications.errors.mark_read_failed',
  'Failed to mark notifications read': 'notifications.errors.mark_all_read_failed',
  'Failed to save notification preference': 'notifications.errors.preference_save_failed',
  'Failed to save notification rule': 'notifications.errors.rule_save_failed',
  'Failed to save SMTP settings': 'notifications.errors.smtp_settings_save_failed',
  'Failed to verify notification email': 'notifications.errors.email_verification_failed',
  Forbidden: 'errors.forbidden',
  'Internal Server Error': 'errors.internal_server_error',
  'Notification not found': 'notifications.errors.notification_not_found',
  'Notification rule not found': 'notifications.errors.rule_not_found',
  'Not Found': 'errors.not_found',
  'SMTP settings not configured': 'notifications.errors.smtp_settings_not_configured',
  'Unknown error': 'errors.unknown_error',
  Unauthorized: 'errors.unauthorized',
  authorization_failed: 'errors.unauthorized',
  email_verification_failed: 'notifications.errors.email_verification_failed',
  fetch_failed: 'errors.fetch_failed',
  'field device is required': 'field_device.multi_create.validation.field_device_required',
  'notification provider disabled': 'notifications.errors.smtp_disabled',
  'notification provider not configured': 'notifications.errors.smtp_settings_not_configured',
  'Failed to delete notification rule': 'notifications.errors.rule_delete_failed',
  'no available ga_device for control cabinet': 'facility.no_available_ga_device',
  'object_data_id and bacnet_objects are mutually exclusive': 'facility.mutually_exclusive_error',
  'one or more parent entities (SPS controller, apparat, system part) not found':
    'field_device.multi_create.validation.parents_not_found',
  'apparat_nr is required': 'field_device.multi_create.validation.apparat_nr_required',
  'apparat_nr must be between 1 and 99': 'field_device.validation.apparat_nr_range',
  'apparatnummer ist bereits vergeben': 'field_device.multi_create.validation.apparat_nr_used',
  referenced_entity_in_use: 'facility.referenced_entity_in_use',
  smtp_test_failed: 'notifications.errors.smtp_test_failed',
  update_failed: 'errors.update_failed',
  validation_error: 'errors.validation_error'
};

function translateIfExists(key: string, params?: Record<string, string | number>): string | null {
  const translated = t(key, params);
  return translated !== key ? translated : null;
}

function extractFieldSegment(fieldPath?: string): string | undefined {
  if (!fieldPath) return undefined;
  const segments = fieldPath.split('.').filter(Boolean);
  return segments.length > 0 ? segments[segments.length - 1] : fieldPath;
}

function humanizeFieldName(field?: string): string {
  if (!field) return '';
  return field.replaceAll('_', ' ');
}

function getFieldLabel(fieldPath?: string): string {
  const field = extractFieldSegment(fieldPath);
  if (!field) return '';
  const translationKey = FIELD_LABEL_KEYS[field];
  if (translationKey) {
    return t(translationKey);
  }
  return humanizeFieldName(field);
}

function getScopeLabel(scope: string): string {
  const normalized = scope.trim().toLowerCase();
  const translationKey = SCOPE_LABEL_KEYS[normalized];
  if (translationKey) {
    return t(translationKey);
  }
  return humanizeFieldName(normalized);
}

export function localizeErrorText(message: string, fieldPath?: string): string {
  const trimmed = message.trim();
  if (!trimmed) return trimmed;

  const translatedByKey = translateIfExists(trimmed);
  if (translatedByKey) {
    return translatedByKey;
  }

  const translatedByErrorKey = translateIfExists(`errors.${trimmed}`);
  if (translatedByErrorKey) {
    return translatedByErrorKey;
  }

  const directTranslationKey = DIRECT_MESSAGE_KEYS[trimmed];
  if (directTranslationKey) {
    return t(directTranslationKey);
  }

  if (trimmed === 'is required') {
    return t('validation.required', { field: getFieldLabel(fieldPath) });
  }

  if (trimmed === 'must be a valid email') {
    return t('validation.email_invalid', { field: getFieldLabel(fieldPath) });
  }

  if (trimmed === 'invalid') {
    return t('validation.invalid', { field: getFieldLabel(fieldPath) });
  }

  if (
    /^(smtp dial|smtp tls dial|smtp client|smtp starttls|smtp auth|smtp mail from|smtp rcpt to|smtp data|smtp write|smtp close data|smtp quit):/i.test(
      trimmed
    )
  ) {
    return t('notifications.errors.smtp_delivery_failed');
  }

  if (/^(decode secret|decrypt secret):/i.test(trimmed)) {
    return t('notifications.errors.smtp_secret_failed');
  }

  let match = trimmed.match(/^([a-z0-9_.-]+) is required$/i);
  if (match) {
    return t('validation.required', { field: getFieldLabel(match[1]) });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) is required when auth_mode is plain$/i);
  if (match) {
    return t('validation.required_when_plain_auth', { field: getFieldLabel(match[1]) });
  }

  match = trimmed.match(/^max (\d+)$/i);
  if (match) {
    return t('validation.max_generic', { field: getFieldLabel(fieldPath), max: match[1] });
  }

  match = trimmed.match(/^min (\d+)$/i);
  if (match) {
    return t('validation.min_generic', { field: getFieldLabel(fieldPath), min: match[1] });
  }

  match = trimmed.match(/^length (\d+)$/i);
  if (match) {
    return t('validation.exact_length', { field: getFieldLabel(fieldPath), length: match[1] });
  }

  match = trimmed.match(/^must be one of: (.+)$/i);
  if (match) {
    return t('validation.one_of', { field: getFieldLabel(fieldPath), options: match[1] });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be (\d+) characters or less$/i);
  if (match) {
    return t('validation.max_length', { field: getFieldLabel(match[1]), max: match[2] });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be a valid IPv4 address$/i);
  if (match) {
    return t('validation.valid_ipv4', { field: getFieldLabel(match[1]) });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be a valid IPv4 subnet mask$/i);
  if (match) {
    return t('validation.valid_ipv4_subnet', { field: getFieldLabel(match[1]) });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be a number between (\d+) and (\d+)$/i);
  if (match) {
    return t('validation.number_between', {
      field: getFieldLabel(match[1]),
      min: match[2],
      max: match[3]
    });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be between (\d+) and (\d+)$/i);
  if (match) {
    return t('validation.range', { field: getFieldLabel(match[1]), min: match[2], max: match[3] });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be exactly (\d+) uppercase letters \(A-Z\)$/i);
  if (match) {
    return t('validation.exact_uppercase_letters', {
      field: getFieldLabel(match[1]),
      count: match[2]
    });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be unique within the (.+)$/i);
  if (match) {
    return t('validation.unique_within', {
      field: getFieldLabel(match[1]),
      scope: getScopeLabel(match[2])
    });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be unique per (.+)$/i);
  if (match) {
    return t('validation.unique_per', {
      field: getFieldLabel(match[1]),
      scope: getScopeLabel(match[2])
    });
  }

  return trimmed;
}

export function localizeFieldErrorMap(errors: Record<string, string>): Record<string, string> {
  return Object.fromEntries(
    Object.entries(errors).map(([field, value]) => [field, localizeErrorText(value, field)])
  );
}
