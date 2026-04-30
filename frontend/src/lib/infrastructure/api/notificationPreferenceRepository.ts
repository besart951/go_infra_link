import { api } from '$lib/api/client.js';
import type {
  UpsertUserNotificationPreferenceRequest,
  VerifyUserNotificationEmailRequest,
  UserNotificationPreference
} from '$lib/domain/notification/index.js';
import type { NotificationPreferenceRepository } from '$lib/domain/ports/notification/notificationPreferenceRepository.js';

export const notificationPreferenceRepository: NotificationPreferenceRepository = {
  getUserPreference(signal?: AbortSignal): Promise<UserNotificationPreference> {
    return api<UserNotificationPreference>('/account/notifications/preferences', { signal });
  },

  saveUserPreference(
    data: UpsertUserNotificationPreferenceRequest,
    signal?: AbortSignal
  ): Promise<UserNotificationPreference> {
    return api<UserNotificationPreference>('/account/notifications/preferences', {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },

  sendEmailVerificationCode(signal?: AbortSignal): Promise<UserNotificationPreference> {
    return api<UserNotificationPreference>(
      '/account/notifications/preferences/email-verification',
      {
        method: 'POST',
        signal
      }
    );
  },

  verifyEmail(
    data: VerifyUserNotificationEmailRequest,
    signal?: AbortSignal
  ): Promise<UserNotificationPreference> {
    return api<UserNotificationPreference>(
      '/account/notifications/preferences/email-verification/verify',
      {
        method: 'POST',
        body: JSON.stringify(data),
        signal
      }
    );
  }
};
