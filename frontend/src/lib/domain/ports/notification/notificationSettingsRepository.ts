import type {
  SendSMTPTestEmailRequest,
  SMTPSettings,
  UpsertSMTPSettingsRequest
} from '$lib/domain/notification/index.js';

export interface NotificationSettingsRepository {
  getSMTPSettings(signal?: AbortSignal): Promise<SMTPSettings | null>;
  saveSMTPSettings(data: UpsertSMTPSettingsRequest, signal?: AbortSignal): Promise<SMTPSettings>;
  sendSMTPTestEmail(data: SendSMTPTestEmailRequest, signal?: AbortSignal): Promise<void>;
}
