package realtime

import (
	"context"
	"sync"

	apprealtime "github.com/besart951/go_infra_link/backend/internal/application/realtime"
)

const defaultRealtimeBusSubscriberBuffer = 64

type InMemoryBusConfig struct {
	SubscriberBuffer int
}

type InMemoryBus struct {
	mu          sync.RWMutex
	buffer      int
	closed      bool
	subscribers map[*inMemoryBusSubscription]struct{}
	ctx         context.Context
	cancel      context.CancelFunc
	closeOnce   sync.Once
}

type inMemoryBusSubscription struct {
	topics map[apprealtime.Topic]struct{}
	events chan apprealtime.Event
}

func NewInMemoryBus(config ...InMemoryBusConfig) *InMemoryBus {
	cfg := InMemoryBusConfig{SubscriberBuffer: defaultRealtimeBusSubscriberBuffer}
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.SubscriberBuffer <= 0 {
		cfg.SubscriberBuffer = defaultRealtimeBusSubscriberBuffer
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &InMemoryBus{
		buffer:      cfg.SubscriberBuffer,
		subscribers: make(map[*inMemoryBusSubscription]struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (b *InMemoryBus) Publish(ctx context.Context, event apprealtime.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	event = apprealtime.NormalizeEvent(event)

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
		case <-ctx.Done():
			return ctx.Err()
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

func (b *InMemoryBus) Subscribe(ctx context.Context, topics ...apprealtime.Topic) (<-chan apprealtime.Event, error) {
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

	sub := &inMemoryBusSubscription{
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

func (b *InMemoryBus) Close() error {
	b.closeOnce.Do(func() {
		b.cancel()
		b.close()
	})
	return nil
}

func (b *InMemoryBus) close() {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	b.closed = true
	subscribers := make([]*inMemoryBusSubscription, 0, len(b.subscribers))
	for sub := range b.subscribers {
		subscribers = append(subscribers, sub)
		delete(b.subscribers, sub)
	}
	b.mu.Unlock()

	for _, sub := range subscribers {
		close(sub.events)
	}
}

func (b *InMemoryBus) removeSubscription(sub *inMemoryBusSubscription) {
	b.mu.Lock()
	if _, ok := b.subscribers[sub]; !ok {
		b.mu.Unlock()
		return
	}
	delete(b.subscribers, sub)
	b.mu.Unlock()

	close(sub.events)
}
