import type {
  UserNotificationPreference,
  VerifyUserNotificationEmailRequest
} from '$lib/domain/notification/index.js';
import type { NotificationPreferenceRepository } from '$lib/domain/ports/notification/notificationPreferenceRepository.js';

export class VerifyNotificationEmailUseCase {
  constructor(private readonly repository: NotificationPreferenceRepository) {}

  execute(
    data: VerifyUserNotificationEmailRequest,
    signal?: AbortSignal
  ): Promise<UserNotificationPreference> {
    return this.repository.verifyEmail(data, signal);
  }
}
