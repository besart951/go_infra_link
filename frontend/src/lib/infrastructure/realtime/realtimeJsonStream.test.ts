import { describe, expect, it, vi } from 'vitest';
import type {
  ReconnectingWebSocketOptions,
  RealtimeSocketConnection
} from './reconnectingWebSocket.js';
import { buildSameOriginWebSocketUrl, RealtimeJsonStream } from './realtimeJsonStream.js';

class FakeConnection implements RealtimeSocketConnection {
  connected = false;
  disconnectedWith: { clearQueue?: boolean } | undefined;
  sent: Array<{
    payload: Record<string, unknown>;
    options?: { queueWhenClosed?: boolean };
  }> = [];

  constructor(readonly options: ReconnectingWebSocketOptions) {}

  connect(): void {
    this.connected = true;
  }

  disconnect(options?: { clearQueue?: boolean }): void {
    this.disconnectedWith = options;
  }

  send(payload: Record<string, unknown>, options?: { queueWhenClosed?: boolean }): void {
    this.sent.push({ payload, options });
  }
}

describe('RealtimeJsonStream', () => {
  it('parses inbound JSON before crossing the stream interface', () => {
    let connection: FakeConnection | undefined;
    const onMessage = vi.fn();
    const stream = new RealtimeJsonStream<{ type: string }>({
      url: () => '/stream',
      onMessage,
      createConnection: (options) => {
        connection = new FakeConnection(options);
        return connection;
      }
    });

    connection?.options.onMessage('{"type":"created"}');

    expect(onMessage).toHaveBeenCalledWith({ type: 'created' });
    expect(stream).toBeDefined();
  });

  it('keeps invalid JSON behind the stream interface', () => {
    let connection: FakeConnection | undefined;
    const onMessage = vi.fn();
    const onInvalidMessage = vi.fn();
    new RealtimeJsonStream<{ type: string }>({
      url: () => '/stream',
      onMessage,
      onInvalidMessage,
      createConnection: (options) => {
        connection = new FakeConnection(options);
        return connection;
      }
    });

    connection?.options.onMessage('not json');

    expect(onMessage).not.toHaveBeenCalled();
    expect(onInvalidMessage).toHaveBeenCalledWith('not json');
  });

  it('delegates lifecycle and outbound messages to the socket connection', () => {
    let connection: FakeConnection | undefined;
    const stream = new RealtimeJsonStream<{ type: string }>({
      url: () => '/stream',
      onMessage: vi.fn(),
      createConnection: (options) => {
        connection = new FakeConnection(options);
        return connection;
      }
    });

    stream.connect();
    stream.send({ type: 'refresh' }, { queueWhenClosed: true });
    stream.disconnect({ clearQueue: false });

    expect(connection?.connected).toBe(true);
    expect(connection?.sent).toEqual([
      { payload: { type: 'refresh' }, options: { queueWhenClosed: true } }
    ]);
    expect(connection?.disconnectedWith).toEqual({ clearQueue: false });
  });

  it('builds same-origin websocket URLs', () => {
    expect(buildSameOriginWebSocketUrl('api/v1/stream')).toBe(
      `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/stream`
    );
  });
});
