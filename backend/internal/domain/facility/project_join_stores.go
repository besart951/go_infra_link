package facility

import "github.com/google/uuid"

// ProjectControlCabinetStore manages the many-to-many relationship between projects and control cabinets
type ProjectControlCabinetStore interface {
	// Link associates a control cabinet with a project
	Link(projectID uuid.UUID, controlCabinetID uuid.UUID) error
	
	// Unlink removes the association between a control cabinet and a project
	Unlink(projectID uuid.UUID, controlCabinetID uuid.UUID) error
	
	// GetProjectIDsByControlCabinet returns all project IDs associated with a control cabinet
	GetProjectIDsByControlCabinet(controlCabinetID uuid.UUID) ([]uuid.UUID, error)
	
	// GetControlCabinetIDsByProject returns all control cabinet IDs associated with a project
	GetControlCabinetIDsByProject(projectID uuid.UUID) ([]uuid.UUID, error)
}

// ProjectSPSControllerStore manages the many-to-many relationship between projects and SPS controllers
type ProjectSPSControllerStore interface {
	// Link associates an SPS controller with a project
	Link(projectID uuid.UUID, spsControllerID uuid.UUID) error
	
	// Unlink removes the association between an SPS controller and a project
	Unlink(projectID uuid.UUID, spsControllerID uuid.UUID) error
	
	// GetProjectIDsBySPSController returns all project IDs associated with an SPS controller
	GetProjectIDsBySPSController(spsControllerID uuid.UUID) ([]uuid.UUID, error)
	
	// GetSPSControllerIDsByProject returns all SPS controller IDs associated with a project
	GetSPSControllerIDsByProject(projectID uuid.UUID) ([]uuid.UUID, error)
}

// ProjectFieldDeviceStore manages the many-to-many relationship between projects and field devices
type ProjectFieldDeviceStore interface {
	// Link associates a field device with a project
	Link(projectID uuid.UUID, fieldDeviceID uuid.UUID) error
	
	// Unlink removes the association between a field device and a project
	Unlink(projectID uuid.UUID, fieldDeviceID uuid.UUID) error
	
	// GetProjectIDsByFieldDevice returns all project IDs associated with a field device
	GetProjectIDsByFieldDevice(fieldDeviceID uuid.UUID) ([]uuid.UUID, error)
	
	// GetFieldDeviceIDsByProject returns all field device IDs associated with a project
	GetFieldDeviceIDsByProject(projectID uuid.UUID) ([]uuid.UUID, error)
}
