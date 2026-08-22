package wire

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	facilityjobs "github.com/besart951/go_infra_link/backend/internal/application/facilityjobs"
	apprealtime "github.com/besart951/go_infra_link/backend/internal/application/realtime"
	"github.com/besart951/go_infra_link/backend/internal/infrastructure/realtime"
	"github.com/besart951/go_infra_link/backend/internal/repository/facilityjobsql"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RuntimeAdapters struct {
	ProjectCollaboration     *realtime.ProjectCollaborationHub
	SystemNotificationStream *realtime.SystemNotificationHub
	FacilityReferenceData    *realtime.FacilityReferenceDataHub
	FacilityJobs             *facilityservice.FacilityJobManager
	FacilityJobSteps         facilityjobs.StepStore
	FieldDeviceUpdatePlans   facilityjobs.FieldDeviceUpdatePlanStore
	DB                       *gorm.DB
	bus                      apprealtime.Bus
	ownsBus                  bool
	closeOnce                sync.Once
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
	DB               *gorm.DB
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
	adapters := NewRuntimeAdaptersWithBusAndStore(bus, nodeID, true, cfg.DB)
	return adapters, adapters.Close, nil
}

func NewRuntimeAdaptersWithBus(bus apprealtime.Bus, nodeID string, ownsBus bool) *RuntimeAdapters {
	return NewRuntimeAdaptersWithBusAndStore(bus, nodeID, ownsBus, nil)
}

func NewRuntimeAdaptersWithBusAndStore(bus apprealtime.Bus, nodeID string, ownsBus bool, db *gorm.DB) *RuntimeAdapters {
	if strings.TrimSpace(nodeID) == "" {
		nodeID = uuid.NewString()
	}
	options := []realtime.ProjectCollaborationHubOption{realtime.WithProjectCollaborationBus(bus, nodeID)}
	if db != nil {
		store := realtime.NewSQLProjectCollaborationStore(db)
		options = append(options, realtime.WithProjectRevisionSource(store), realtime.WithProjectDraftStore(store), realtime.WithProjectPresenceStore(store))
	}
	facilityReferenceData := realtime.NewFacilityReferenceDataHub(
		realtime.WithFacilityReferenceDataBus(bus, nodeID),
	)
	var jobSteps facilityjobs.StepStore
	var updatePlans facilityjobs.FieldDeviceUpdatePlanStore
	if db != nil {
		jobSteps = facilityjobsql.NewStepStore(db)
		updatePlans = facilityjobsql.NewFieldDeviceUpdatePlanStore(db)
	}
	return &RuntimeAdapters{
		ProjectCollaboration: realtime.NewProjectCollaborationHub(options...),
		SystemNotificationStream: realtime.NewSystemNotificationHub(
			realtime.WithSystemNotificationBus(bus, nodeID),
		),
		FacilityReferenceData:  facilityReferenceData,
		FacilityJobs:           facilityservice.NewFacilityJobManagerWithDB(facilityReferenceData, db),
		FacilityJobSteps:       jobSteps,
		FieldDeviceUpdatePlans: updatePlans,
		DB:                     db,
		bus:                    bus,
		ownsBus:                ownsBus,
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
	r.closeOnce.Do(func() {
		if r.FacilityJobs != nil {
			r.FacilityJobs.Close()
		}
		if r.ProjectCollaboration != nil {
			r.ProjectCollaboration.Close()
		}
		if r.SystemNotificationStream != nil {
			r.SystemNotificationStream.Close()
		}
		if r.FacilityReferenceData != nil {
			r.FacilityReferenceData.Close()
		}
		if r.ownsBus && r.bus != nil {
			_ = r.bus.Close()
		}
	})
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
