import type { SendSMTPTestEmailRequest } from '$lib/domain/notification/index.js';
import type { NotificationSettingsRepository } from '$lib/domain/ports/notification/notificationSettingsRepository.js';

export class SendSMTPTestEmailUseCase {
  constructor(private readonly repository: NotificationSettingsRepository) {}

  execute(data: SendSMTPTestEmailRequest, signal?: AbortSignal): Promise<void> {
    return this.repository.sendSMTPTestEmail(data, signal);
  }
}
