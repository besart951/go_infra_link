package collaboration

import (
	"context"
	"fmt"
	"time"

	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
)

// OutboxConsumer is intentionally idempotent by EventID. A process can fail
// after handing an event to a websocket/bus but before recording delivery;
// replaying that event is therefore expected, never exceptional.
type OutboxConsumer interface {
	ConsumerID() string
	Deliver(context.Context, domainCollaboration.OutboxEvent) error
}

type RetryDelay func(attempt int) time.Duration

type OutboxProcessor struct {
	store      domainCollaboration.OutboxStore
	consumer   OutboxConsumer
	retryDelay RetryDelay
}

func NewOutboxProcessor(store domainCollaboration.OutboxStore, consumer OutboxConsumer, retryDelay RetryDelay) (*OutboxProcessor, error) {
	if store == nil {
		return nil, fmt.Errorf("collaboration outbox store is required")
	}
	if consumer == nil || consumer.ConsumerID() == "" {
		return nil, fmt.Errorf("collaboration outbox consumer is required")
	}
	if retryDelay == nil {
		retryDelay = func(attempt int) time.Duration {
			if attempt < 1 {
				attempt = 1
			}
			return time.Second * time.Duration(1<<(attempt-1))
		}
	}
	return &OutboxProcessor{store: store, consumer: consumer, retryDelay: retryDelay}, nil
}

// RunOnce claims a bounded batch. Delivery failures are persisted for retry and
// block only later events for the same project during this pass.
func (p *OutboxProcessor) RunOnce(ctx context.Context, now time.Time, limit int) (int, error) {
	if p == nil {
		return 0, fmt.Errorf("collaboration outbox processor is not configured")
	}
	events, err := p.store.ClaimDue(ctx, now.UTC(), limit)
	if err != nil {
		return 0, fmt.Errorf("claim collaboration outbox events: %w", err)
	}
	var firstError error
	blockedProjects := make(map[string]struct{})
	for _, event := range events {
		projectKey := event.ProjectID.String()
		if _, blocked := blockedProjects[projectKey]; blocked {
			continue
		}
		processed, err := p.store.WasProcessed(ctx, p.consumer.ConsumerID(), event.EventID)
		if err != nil {
			if firstError == nil {
				firstError = fmt.Errorf("check collaboration event %s idempotency: %w", event.EventID, err)
			}
			blockedProjects[projectKey] = struct{}{}
			continue
		}
		if !processed {
			if err := p.consumer.Deliver(ctx, event); err != nil {
				if markErr := p.store.MarkFailed(ctx, event, err.Error(), now.UTC(), now.UTC().Add(p.retryDelay(event.Attempts))); markErr != nil && firstError == nil {
					firstError = fmt.Errorf("record collaboration delivery failure: %w", markErr)
				}
				blockedProjects[projectKey] = struct{}{}
				continue
			}
		}
		if err := p.store.MarkDelivered(ctx, p.consumer.ConsumerID(), event, now.UTC()); err != nil {
			if firstError == nil {
				firstError = fmt.Errorf("record collaboration delivery: %w", err)
			}
			blockedProjects[projectKey] = struct{}{}
		}
	}
	return len(events), firstError
}

// StartWorker polls until the returned stop function is called. One immediate
// pass drains events left by the previous process before the first interval.
func (p *OutboxProcessor) StartWorker(
	interval time.Duration,
	limit int,
	reportError func(error),
) func() {
	if interval <= 0 {
		interval = time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		run := func() {
			if _, err := p.RunOnce(ctx, time.Now().UTC(), limit); err != nil &&
				ctx.Err() == nil && reportError != nil {
				reportError(err)
			}
		}
		run()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
	return func() {
		cancel()
		<-done
	}
}
