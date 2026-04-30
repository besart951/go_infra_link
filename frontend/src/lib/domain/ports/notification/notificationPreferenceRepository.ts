import type {
  UpsertUserNotificationPreferenceRequest,
  VerifyUserNotificationEmailRequest,
  UserNotificationPreference
} from '$lib/domain/notification/index.js';

export interface NotificationPreferenceRepository {
  getUserPreference(signal?: AbortSignal): Promise<UserNotificationPreference>;
  saveUserPreference(
    data: UpsertUserNotificationPreferenceRequest,
    signal?: AbortSignal
  ): Promise<UserNotificationPreference>;
  sendEmailVerificationCode(signal?: AbortSignal): Promise<UserNotificationPreference>;
  verifyEmail(
    data: VerifyUserNotificationEmailRequest,
    signal?: AbortSignal
  ): Promise<UserNotificationPreference>;
}
