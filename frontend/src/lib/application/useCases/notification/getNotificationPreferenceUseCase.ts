import type { UserNotificationPreference } from '$lib/domain/notification/index.js';
import type { NotificationPreferenceRepository } from '$lib/domain/ports/notification/notificationPreferenceRepository.js';

export class GetNotificationPreferenceUseCase {
  constructor(private readonly repository: NotificationPreferenceRepository) {}

  execute(signal?: AbortSignal): Promise<UserNotificationPreference> {
    return this.repository.getUserPreference(signal);
  }
}
