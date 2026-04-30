import { getErrorMessage } from '$lib/api/client.js';
import type { SystemNotification } from '$lib/domain/notification/index.js';
import { systemNotificationRepository } from '$lib/infrastructure/api/systemNotificationRepository.js';

export class NotificationInboxPageState {
  notifications = $state<SystemNotification[]>([]);
  unreadCount = $state(0);
  page = $state(1);
  totalPages = $state(1);
  unreadOnly = $state(false);
  isLoading = $state(true);
  error = $state<string | null>(null);

  async loadNotifications(nextPage = this.page): Promise<void> {
    this.isLoading = true;
    this.error = null;
    try {
      const result = await systemNotificationRepository.list({
        page: nextPage,
        limit: 20,
        unread_only: this.unreadOnly
      });
      this.notifications = result.items;
      this.unreadCount = result.unread_count;
      this.page = result.page;
      this.totalPages = result.total_pages || 1;
    } catch (error) {
      this.error = getErrorMessage(error);
    } finally {
      this.isLoading = false;
    }
  }

  async markRead(notification: SystemNotification): Promise<void> {
    if (notification.read_at) return;
    try {
      await systemNotificationRepository.markRead(notification.id);
      await this.loadNotifications();
    } catch (error) {
      this.error = getErrorMessage(error);
    }
  }

  async markAllRead(): Promise<void> {
    try {
      await systemNotificationRepository.markAllRead();
      await this.loadNotifications(1);
    } catch (error) {
      this.error = getErrorMessage(error);
    }
  }

  toggleUnreadOnly(): void {
    this.unreadOnly = !this.unreadOnly;
    void this.loadNotifications(1);
  }

  formatDateTime(value: string): string {
    return new Intl.DateTimeFormat('de-CH', {
      dateStyle: 'medium',
      timeStyle: 'short'
    }).format(new Date(value));
  }
}
