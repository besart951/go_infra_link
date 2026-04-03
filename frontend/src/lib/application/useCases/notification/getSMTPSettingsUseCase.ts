import type { SMTPSettings } from '$lib/domain/notification/index.js';
import type { NotificationSettingsRepository } from '$lib/domain/ports/notification/notificationSettingsRepository.js';

export class GetSMTPSettingsUseCase {
  constructor(private readonly repository: NotificationSettingsRepository) {}

  execute(signal?: AbortSignal): Promise<SMTPSettings | null> {
    return this.repository.getSMTPSettings(signal);
  }
}
