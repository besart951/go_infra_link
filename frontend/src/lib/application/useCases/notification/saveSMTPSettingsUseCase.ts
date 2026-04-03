import type { SMTPSettings, UpsertSMTPSettingsRequest } from '$lib/domain/notification/index.js';
import type { NotificationSettingsRepository } from '$lib/domain/ports/notification/notificationSettingsRepository.js';

export class SaveSMTPSettingsUseCase {
  constructor(private readonly repository: NotificationSettingsRepository) {}

  execute(data: UpsertSMTPSettingsRequest, signal?: AbortSignal): Promise<SMTPSettings> {
    return this.repository.saveSMTPSettings(data, signal);
  }
}
