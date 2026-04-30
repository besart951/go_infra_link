import type {
  UpsertUserNotificationPreferenceRequest,
  UserNotificationPreference
} from '$lib/domain/notification/index.js';
import type { NotificationPreferenceRepository } from '$lib/domain/ports/notification/notificationPreferenceRepository.js';

export class SaveNotificationPreferenceUseCase {
  constructor(private readonly repository: NotificationPreferenceRepository) {}

  execute(
    data: UpsertUserNotificationPreferenceRequest,
    signal?: AbortSignal
  ): Promise<UserNotificationPreference> {
    return this.repository.saveUserPreference(data, signal);
  }
}
