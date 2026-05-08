package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	apprealtime "github.com/besart951/go_infra_link/backend/internal/application/realtime"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultPostgresRealtimeChannel = "go_infra_link_realtime"
	defaultPostgresRealtimeTTL     = 10 * time.Minute
	postgresRealtimeFetchTimeout   = 5 * time.Second
	postgresRealtimeReconnectDelay = time.Second
	postgresRealtimeCleanupPeriod  = time.Minute
)

type PostgresBusConfig struct {
	DSN              string
	Channel          string
	SubscriberBuffer int
	EventTTL         time.Duration
}

type PostgresBus struct {
	dsn     string
	channel string
	buffer  int
	ttl     time.Duration
	pool    *pgxpool.Pool

	mu          sync.RWMutex
	closed      bool
	subscribers map[*postgresBusSubscription]struct{}

	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
	wg        sync.WaitGroup
}

type postgresBusSubscription struct {
	topics map[apprealtime.Topic]struct{}
	events chan apprealtime.Event
}

func NewPostgresBus(ctx context.Context, config PostgresBusConfig) (*PostgresBus, error) {
	channel := normalizePostgresRealtimeChannel(config.Channel)
	buffer := config.SubscriberBuffer
	if buffer <= 0 {
		buffer = defaultRealtimeBusSubscriberBuffer
	}
	ttl := config.EventTTL
	if ttl <= 0 {
		ttl = defaultPostgresRealtimeTTL
	}
	if strings.TrimSpace(config.DSN) == "" {
		return nil, fmt.Errorf("postgres realtime bus dsn is required")
	}

	pool, err := pgxpool.New(ctx, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("postgres realtime bus pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres realtime bus ping: %w", err)
	}
	if err := ensurePostgresBusSchema(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	lifetimeCtx, cancel := context.WithCancel(context.Background())
	bus := &PostgresBus{
		dsn:         config.DSN,
		channel:     channel,
		buffer:      buffer,
		ttl:         ttl,
		pool:        pool,
		subscribers: make(map[*postgresBusSubscription]struct{}),
		ctx:         lifetimeCtx,
		cancel:      cancel,
	}

	bus.wg.Add(2)
	go bus.listenLoop()
	go bus.cleanupLoop()

	return bus, nil
}

func (b *PostgresBus) Publish(ctx context.Context, event apprealtime.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	event = apprealtime.NormalizeEvent(event)
	payload := string(event.Payload)
	expiresAt := event.PublishedAt.Add(b.ttl)

	tx, err := b.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres realtime publish begin: %w", err)
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	if _, err := tx.Exec(ctx, `
		insert into realtime_events (id, topic, source, payload, published_at, expires_at)
		values ($1, $2, $3, $4::jsonb, $5, $6)
	`, event.ID, string(event.Topic), event.Source, payload, event.PublishedAt, expiresAt); err != nil {
		return fmt.Errorf("postgres realtime publish insert: %w", err)
	}

	if _, err := tx.Exec(ctx, `select pg_notify($1, $2)`, b.channel, event.ID); err != nil {
		return fmt.Errorf("postgres realtime publish notify: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres realtime publish commit: %w", err)
	}
	return nil
}

func (b *PostgresBus) Subscribe(ctx context.Context, topics ...apprealtime.Topic) (<-chan apprealtime.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	topicSet := make(map[apprealtime.Topic]struct{}, len(topics))
	for _, topic := range topics {
		if topic != "" {
			topicSet[topic] = struct{}{}
		}
	}
	if len(topicSet) == 0 {
		return nil, apprealtime.ErrNoTopics
	}

	sub := &postgresBusSubscription{
		topics: topicSet,
		events: make(chan apprealtime.Event, b.buffer),
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		close(sub.events)
		return nil, apprealtime.ErrClosed
	}
	b.subscribers[sub] = struct{}{}
	b.mu.Unlock()

	go func() {
		select {
		case <-ctx.Done():
		case <-b.ctx.Done():
		}
		b.removeSubscription(sub)
	}()

	return sub.events, nil
}

func (b *PostgresBus) Close() error {
	b.closeOnce.Do(func() {
		b.cancel()
		b.wg.Wait()
		b.pool.Close()

		b.mu.Lock()
		b.closed = true
		subscribers := make([]*postgresBusSubscription, 0, len(b.subscribers))
		for sub := range b.subscribers {
			subscribers = append(subscribers, sub)
			delete(b.subscribers, sub)
		}
		b.mu.Unlock()

		for _, sub := range subscribers {
			close(sub.events)
		}
	})
	return nil
}

func (b *PostgresBus) listenLoop() {
	defer b.wg.Done()
	for b.ctx.Err() == nil {
		if err := b.listenOnce(); err != nil && b.ctx.Err() == nil {
			slog.Warn("postgres realtime listener reconnecting", "err", err)
			select {
			case <-b.ctx.Done():
			case <-time.After(postgresRealtimeReconnectDelay):
			}
		}
	}
}

func (b *PostgresBus) listenOnce() error {
	conn, err := pgx.Connect(b.ctx, b.dsn)
	if err != nil {
		return fmt.Errorf("connect listener: %w", err)
	}
	defer func() {
		_ = conn.Close(context.Background())
	}()

	if _, err := conn.Exec(b.ctx, "listen "+pgx.Identifier{b.channel}.Sanitize()); err != nil {
		return fmt.Errorf("listen channel: %w", err)
	}

	for b.ctx.Err() == nil {
		notification, err := conn.WaitForNotification(b.ctx)
		if err != nil {
			return fmt.Errorf("wait notification: %w", err)
		}
		if notification == nil || notification.Channel != b.channel {
			continue
		}
		if err := b.loadAndDispatch(strings.TrimSpace(notification.Payload)); err != nil {
			slog.Warn("postgres realtime notification ignored", "err", err)
		}
	}
	return nil
}

func (b *PostgresBus) loadAndDispatch(id string) error {
	if id == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(b.ctx, postgresRealtimeFetchTimeout)
	defer cancel()

	var (
		event       apprealtime.Event
		topic       string
		payloadText string
	)
	if err := b.pool.QueryRow(ctx, `
		select id::text, topic, source, payload::text, published_at
		from realtime_events
		where id = $1 and expires_at > now()
	`, id).Scan(&event.ID, &topic, &event.Source, &payloadText, &event.PublishedAt); err != nil {
		return fmt.Errorf("load event %s: %w", id, err)
	}
	event.Topic = apprealtime.Topic(topic)
	event.Payload = json.RawMessage([]byte(payloadText))
	return b.dispatch(event)
}

func (b *PostgresBus) dispatch(event apprealtime.Event) error {
	var dropped bool
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return apprealtime.ErrClosed
	}
	for sub := range b.subscribers {
		if _, ok := sub.topics[event.Topic]; !ok {
			continue
		}
		select {
		case sub.events <- event:
		default:
			dropped = true
		}
	}
	if dropped {
		return apprealtime.ErrBackpressure
	}
	return nil
}

