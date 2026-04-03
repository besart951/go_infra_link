export const SMTP_PROVIDER = 'smtp' as const;

export const SMTP_SECURITY_MODES = ['none', 'starttls', 'tls'] as const;
export type SMTPSecurityMode = (typeof SMTP_SECURITY_MODES)[number];

export const SMTP_AUTH_MODES = ['none', 'plain'] as const;
export type SMTPAuthMode = (typeof SMTP_AUTH_MODES)[number];

export const SMTP_DEFAULT_PORTS: Record<SMTPSecurityMode, number> = {
  none: 25,
  starttls: 587,
  tls: 465
};

export interface SMTPSettings {
  id: string;
  provider: typeof SMTP_PROVIDER;
  enabled: boolean;
  host: string;
  port: number;
  username: string;
  has_password: boolean;
  from_email: string;
  from_name: string;
  reply_to: string;
  security: SMTPSecurityMode;
  auth_mode: SMTPAuthMode;
  allow_insecure_tls: boolean;
  updated_at: string;
  updated_by_id?: string | null;
}

export interface UpsertSMTPSettingsRequest {
  enabled: boolean;
  host: string;
  port: number;
  username: string;
  password: string;
  from_email: string;
  from_name: string;
  reply_to: string;
  security: SMTPSecurityMode;
  auth_mode: SMTPAuthMode;
  allow_insecure_tls: boolean;
}

export interface SendSMTPTestEmailRequest {
  to: string;
  subject: string;
  body: string;
}

export type SMTPSettingsFormValues = UpsertSMTPSettingsRequest;
export type SMTPTestEmailFormValues = SendSMTPTestEmailRequest;

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function createSMTPSettingsFormValues(
  settings?: SMTPSettings | null
): SMTPSettingsFormValues {
  return {
    enabled: settings?.enabled ?? true,
    host: settings?.host ?? '',
    port: settings?.port ?? SMTP_DEFAULT_PORTS.starttls,
    username: settings?.username ?? '',
    password: '',
    from_email: settings?.from_email ?? '',
    from_name: settings?.from_name ?? '',
    reply_to: settings?.reply_to ?? '',
    security: settings?.security ?? 'starttls',
    auth_mode: settings?.auth_mode ?? 'plain',
    allow_insecure_tls: settings?.allow_insecure_tls ?? false
  };
}

export function createSMTPTestEmailFormValues(defaultRecipient = ''): SMTPTestEmailFormValues {
  return {
    to: defaultRecipient,
    subject: 'SMTP-Konfigurationstest',
    body: 'Dies ist eine Test-E-Mail aus Infra Link.'
  };
}

export function getSMTPDefaultPort(security: SMTPSecurityMode): number {
  return SMTP_DEFAULT_PORTS[security];
}

export function hasSMTPConfiguration(settings?: SMTPSettings | null): settings is SMTPSettings {
  return settings !== null && settings !== undefined;
}

export function isSMTPPasswordRequired(
  authMode: SMTPAuthMode,
  hasStoredPassword: boolean
): boolean {
  return authMode === 'plain' && !hasStoredPassword;
}

export function normalizeSMTPSettingsInput(
  values: SMTPSettingsFormValues
): UpsertSMTPSettingsRequest {
  const normalized: UpsertSMTPSettingsRequest = {
    enabled: values.enabled,
    host: values.host.trim(),
    port: Number(values.port) || 0,
    username: values.username.trim(),
    password: values.password.trim(),
    from_email: values.from_email.trim(),
    from_name: values.from_name.trim(),
    reply_to: values.reply_to.trim(),
    security: values.security,
    auth_mode: values.auth_mode,
    allow_insecure_tls: values.security === 'none' ? false : values.allow_insecure_tls
  };

  if (normalized.auth_mode === 'none') {
    normalized.username = '';
    normalized.password = '';
  }

  return normalized;
}

export function normalizeSMTPTestEmailInput(
  values: SMTPTestEmailFormValues
): SendSMTPTestEmailRequest {
  return {
    to: values.to.trim(),
    subject: values.subject.trim(),
    body: values.body.trim()
  };
}

export function validateSMTPSettingsInput(
  input: UpsertSMTPSettingsRequest,
  hasStoredPassword: boolean
): Record<string, string> {
  const errors: Record<string, string> = {};

  if (!input.host) {
    errors.host = 'notifications.validation.host_required';
  }

  if (!Number.isInteger(input.port) || input.port < 1 || input.port > 65535) {
    errors.port = 'notifications.validation.port_range';
  }

  if (!input.from_email || !EMAIL_PATTERN.test(input.from_email)) {
    errors.from_email = 'notifications.validation.email_invalid';
  }

  if (input.reply_to && !EMAIL_PATTERN.test(input.reply_to)) {
    errors.reply_to = 'notifications.validation.email_invalid';
  }

  if (input.auth_mode === 'plain') {
    if (!input.username) {
      errors.username = 'notifications.validation.username_required';
    }

    if (isSMTPPasswordRequired(input.auth_mode, hasStoredPassword) && !input.password) {
      errors.password = 'notifications.validation.password_required';
    }
  }

  return errors;
}

export function validateSMTPTestEmailInput(
  input: SendSMTPTestEmailRequest
): Record<string, string> {
  const errors: Record<string, string> = {};

  if (!input.to || !EMAIL_PATTERN.test(input.to)) {
    errors.to = 'notifications.validation.email_invalid';
  }

  return errors;
}
