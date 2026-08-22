package exporting

import (
	"context"
	"time"

	domainuser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type AccessScope string

const (
	AccessScopeGlobal  AccessScope = "global"
	AccessScopeProject AccessScope = "project"
)

type Request struct {
	ProjectIDs                 []uuid.UUID
	BuildingIDs                []uuid.UUID
	ControlCabinetIDs          []uuid.UUID
	SPSControllerIDs           []uuid.UUID
	SPSControllerSystemTypeIDs []uuid.UUID
	Search                     string
	ExportAll                  bool
	ForceAsync                 bool
	SnapshotAt                 time.Time
	SchemaVersion              int
	DeviceCount                int64
	AccessScope                AccessScope
	Manifest                   Manifest
}

type Manifest struct {
	Counts            Counts
	Warnings          []string
	WorkbookShards    []string
	SnapshotChecksums map[string]string
}

type Counts struct {
	FieldDevices       int64
	Specifications     int64
	BacnetObjects      int64
	SoftwareReferences int64
	AlarmValues        int64
}

func (c *Counts) Add(other Counts) {
	c.FieldDevices += other.FieldDevices
	c.Specifications += other.Specifications
	c.BacnetObjects += other.BacnetObjects
	c.SoftwareReferences += other.SoftwareReferences
	c.AlarmValues += other.AlarmValues
}

type Status string

const (
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

type OutputType string

const (
	OutputTypeExcel OutputType = "excel"
	OutputTypeZip   OutputType = "zip"
)

type Job struct {
	ID          uuid.UUID
	Status      Status
	Progress    int
	Message     string
	OutputType  OutputType
	FileName    string
	ContentType string
	FilePath    string
	Error       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Scope       Scope
}

type Scope struct {
	Kind       AccessScope
	ProjectIDs []uuid.UUID
}

type DownloadAuthorization struct {
	RequesterID   uuid.UUID
	RequesterRole domainuser.Role
	Scope         Scope
}

type DownloadAuthorizer interface {
	CanDownload(ctx context.Context, authorization DownloadAuthorization) (bool, error)
}

type Controller struct {
	ID               uuid.UUID
	ControlCabinetID uuid.UUID
	GADevice         string

	IWSCode             string
	BuildingGroup       int
	ControlCabinetNr    string
	MinSystemPartNumber string // 4-digit zero-padded, e.g. "0100"
	DeviceName          string // computed: {iwsCode}_{buildingGroup}_{minSysPart}_{gaDevice}
	DeviceInstance      string // computed: {lastTwoIws}{gaDeviceIndex}{buildingGroup}
	DeviceDescription   string
	DeviceLocation      string
	IPAddress           string
	Subnet              string
	Gateway             string
	VLAN                string
}
