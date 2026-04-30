export const NOTIFICATION_CHANNELS = ['email', 'system', 'both'] as const;
export type NotificationChannel = (typeof NOTIFICATION_CHANNELS)[number];

export const NOTIFICATION_FREQUENCIES = ['immediate', 'hourly', 'daily', 'weekly'] as const;
export type NotificationFrequency = (typeof NOTIFICATION_FREQUENCIES)[number];

export interface UserNotificationPreference {
  id?: string;
  user_id: string;
  notification_email: string;
  notification_email_verified_at?: string | null;
  email_verification_sent_at?: string | null;
  email_verification_expires_at?: string | null;
  channel: NotificationChannel;
  frequency: NotificationFrequency;
  created_at?: string;
  updated_at?: string;
}

export interface UpsertUserNotificationPreferenceRequest {
  notification_email: string;
  channel: NotificationChannel;
  frequency: NotificationFrequency;
}

export interface VerifyUserNotificationEmailRequest {
  code: string;
}

export type NotificationPreferenceFormValues = UpsertUserNotificationPreferenceRequest;

export function createNotificationPreferenceFormValues(
  preference?: UserNotificationPreference | null
): NotificationPreferenceFormValues {
  return {
    notification_email: preference?.notification_email ?? '',
    channel: preference?.channel ?? 'both',
    frequency: preference?.frequency ?? 'immediate'
  };
}

export function normalizeNotificationPreferenceInput(
  values: NotificationPreferenceFormValues
): UpsertUserNotificationPreferenceRequest {
  return {
    notification_email: values.notification_email.trim(),
    channel: values.channel,
    frequency: values.frequency
  };
}

export function hasVerifiedNotificationEmail(
  preference?: UserNotificationPreference | null
): boolean {
  return Boolean(preference?.notification_email && preference.notification_email_verified_at);
}

export function isNotificationChannel(value: string): value is NotificationChannel {
  return NOTIFICATION_CHANNELS.includes(value as NotificationChannel);
}

export function isNotificationFrequency(value: string): value is NotificationFrequency {
  return NOTIFICATION_FREQUENCIES.includes(value as NotificationFrequency);
}
