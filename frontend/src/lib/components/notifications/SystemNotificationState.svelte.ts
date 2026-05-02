import { getErrorMessage } from '$lib/api/client.js';
import type {
  SystemNotification,
  SystemNotificationStreamEvent
} from '$lib/domain/notification/index.js';
import { reduceSystemNotificationStreamEvent } from '$lib/domain/notification/index.js';
import type { SystemNotificationRepository } from '$lib/domain/ports/notification/systemNotificationRepository.js';
import { systemNotificationRepository } from '$lib/infrastructure/api/systemNotificationRepository.js';
import type { RealtimeSocketStatus } from '$lib/infrastructure/realtime/reconnectingWebSocket.js';
import {
  buildSameOriginWebSocketUrl,
  RealtimeJsonStream
} from '$lib/infrastructure/realtime/realtimeJsonStream.js';

const PREVIEW_LIMIT = 5;
const INBOX_LIMIT = 20;

interface SystemNotificationStateDeps {
  repository?: SystemNotificationRepository;
  stream?: RealtimeJsonStream<SystemNotificationStreamEvent>;
}

export class SystemNotificationState {
  previewItems = $state<SystemNotification[]>([]);
  inboxItems = $state<SystemNotification[]>([]);
  unreadCount = $state(0);
  inboxPage = $state(1);
  inboxTotalPages = $state(1);
  inboxUnreadOnly = $state(false);
  isPreviewLoading = $state(false);
  isInboxLoading = $state(true);
  previewError = $state<string | null>(null);
  inboxError = $state<string | null>(null);
  socketStatus = $state<RealtimeSocketStatus>('disconnected');

  private readonly repository: SystemNotificationRepository;
  private readonly stream: RealtimeJsonStream<SystemNotificationStreamEvent>;
  private connectionRefs = 0;
  private inboxLoaded = false;

  constructor(deps: SystemNotificationStateDeps = {}) {
    this.repository = deps.repository ?? systemNotificationRepository;
    this.stream =
      deps.stream ??
      new RealtimeJsonStream<SystemNotificationStreamEvent>({
        url: buildSystemNotificationStreamUrl,
        onMessage: (event) => this.handleMessage(event),
        onOpen: () => void this.refreshVisible(),
        onStatusChange: (status) => {
          this.socketStatus = status;
        }
      });
  }

  connect(): void {
    if (typeof window === 'undefined') return;

    this.connectionRefs += 1;
    this.stream.connect();
  }

  disconnect(): void {
    this.connectionRefs = Math.max(0, this.connectionRefs - 1);
    if (this.connectionRefs > 0) return;

    this.inboxLoaded = false;
    this.stream.disconnect();
  }

  async loadPreview(): Promise<void> {
    this.isPreviewLoading = true;
    this.previewError = null;
    try {
      const result = await this.repository.list({ page: 1, limit: PREVIEW_LIMIT });
      this.previewItems = result.items;
      this.unreadCount = result.unread_count;
    } catch (error) {
      this.previewError = getErrorMessage(error);
    } finally {
      this.isPreviewLoading = false;
    }
  }

  async loadInbox(nextPage = this.inboxPage): Promise<void> {
    this.inboxLoaded = true;
    this.isInboxLoading = true;
    this.inboxError = null;
    try {
      const result = await this.repository.list({
        page: nextPage,
        limit: INBOX_LIMIT,
        unread_only: this.inboxUnreadOnly
      });
      this.inboxItems = result.items;
      this.unreadCount = result.unread_count;
      this.inboxPage = result.page;
      this.inboxTotalPages = result.total_pages || 1;
    } catch (error) {
      this.inboxError = getErrorMessage(error);
    } finally {
      this.isInboxLoading = false;
    }
  }

  async markRead(notification: SystemNotification): Promise<void> {
    if (notification.read_at) return;
    try {
      await this.repository.markRead(notification.id);
      await this.refreshVisible();
    } catch (error) {
      this.previewError = getErrorMessage(error);
      this.inboxError = getErrorMessage(error);
    }
  }

  async toggleRead(notification: SystemNotification): Promise<void> {
    try {
      await this.repository.toggleRead(notification.id);
      await this.refreshVisible();
    } catch (error) {
      this.previewError = getErrorMessage(error);
      this.inboxError = getErrorMessage(error);
    }
  }

  async toggleImportant(notification: SystemNotification): Promise<void> {
    try {
      await this.repository.toggleImportant(notification.id);
      await this.refreshVisible();
    } catch (error) {
      this.previewError = getErrorMessage(error);
      this.inboxError = getErrorMessage(error);
    }
  }

  async deleteNotification(notification: SystemNotification): Promise<void> {
    const nextInboxPage =
      this.inboxItems.length === 1 && this.inboxPage > 1 ? this.inboxPage - 1 : this.inboxPage;
    try {
      await this.repository.delete(notification.id);
      await this.refreshVisible(nextInboxPage);
    } catch (error) {
      this.previewError = getErrorMessage(error);
      this.inboxError = getErrorMessage(error);
    }
  }

  async markAllRead(): Promise<void> {
    try {
      await this.repository.markAllRead();
      await this.refreshVisible(1);
    } catch (error) {
      this.previewError = getErrorMessage(error);
      this.inboxError = getErrorMessage(error);
    }
  }

  toggleInboxUnreadOnly(): void {
    this.inboxUnreadOnly = !this.inboxUnreadOnly;
    void this.loadInbox(1);
  }

  deactivateInbox(): void {
    this.inboxLoaded = false;
  }

  formatDateTime(value: string): string {
    return new Intl.DateTimeFormat('de-CH', {
      dateStyle: 'short',
      timeStyle: 'short'
    }).format(new Date(value));
  }

  formatInboxDateTime(value: string): string {
    return new Intl.DateTimeFormat('de-CH', {
      dateStyle: 'medium',
      timeStyle: 'short'
    }).format(new Date(value));
  }

  private handleMessage(event: SystemNotificationStreamEvent): void {
    const next = reduceSystemNotificationStreamEvent(
      {
        previewItems: this.previewItems,
        inboxItems: this.inboxItems,
        unreadCount: this.unreadCount,
        inboxUnreadOnly: this.inboxUnreadOnly,
        inboxLoaded: this.inboxLoaded
      },
      event,
      PREVIEW_LIMIT
    );
    this.previewItems = next.previewItems;
    this.inboxItems = next.inboxItems;
    this.unreadCount = next.unreadCount;
    if (next.reloadInboxPage) {
      void this.loadInbox(next.reloadInboxPage === 'first' ? 1 : this.inboxPage);
    }
  }

  private async refreshVisible(nextInboxPage = this.inboxPage): Promise<void> {
    const tasks: Promise<void>[] = [this.loadPreview()];
    if (this.inboxLoaded) {
      tasks.push(this.loadInbox(nextInboxPage));
    }
    await Promise.all(tasks);
  }
}

function buildSystemNotificationStreamUrl(): string | null {
  return buildSameOriginWebSocketUrl('/api/v1/account/notifications/stream');
}

export const systemNotificationState = new SystemNotificationState();
