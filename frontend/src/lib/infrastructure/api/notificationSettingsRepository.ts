import { ApiException, api } from '$lib/api/client.js';
import type {
  SendSMTPTestEmailRequest,
  SMTPSettings,
  UpsertSMTPSettingsRequest
} from '$lib/domain/notification/index.js';
import type { NotificationSettingsRepository } from '$lib/domain/ports/notification/notificationSettingsRepository.js';

export const notificationSettingsRepository: NotificationSettingsRepository = {
  async getSMTPSettings(signal?: AbortSignal): Promise<SMTPSettings | null> {
    try {
      return await api<SMTPSettings>('/admin/notifications/smtp', { signal });
    } catch (error) {
      if (error instanceof ApiException && error.status === 404) {
        return null;
      }

      throw error;
    }
  },

  async saveSMTPSettings(
    data: UpsertSMTPSettingsRequest,
    signal?: AbortSignal
  ): Promise<SMTPSettings> {
    return api<SMTPSettings>('/admin/notifications/smtp', {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },

  async sendSMTPTestEmail(data: SendSMTPTestEmailRequest, signal?: AbortSignal): Promise<void> {
    return api<void>('/admin/notifications/smtp/test', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  }
};
