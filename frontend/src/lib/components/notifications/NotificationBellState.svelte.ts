import { getErrorMessage } from '$lib/api/client.js';
import type { SystemNotification } from '$lib/domain/notification/index.js';
import { systemNotificationRepository } from '$lib/infrastructure/api/systemNotificationRepository.js';

export class NotificationBellState {
  items = $state<SystemNotification[]>([]);
  unreadCount = $state(0);
  isLoading = $state(false);
  error = $state<string | null>(null);
  refreshTimer: ReturnType<typeof setInterval> | undefined;

  async loadNotifications(): Promise<void> {
    this.isLoading = true;
    this.error = null;
    try {
      const result = await systemNotificationRepository.list({ page: 1, limit: 5 });
      this.items = result.items;
      this.unreadCount = result.unread_count;
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

  async toggleRead(notification: SystemNotification): Promise<void> {
    try {
      const updated = await systemNotificationRepository.toggleRead(notification.id);
      this.items = this.items.map((item) => (item.id === updated.id ? updated : item));
      this.unreadCount += updated.read_at ? -1 : 1;
      if (this.unreadCount < 0) this.unreadCount = 0;
    } catch (error) {
      this.error = getErrorMessage(error);
    }
  }

  async markAllRead(): Promise<void> {
    try {
      await systemNotificationRepository.markAllRead();
      await this.loadNotifications();
    } catch (error) {
      this.error = getErrorMessage(error);
    }
  }

  async toggleImportant(notification: SystemNotification): Promise<void> {
    try {
      const updated = await systemNotificationRepository.toggleImportant(notification.id);
      this.items = this.items.map((item) => (item.id === updated.id ? updated : item));
    } catch (error) {
      this.error = getErrorMessage(error);
    }
  }

  async deleteNotification(notification: SystemNotification): Promise<void> {
    try {
      await systemNotificationRepository.delete(notification.id);
      await this.loadNotifications();
    } catch (error) {
      this.error = getErrorMessage(error);
    }
  }

  startPolling(): void {
    void this.loadNotifications();
    this.refreshTimer = setInterval(() => {
      void this.loadNotifications();
    }, 60000);
  }

  stopPolling(): void {
    if (this.refreshTimer) clearInterval(this.refreshTimer);
  }

  formatDateTime(value: string): string {
    return new Intl.DateTimeFormat('de-CH', {
      dateStyle: 'short',
      timeStyle: 'short'
    }).format(new Date(value));
  }
}
