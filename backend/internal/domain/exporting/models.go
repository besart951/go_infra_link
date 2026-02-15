package exporting

import (
	"time"

	"github.com/google/uuid"
)

type Request struct {
	ProjectIDs        []uuid.UUID
	BuildingIDs       []uuid.UUID
	ControlCabinetIDs []uuid.UUID
	SPSControllerIDs  []uuid.UUID
	ForceAsync        bool
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
