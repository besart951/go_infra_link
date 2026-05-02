export type RealtimeSocketStatus = 'disconnected' | 'connecting' | 'connected';

export interface ReconnectingWebSocketOptions {
  url: () => string | null;
  onMessage: (data: string) => void;
  onOpen?: (event: { wasReconnect: boolean }) => void;
  onStatusChange?: (status: RealtimeSocketStatus) => void;
  reconnectBaseDelayMs?: number;
  reconnectMaxDelayMs?: number;
  maxQueuedMessages?: number;
}

export interface RealtimeSocketConnection {
  connect(): void;
  disconnect(options?: { clearQueue?: boolean }): void;
  send(payload: Record<string, unknown>, options?: { queueWhenClosed?: boolean }): void;
}

export class ReconnectingWebSocket implements RealtimeSocketConnection {
  private readonly url: ReconnectingWebSocketOptions['url'];
  private readonly onMessage: ReconnectingWebSocketOptions['onMessage'];
  private readonly onOpen?: ReconnectingWebSocketOptions['onOpen'];
  private readonly onStatusChange?: ReconnectingWebSocketOptions['onStatusChange'];
  private readonly reconnectBaseDelayMs: number;
  private readonly reconnectMaxDelayMs: number;
  private readonly maxQueuedMessages: number;

  private socket: WebSocket | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectDelayMs: number;
  private hasConnectedOnce = false;
  private destroyed = true;
  private queuedMessages: Array<Record<string, unknown>> = [];

  constructor(options: ReconnectingWebSocketOptions) {
    this.url = options.url;
    this.onMessage = options.onMessage;
    this.onOpen = options.onOpen;
    this.onStatusChange = options.onStatusChange;
    this.reconnectBaseDelayMs = options.reconnectBaseDelayMs ?? 2000;
    this.reconnectMaxDelayMs = options.reconnectMaxDelayMs ?? 10000;
    this.maxQueuedMessages = options.maxQueuedMessages ?? 50;
    this.reconnectDelayMs = this.reconnectBaseDelayMs;
  }

  connect(): void {
    if (typeof window === 'undefined') return;

    this.destroyed = false;
    if (!this.socket && !this.reconnectTimer) {
      this.openSocket();
    }
  }

  disconnect(options: { clearQueue?: boolean } = {}): void {
    this.destroyed = true;
    this.clearReconnectTimer();
    this.setStatus('disconnected');
    this.hasConnectedOnce = false;
    this.reconnectDelayMs = this.reconnectBaseDelayMs;
    if (options.clearQueue ?? true) {
      this.queuedMessages = [];
    }

    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }

  send(payload: Record<string, unknown>, options: { queueWhenClosed?: boolean } = {}): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      if (options.queueWhenClosed) {
        this.queueMessage(payload);
      }
      return;
    }

    this.socket.send(JSON.stringify(payload));
  }

  private openSocket(): void {
    if (this.destroyed || typeof window === 'undefined') return;

    const url = this.url();
    if (!url) return;

    this.setStatus('connecting');
    const socket = new WebSocket(url);
    this.socket = socket;

    socket.addEventListener('open', () => {
      if (this.socket !== socket) return;

      const wasReconnect = this.hasConnectedOnce;
      this.hasConnectedOnce = true;
      this.setStatus('connected');
      this.reconnectDelayMs = this.reconnectBaseDelayMs;
      this.onOpen?.({ wasReconnect });
      this.flushQueuedMessages();
    });

    socket.addEventListener('message', (event) => {
      if (this.socket !== socket) return;
      this.onMessage(String(event.data));
    });

    socket.addEventListener('close', () => {
      if (this.socket === socket) {
        this.socket = null;
      }
      if (this.destroyed) return;

      this.setStatus('disconnected');
      this.scheduleReconnect();
    });

    socket.addEventListener('error', () => {
      socket.close();
    });
  }

  private queueMessage(payload: Record<string, unknown>): void {
    this.queuedMessages = [...this.queuedMessages, payload].slice(-this.maxQueuedMessages);
  }

  private flushQueuedMessages(): void {
    if (this.queuedMessages.length === 0) return;

    const messages = this.queuedMessages;
    this.queuedMessages = [];
    for (const message of messages) {
      this.send(message);
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer || this.destroyed) return;

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.openSocket();
    }, this.reconnectDelayMs);
    this.reconnectDelayMs = Math.min(this.reconnectDelayMs * 2, this.reconnectMaxDelayMs);
  }

  private clearReconnectTimer(): void {
    if (!this.reconnectTimer) return;

    clearTimeout(this.reconnectTimer);
    this.reconnectTimer = null;
  }

  private setStatus(status: RealtimeSocketStatus): void {
    this.onStatusChange?.(status);
  }
}
