package wire

import (
	"context"
	"fmt"
	"strings"
	"time"

	apprealtime "github.com/besart951/go_infra_link/backend/internal/application/realtime"
	"github.com/besart951/go_infra_link/backend/internal/infrastructure/realtime"
	"github.com/google/uuid"
)

type RuntimeAdapters struct {
	ProjectCollaboration     *realtime.ProjectCollaborationHub
	SystemNotificationStream *realtime.SystemNotificationHub
	bus                      apprealtime.Bus
	ownsBus                  bool
}

func NewRuntimeAdapters() *RuntimeAdapters {
	bus := realtime.NewInMemoryBus()
	return NewRuntimeAdaptersWithBus(bus, uuid.NewString(), true)
}

type RuntimeConfig struct {
	Bus              string
	NodeID           string
	PostgresDSN      string
	PostgresChannel  string
	SubscriberBuffer int
	EventTTL         time.Duration
}

const (
	RuntimeBusMemory   = "memory"
	RuntimeBusPostgres = "postgres"
)

func NewRuntimeAdaptersFromConfig(ctx context.Context, cfg RuntimeConfig) (*RuntimeAdapters, func(), error) {
	bus, err := newRuntimeBus(ctx, cfg)
	if err != nil {
		return nil, func() {}, err
	}
	nodeID := strings.TrimSpace(cfg.NodeID)
	if nodeID == "" {
		nodeID = uuid.NewString()
	}
	adapters := NewRuntimeAdaptersWithBus(bus, nodeID, true)
	return adapters, adapters.Close, nil
}

func NewRuntimeAdaptersWithBus(bus apprealtime.Bus, nodeID string, ownsBus bool) *RuntimeAdapters {
	if strings.TrimSpace(nodeID) == "" {
		nodeID = uuid.NewString()
	}
	return &RuntimeAdapters{
		ProjectCollaboration: realtime.NewProjectCollaborationHub(
			realtime.WithProjectCollaborationBus(bus, nodeID),
		),
		SystemNotificationStream: realtime.NewSystemNotificationHub(
			realtime.WithSystemNotificationBus(bus, nodeID),
		),
		bus:     bus,
		ownsBus: ownsBus,
	}
}

func runtimeOrDefault(runtime *RuntimeAdapters) *RuntimeAdapters {
	if runtime == nil {
		return NewRuntimeAdapters()
	}
	return runtime
}

func runtimeOrNil(runtime *RuntimeAdapters) *RuntimeAdapters {
	return runtime
}

func (r *RuntimeAdapters) Close() {
	if r == nil {
		return
	}
	if r.ProjectCollaboration != nil {
		r.ProjectCollaboration.Close()
	}
	if r.SystemNotificationStream != nil {
		r.SystemNotificationStream.Close()
	}
	if r.ownsBus && r.bus != nil {
		_ = r.bus.Close()
	}
}

func newRuntimeBus(ctx context.Context, cfg RuntimeConfig) (apprealtime.Bus, error) {
	switch strings.TrimSpace(strings.ToLower(cfg.Bus)) {
	case "", RuntimeBusMemory:
		return realtime.NewInMemoryBus(realtime.InMemoryBusConfig{
			SubscriberBuffer: cfg.SubscriberBuffer,
		}), nil
	case RuntimeBusPostgres:
		return realtime.NewPostgresBus(ctx, realtime.PostgresBusConfig{
			DSN:              cfg.PostgresDSN,
			Channel:          cfg.PostgresChannel,
			SubscriberBuffer: cfg.SubscriberBuffer,
			EventTTL:         cfg.EventTTL,
		})
	default:
		return nil, fmt.Errorf("unsupported realtime bus %q: expected %q or %q", cfg.Bus, RuntimeBusMemory, RuntimeBusPostgres)
	}
}