func (b *PostgresBus) cleanupLoop() {
	defer b.wg.Done()
	ticker := time.NewTicker(postgresRealtimeCleanupPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(b.ctx, postgresRealtimeFetchTimeout)
			_, err := b.pool.Exec(ctx, `delete from realtime_events where expires_at < now()`)
			cancel()
			if err != nil && b.ctx.Err() == nil {
				slog.Warn("postgres realtime cleanup failed", "err", err)
			}
		}
	}
}

func (b *PostgresBus) removeSubscription(sub *postgresBusSubscription) {
	b.mu.Lock()
	if _, ok := b.subscribers[sub]; !ok {
		b.mu.Unlock()
		return
	}
	delete(b.subscribers, sub)
	b.mu.Unlock()

	close(sub.events)
}

func ensurePostgresBusSchema(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `
		create table if not exists realtime_events (
			id uuid primary key,
			topic text not null,
			source text not null,
			payload jsonb not null,
			published_at timestamptz not null,
			expires_at timestamptz not null
		)
	`); err != nil {
		return fmt.Errorf("postgres realtime schema table: %w", err)
	}
	if _, err := pool.Exec(ctx, `
		create index if not exists realtime_events_expires_at_idx
		on realtime_events (expires_at)
	`); err != nil {
		return fmt.Errorf("postgres realtime schema index: %w", err)
	}
	return nil
}

func normalizePostgresRealtimeChannel(channel string) string {
	channel = strings.TrimSpace(channel)
	if isSafePostgresRealtimeChannel(channel) {
		return channel
	}
	return defaultPostgresRealtimeChannel
}

func isSafePostgresRealtimeChannel(channel string) bool {
	if channel == "" || len(channel) > 63 {
		return false
	}
	for index, r := range channel {
		switch {
		case r == '_':
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case index > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}
