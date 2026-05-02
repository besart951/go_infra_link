import {
  ReconnectingWebSocket,
  type ReconnectingWebSocketOptions,
  type RealtimeSocketConnection,
  type RealtimeSocketStatus
} from './reconnectingWebSocket.js';

export interface RealtimeJsonStreamOptions<TMessage> {
  url: () => string | null;
  onMessage: (message: TMessage) => void;
  onInvalidMessage?: (raw: string) => void;
  onOpen?: (event: { wasReconnect: boolean }) => void;
  onStatusChange?: (status: RealtimeSocketStatus) => void;
  reconnectBaseDelayMs?: number;
  reconnectMaxDelayMs?: number;
  maxQueuedMessages?: number;
  createConnection?: (options: ReconnectingWebSocketOptions) => RealtimeSocketConnection;
}

export class RealtimeJsonStream<TMessage> implements RealtimeSocketConnection {
  private readonly connection: RealtimeSocketConnection;
  private readonly onMessage: (message: TMessage) => void;
  private readonly onInvalidMessage?: (raw: string) => void;

  constructor(options: RealtimeJsonStreamOptions<TMessage>) {
    this.onMessage = options.onMessage;
    this.onInvalidMessage = options.onInvalidMessage;

    const createConnection =
      options.createConnection ??
      ((connectionOptions) => new ReconnectingWebSocket(connectionOptions));

    this.connection = createConnection({
      url: options.url,
      onMessage: (raw) => this.handleRawMessage(raw),
      onOpen: options.onOpen,
      onStatusChange: options.onStatusChange,
      reconnectBaseDelayMs: options.reconnectBaseDelayMs,
      reconnectMaxDelayMs: options.reconnectMaxDelayMs,
      maxQueuedMessages: options.maxQueuedMessages
    });
  }

  connect(): void {
    this.connection.connect();
  }

  disconnect(options: { clearQueue?: boolean } = {}): void {
    this.connection.disconnect(options);
  }

  send(payload: Record<string, unknown>, options: { queueWhenClosed?: boolean } = {}): void {
    this.connection.send(payload, options);
  }

  private handleRawMessage(raw: string): void {
    try {
      this.onMessage(JSON.parse(raw) as TMessage);
    } catch {
      this.onInvalidMessage?.(raw);
    }
  }
}

export function buildSameOriginWebSocketUrl(path: string): string | null {
  if (typeof window === 'undefined') return null;

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  return `${protocol}//${window.location.host}${normalizedPath}`;
}
