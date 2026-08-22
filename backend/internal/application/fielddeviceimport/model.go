package fielddeviceimport

import (
	"context"
	"io"
	"time"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

const SchemaVersion = 2

type Manifest struct {
	SchemaVersion int
	SnapshotAt    time.Time
	Scope         string
	DeviceCount   int64
	Counts        Counts
}

type Counts struct {
	Specifications     int64
	BacnetObjects      int64
	SoftwareReferences int64
	AlarmValues        int64
}

type SoftwareReference struct {
	SourceObjectID uuid.UUID `json:"source_object_id"`
	TargetObjectID uuid.UUID `json:"target_object_id"`
	FieldDeviceID  uuid.UUID `json:"field_device_id"`
}

type Aggregate struct {
	FieldDevice        domainFacility.FieldDevice
	Specification      *domainFacility.Specification
	BacnetObjects      []domainFacility.BacnetObject
	SoftwareReferences []SoftwareReference
}

type Issue struct {
	Code     string    `json:"code"`
	Entity   string    `json:"entity"`
	SourceID uuid.UUID `json:"source_id,omitempty"`
	Field    string    `json:"field,omitempty"`
	Message  string    `json:"message"`
}

type Result struct {
	ImportID uuid.UUID `json:"import_id"`
	Total    int64     `json:"total"`
	Imported int64     `json:"imported"`
	Failed   int64     `json:"failed"`
	Issues   []Issue   `json:"issues,omitempty"`
}

type Command struct {
	OwnerID uuid.UUID
	Source  io.Reader
}

type WorkbookReader interface {
	Read(ctx context.Context, source io.Reader, sink Sink) (Manifest, error)
}

type Sink interface {
	FieldDevices(ctx context.Context, values []domainFacility.FieldDevice) error
	Specifications(ctx context.Context, values []domainFacility.Specification) error
	BacnetObjects(ctx context.Context, values []domainFacility.BacnetObject) error
	SoftwareReferences(ctx context.Context, values []SoftwareReference) error
	AlarmValues(ctx context.Context, values []domainFacility.BacnetObjectAlarmValue) error
}

type AggregatePage struct {
	Items      []Aggregate
	NextCursor string
}

type Session interface {
	Sink
	Seal(ctx context.Context, manifest Manifest) error
	Validate(ctx context.Context) ([]Issue, error)
	Aggregates(ctx context.Context, cursor string) (AggregatePage, error)
	Complete(ctx context.Context) error
	Discard(ctx context.Context) error
}

type StagingStore interface {
	Start(ctx context.Context, ownerID uuid.UUID) (uuid.UUID, Session, error)
}

type AggregateWriter interface {
	Import(ctx context.Context, aggregate Aggregate) error
}
