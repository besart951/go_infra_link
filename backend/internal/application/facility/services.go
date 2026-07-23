package facility

import (
	appbacnetobject "github.com/besart951/go_infra_link/backend/internal/application/facility/bacnetobject"
	appcontrolcabinet "github.com/besart951/go_infra_link/backend/internal/application/facility/controlcabinet"
	appfielddevice "github.com/besart951/go_infra_link/backend/internal/application/facility/fielddevice"
	appobjectdata "github.com/besart951/go_infra_link/backend/internal/application/facility/objectdata"
	appspscontroller "github.com/besart951/go_infra_link/backend/internal/application/facility/spscontroller"
)

type BacnetObjectModule struct {
	Create             *appbacnetobject.CreateHandler
	Update             *appbacnetobject.UpdateHandler
	ReplaceAlarmValues *appbacnetobject.ReplaceAlarmValuesHandler
}

type ControlCabinetModule struct {
	Create              *appcontrolcabinet.CreateHandler
	AssignToProject     *appcontrolcabinet.AssignToProjectHandler
	ReassignProjectLink *appcontrolcabinet.ReassignProjectLinkHandler
	Clone               *appcontrolcabinet.CloneHandler
	CloneForProject     *appcontrolcabinet.CloneForProjectHandler
	RestoreForProject   *appcontrolcabinet.RestoreForProjectHandler
	Update              *appcontrolcabinet.UpdateHandler
	Delete              *appcontrolcabinet.DeleteHandler
}

type FieldDeviceModule struct {
	MultiCreate           *appfielddevice.MultiCreateHandler
	MultiCreateForProject *appfielddevice.MultiCreateForProjectHandler
	AssignToProject       *appfielddevice.AssignToProjectHandler
	BulkAssignToProject   *appfielddevice.BulkAssignToProjectHandler
	ReassignProjectLink   *appfielddevice.ReassignProjectLinkHandler
	Update                *appfielddevice.UpdateHandler
	Delete                *appfielddevice.DeleteHandler
	BulkUpdate            *appfielddevice.BulkUpdateHandler
	BulkDelete            *appfielddevice.BulkDeleteHandler
}

type ObjectDataModule struct {
	ProjectAssociation *appobjectdata.ProjectAssociationHandler
}

type SPSControllerModule struct {
	Create                    *appspscontroller.CreateHandler
	AssignToProject           *appspscontroller.AssignToProjectHandler
	ReassignProjectLink       *appspscontroller.ReassignProjectLinkHandler
	Clone                     *appspscontroller.CloneHandler
	CloneSystemType           *appspscontroller.CloneSystemTypeHandler
	DeleteSystemType          *appspscontroller.DeleteSystemTypeHandler
	CloneForProject           *appspscontroller.CloneForProjectHandler
	CloneSystemTypeForProject *appspscontroller.CloneSystemTypeForProjectHandler
	Update                    *appspscontroller.UpdateHandler
	Delete                    *appspscontroller.DeleteHandler
}

type Services struct {
	BacnetObject   *BacnetObjectModule
	ControlCabinet *ControlCabinetModule
	FieldDevice    *FieldDeviceModule
	ObjectData     *ObjectDataModule
	SPSController  *SPSControllerModule
}
