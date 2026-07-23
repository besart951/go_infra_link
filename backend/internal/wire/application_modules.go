package wire

import (
	"context"
	"fmt"
	"log/slog"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	appfacility "github.com/besart951/go_infra_link/backend/internal/application/facility"
	appbacnetobject "github.com/besart951/go_infra_link/backend/internal/application/facility/bacnetobject"
	appcontrolcabinet "github.com/besart951/go_infra_link/backend/internal/application/facility/controlcabinet"
	appfielddevice "github.com/besart951/go_infra_link/backend/internal/application/facility/fielddevice"
	appobjectdata "github.com/besart951/go_infra_link/backend/internal/application/facility/objectdata"
	appspscontroller "github.com/besart951/go_infra_link/backend/internal/application/facility/spscontroller"
	apphistory "github.com/besart951/go_infra_link/backend/internal/application/history"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	infraltime "github.com/besart951/go_infra_link/backend/internal/infrastructure/realtime"
	infratransaction "github.com/besart951/go_infra_link/backend/internal/infrastructure/transaction"
	"github.com/besart951/go_infra_link/backend/internal/service/auditctx"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectSPSControllerReassignmentWorkflow struct {
	links         domainProject.ProjectSPSControllerRepository
	facilityLinks *projectservice.ProjectFacilityLinkService
}

type projectControlCabinetReassignmentWorkflow struct {
	links         domainProject.ProjectControlCabinetRepository
	facilityLinks *projectservice.ProjectFacilityLinkService
}

type projectObjectDataAssociationWorkflow struct {
	projects   domainProject.ProjectRepository
	objectData domainObjectData.ObjectDataStore
}

type projectRestoreAccessPolicy interface {
	CanAccessProject(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		*domainUser.Role,
	) (bool, error)
}

type projectControlCabinetLookup interface {
	GetByControlCabinetID(
		context.Context,
		uuid.UUID,
	) ([]*domainProject.ProjectControlCabinet, error)
}

type controlCabinetHistoryScopeReader interface {
	HasHistoricalProjectControlCabinet(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (bool, error)
}

type projectControlCabinetRestoreScope struct {
	access  projectRestoreAccessPolicy
	links   projectControlCabinetLookup
	history controlCabinetHistoryScopeReader
}

type projectHistoryAccessScope struct {
	access projectRestoreAccessPolicy
}

func (s *projectHistoryAccessScope) RequireProjectAccess(
	ctx context.Context,
	actorID uuid.UUID,
	projectID uuid.UUID,
) error {
	if s == nil || s.access == nil {
		return apphistory.ErrProjectTimelineNotConfigured
	}
	hasAccess, err := s.access.CanAccessProject(ctx, actorID, projectID, nil)
	if err != nil {
		return err
	}
	if !hasAccess {
		return apphistory.ErrProjectTimelineAccessDenied
	}
	return nil
}

func (s *projectControlCabinetRestoreScope) RequireControlCabinetRestoreScope(
	ctx context.Context,
	actorID uuid.UUID,
	projectID uuid.UUID,
	controlCabinetID uuid.UUID,
) error {
	if s == nil || s.access == nil || s.links == nil || s.history == nil {
		return appcontrolcabinet.ErrProjectRestoreNotConfigured
	}
	hasAccess, err := s.access.CanAccessProject(ctx, actorID, projectID, nil)
	if err != nil {
		return err
	}
	if !hasAccess {
		return appcontrolcabinet.ErrProjectRestoreAccessDenied
	}

	links, err := s.links.GetByControlCabinetID(ctx, controlCabinetID)
	if err != nil {
		return err
	}
	for _, link := range links {
		if link != nil && link.ProjectID == projectID &&
			link.ControlCabinetID == controlCabinetID {
			return nil
		}
	}
	hasHistoricalScope, err := s.history.HasHistoricalProjectControlCabinet(
		ctx,
		projectID,
		controlCabinetID,
	)
	if err != nil {
		return err
	}
	if !hasHistoricalScope {
		return domain.ErrNotFound
	}
	return nil
}

func (w *projectObjectDataAssociationWorkflow) RequireProject(
	ctx context.Context,
	projectID uuid.UUID,
) error {
	_, err := domain.GetByID(ctx, w.projects, projectID)
	return err
}

func (w *projectObjectDataAssociationWorkflow) GetObjectData(
	ctx context.Context,
	objectDataID uuid.UUID,
) (*domainFacility.ObjectData, error) {
	return domain.GetByID(ctx, w.objectData, objectDataID)
}

func (w *projectObjectDataAssociationWorkflow) UpdateObjectData(
	ctx context.Context,
	objectData *domainFacility.ObjectData,
) error {
	return w.objectData.Update(ctx, objectData)
}

func (w *projectControlCabinetReassignmentWorkflow) GetByIds(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectControlCabinet, error) {
	return w.links.GetByIds(ctx, ids)
}

func (w *projectControlCabinetReassignmentWorkflow) UpdateControlCabinet(
	ctx context.Context,
	linkID uuid.UUID,
	projectID uuid.UUID,
	controlCabinetID uuid.UUID,
) (*domainProject.ProjectControlCabinet, error) {
	return w.facilityLinks.UpdateControlCabinet(
		ctx,
		linkID,
		projectID,
		controlCabinetID,
	)
}

func (w *projectSPSControllerReassignmentWorkflow) GetByIds(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectSPSController, error) {
	return w.links.GetByIds(ctx, ids)
}

func (w *projectSPSControllerReassignmentWorkflow) UpdateSPSController(
	ctx context.Context,
	linkID uuid.UUID,
	projectID uuid.UUID,
	spsControllerID uuid.UUID,
) (*domainProject.ProjectSPSController, error) {
	return w.facilityLinks.UpdateSPSController(
		ctx,
		linkID,
		projectID,
		spsControllerID,
	)
}

func newFacilityApplicationServices(
	gormDB *gorm.DB,
	repos *Repositories,
	legacy *facilityservice.Services,
	projectAccess *projectservice.ProjectAccessPolicyService,
	runtime *RuntimeAdapters,
) *appfacility.Services {
	var dispatcher appcollaboration.CommandDispatcher
	if runtime != nil && runtime.ProjectCollaboration != nil {
		port := infraltime.NewCollaborationCommandAdapter(runtime.ProjectCollaboration)
		handler := appcollaboration.NewProjectCommandHandler(port)
		dispatcher = appcollaboration.NewDispatcher(appcollaboration.DispatcherDependencies{
			FacilityHierarchyRefresh:      handler,
			ControlCabinetCreated:         handler,
			ControlCabinetCloned:          handler,
			ControlCabinetDeleted:         handler,
			ControlCabinetUpdated:         handler,
			ControlCabinetMoved:           handler,
			FieldDeviceUpdated:            handler,
			FieldDeviceMoved:              handler,
			FieldDeviceDeleted:            handler,
			FieldDevicesCreated:           handler,
			BacnetObjectCreated:           handler,
			BacnetObjectUpdated:           handler,
			SPSControllerCreated:          handler,
			SPSControllerCloned:           handler,
			SPSControllerSystemTypeCloned: handler,
			SPSControllerUpdated:          handler,
			SPSControllerMoved:            handler,
			SPSControllerDeleted:          handler,
		})
	}

	actor := func(ctx context.Context) *uuid.UUID {
		actorID, _ := auditctx.ActorID(ctx)
		return actorID
	}
	fieldDeviceReportError := func(err error) {
		slog.Warn("FieldDevice collaboration dispatch failed", "err", err)
	}
	spsControllerReportError := func(err error) {
		slog.Warn("SPSController collaboration dispatch failed", "err", err)
	}
	controlCabinetReportError := func(err error) {
		slog.Warn("ControlCabinet collaboration dispatch failed", "err", err)
	}
	bacnetObjectReportError := func(err error) {
		slog.Warn("BACnetObject collaboration dispatch failed", "err", err)
	}
	objectDataReportError := func(err error) {
		slog.Warn("ObjectData collaboration dispatch failed", "err", err)
	}

	bacnetObjectTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appbacnetobject.UpdateWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("BACnetObject application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.BacnetObject, nil
	}
	bacnetObjectCreateTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appbacnetobject.CreateWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("BACnetObject application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.BacnetObject, nil
	}
	bacnetAlarmValuesTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appbacnetobject.ReplaceAlarmValuesWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("BACnet alarm-value application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.BacnetAlarmValue, nil
	}
	bacnetObjectCreate := appbacnetobject.NewCreateHandler(appbacnetobject.CreateDependencies{
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: bacnetObjectCreateTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectFieldDevices,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         bacnetObjectReportError,
	})
	bacnetObjectUpdate := appbacnetobject.NewUpdateHandler(appbacnetobject.UpdateDependencies{
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: bacnetObjectTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectFieldDevices,
		ObjectDataOwners:    repos.FacilityBacnetObjectOwners,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         bacnetObjectReportError,
	})
	bacnetAlarmValuesReplace := appbacnetobject.NewReplaceAlarmValuesHandler(
		appbacnetobject.ReplaceAlarmValuesDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: bacnetAlarmValuesTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			BacnetObjects:       repos.FacilityBacnetObjects,
			ProjectLinks:        repos.ProjectFieldDevices,
			ObjectDataOwners:    repos.FacilityBacnetObjectOwners,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         bacnetObjectReportError,
		},
	)
	objectDataProjectAssociationTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appobjectdata.ProjectAssociationWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("project ObjectData association transaction unit: %w", err)
		}
		return &projectObjectDataAssociationWorkflow{
			projects:   txRepos.Project,
			objectData: txRepos.FacilityObjectData,
		}, nil
	}
	objectDataProjectAssociation := appobjectdata.NewProjectAssociationHandler(
		appobjectdata.ProjectAssociationDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: objectDataProjectAssociationTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         objectDataReportError,
		},
	)

	controlCabinetTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appcontrolcabinet.UpdateWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("ControlCabinet application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.ControlCabinet, nil
	}
	controlCabinetCreateTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appcontrolcabinet.CreateWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("ControlCabinet application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.ControlCabinet, nil
	}
	controlCabinetCloneTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appcontrolcabinet.CloneWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("ControlCabinet clone application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.ControlCabinet, nil
	}
	controlCabinetProjectCloneTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appcontrolcabinet.CloneForProjectWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("project ControlCabinet clone application transaction unit: %w", err)
		}
		txFacilityServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		txProjectServices := projectservice.NewServices(
			buildProjectDependencies(txRepos, txFacilityServices),
		)
		return txProjectServices.FacilityLink, nil
	}
	controlCabinetProjectAssignmentTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appcontrolcabinet.AssignToProjectWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("project ControlCabinet assignment transaction unit: %w", err)
		}
		txFacilityServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		txProjectServices := projectservice.NewServices(
			buildProjectDependencies(txRepos, txFacilityServices),
		)
		return txProjectServices.FacilityLink, nil
	}
	controlCabinetProjectReassignmentTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appcontrolcabinet.ReassignProjectLinkWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("project ControlCabinet reassignment transaction unit: %w", err)
		}
		txFacilityServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		txProjectServices := projectservice.NewServices(
			buildProjectDependencies(txRepos, txFacilityServices),
		)
		return &projectControlCabinetReassignmentWorkflow{
			links:         txRepos.ProjectControlCabinets,
			facilityLinks: txProjectServices.FacilityLink,
		}, nil
	}
	controlCabinetDeleteTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appcontrolcabinet.DeleteWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("ControlCabinet delete application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.ControlCabinet, nil
	}
	controlCabinetCreate := appcontrolcabinet.NewCreateHandler(appcontrolcabinet.CreateDependencies{
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: controlCabinetCreateTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectControlCabinets,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         controlCabinetReportError,
	})
	controlCabinetClone := appcontrolcabinet.NewCloneHandler(appcontrolcabinet.CloneDependencies{
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: controlCabinetCloneTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectControlCabinets,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         controlCabinetReportError,
	})
	controlCabinetProjectClone := appcontrolcabinet.NewCloneForProjectHandler(
		appcontrolcabinet.CloneForProjectDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: controlCabinetProjectCloneTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         controlCabinetReportError,
		},
	)
	controlCabinetProjectAssignment := appcontrolcabinet.NewAssignToProjectHandler(
		appcontrolcabinet.AssignToProjectDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: controlCabinetProjectAssignmentTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         controlCabinetReportError,
		},
	)
	controlCabinetProjectReassignment := appcontrolcabinet.NewReassignProjectLinkHandler(
		appcontrolcabinet.ReassignProjectLinkDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: controlCabinetProjectReassignmentTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         controlCabinetReportError,
		},
	)
	controlCabinetUpdate := appcontrolcabinet.NewUpdateHandler(appcontrolcabinet.UpdateDependencies{
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: controlCabinetTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectControlCabinets,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         controlCabinetReportError,
	})
	controlCabinetDelete := appcontrolcabinet.NewDeleteHandler(appcontrolcabinet.DeleteDependencies{
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: controlCabinetDeleteTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectControlCabinets,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         controlCabinetReportError,
	})
	controlCabinetProjectRestore := appcontrolcabinet.NewRestoreForProjectHandler(
		appcontrolcabinet.RestoreForProjectDependencies{
			Scope: &projectControlCabinetRestoreScope{
				access:  projectAccess,
				links:   repos.ProjectControlCabinets,
				history: repos.History,
			},
			Restorer:     repos.History,
			ProjectLinks: repos.ProjectControlCabinets,
			Dispatcher:   dispatcher,
			Actor:        actor,
			ReportError:  controlCabinetReportError,
		},
	)

	multiCreate := appfielddevice.NewMultiCreateHandler(appfielddevice.MultiCreateDependencies{
		Executor:     legacy.FieldDevice,
		HistoryBatch: auditctx.WithBatchID,
		Actor:        actor,
	})
	projectMultiCreateTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appfielddevice.MultiCreateForProjectWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("project FieldDevice multi-create application transaction unit: %w", err)
		}
		txDB, err := infratransaction.GormDB(unit)
		if err != nil {
			return nil, fmt.Errorf("project FieldDevice multi-create transaction database: %w", err)
		}
		txFacilityServices := facilityservice.NewServices(
			buildFacilityRepositories(txRepos),
			facilityservice.Config{
				TxRunner: infratransaction.NewGormRunner(txDB),
				TxRepositories: func(nestedUnit apptransaction.UnitOfWork) (facilityservice.Repositories, error) {
					nestedRepos, nestedErr := repositoriesFromUnit(nestedUnit)
					if nestedErr != nil {
						return facilityservice.Repositories{}, fmt.Errorf(
							"project FieldDevice item transaction unit: %w",
							nestedErr,
						)
					}
					return buildFacilityRepositories(nestedRepos), nil
				},
			},
		)
		txProjectServices := projectservice.NewServices(
			buildProjectDependencies(txRepos, txFacilityServices),
		)
		return txProjectServices.FacilityLink, nil
	}
	projectMultiCreate := appfielddevice.NewMultiCreateForProjectHandler(
		appfielddevice.MultiCreateForProjectDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: projectMultiCreateTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         fieldDeviceReportError,
		},
	)
	projectAssignmentTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appfielddevice.AssignToProjectWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("ProjectFieldDevice assignment transaction unit: %w", err)
		}
		txFacilityServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		txProjectServices := projectservice.NewServices(
			buildProjectDependencies(txRepos, txFacilityServices),
		)
		return txProjectServices.FacilityLink, nil
	}
	projectAssignment := appfielddevice.NewAssignToProjectHandler(
		appfielddevice.AssignToProjectDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: projectAssignmentTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         fieldDeviceReportError,
		},
	)
	projectBulkAssignment := appfielddevice.NewBulkAssignToProjectHandler(
		appfielddevice.BulkAssignToProjectDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: projectAssignmentTransactionWorkflow,
			Projects:            repos.Project,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         fieldDeviceReportError,
		},
	)
	projectLinkReassignmentTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appfielddevice.ReassignProjectLinkWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("project FieldDevice reassignment transaction unit: %w", err)
		}
		return txRepos.ProjectFieldDevices, nil
	}
	projectLinkReassignment := appfielddevice.NewReassignProjectLinkHandler(
		appfielddevice.ReassignProjectLinkDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: projectLinkReassignmentTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         fieldDeviceReportError,
		},
	)
	bulkUpdate := appfielddevice.NewBulkUpdateHandler(appfielddevice.BulkUpdateDependencies{
		Executor:     legacy.FieldDevice,
		HistoryBatch: auditctx.WithBatchID,
		ProjectLinks: repos.ProjectFieldDevices,
		Dispatcher:   dispatcher,
		Actor:        actor,
		ReportError:  fieldDeviceReportError,
	})

	transactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appfielddevice.UpdateWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("FieldDevice application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.FieldDevice, nil
	}
	deleteTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appfielddevice.DeleteWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("FieldDevice delete application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.FieldDevice, nil
	}
	bulkDeleteTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appfielddevice.BulkDeleteWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("FieldDevice bulk-delete application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.FieldDevice, nil
	}
	update := appfielddevice.NewUpdateHandler(appfielddevice.UpdateDependencies{
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: transactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectFieldDevices,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         fieldDeviceReportError,
	})
	deleteHandler := appfielddevice.NewDeleteHandler(appfielddevice.DeleteDependencies{
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: deleteTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectFieldDevices,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         fieldDeviceReportError,
	})
	bulkDelete := appfielddevice.NewBulkDeleteHandler(appfielddevice.BulkDeleteDependencies{
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: bulkDeleteTransactionWorkflow,
		Snapshots:           repos.FacilityFieldDevices,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectFieldDevices,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         fieldDeviceReportError,
	})

	spsTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appspscontroller.UpdateWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("SPSController application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.SPSController, nil
	}
	spsCreateTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appspscontroller.CreateWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("SPSController application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.SPSController, nil
	}
	spsCloneTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appspscontroller.CloneWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("SPSController clone application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.SPSController, nil
	}
	spsSystemTypeCloneTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appspscontroller.CloneSystemTypeWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("SPSControllerSystemType clone application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.SPSControllerSystemType, nil
	}
	spsSystemTypeDeleteTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appspscontroller.DeleteSystemTypeWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("SPSControllerSystemType delete application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.SPSControllerSystemType, nil
	}
	spsProjectCloneTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appspscontroller.CloneForProjectWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("project SPSController clone application transaction unit: %w", err)
		}
		txFacilityServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		txProjectServices := projectservice.NewServices(
			buildProjectDependencies(txRepos, txFacilityServices),
		)
		return txProjectServices.FacilityLink, nil
	}
	spsProjectAssignmentTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appspscontroller.AssignToProjectWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("project SPSController assignment transaction unit: %w", err)
		}
		txFacilityServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		txProjectServices := projectservice.NewServices(
			buildProjectDependencies(txRepos, txFacilityServices),
		)
		return txProjectServices.FacilityLink, nil
	}
	spsProjectReassignmentTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appspscontroller.ReassignProjectLinkWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("project SPSController reassignment transaction unit: %w", err)
		}
		txFacilityServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		txProjectServices := projectservice.NewServices(
			buildProjectDependencies(txRepos, txFacilityServices),
		)
		return &projectSPSControllerReassignmentWorkflow{
			links:         txRepos.ProjectSPSControllers,
			facilityLinks: txProjectServices.FacilityLink,
		}, nil
	}
	spsProjectSystemTypeCloneTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appspscontroller.CloneSystemTypeForProjectWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("project SPSControllerSystemType clone application transaction unit: %w", err)
		}
		txFacilityServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		txProjectServices := projectservice.NewServices(
			buildProjectDependencies(txRepos, txFacilityServices),
		)
		return txProjectServices.FacilityLink, nil
	}
	spsDeleteTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appspscontroller.DeleteWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("SPSController delete application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return txServices.SPSController, nil
	}
	spsCreate := appspscontroller.NewCreateHandler(appspscontroller.CreateDependencies{
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: spsCreateTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectSPSControllers,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         spsControllerReportError,
	})
	spsClone := appspscontroller.NewCloneHandler(appspscontroller.CloneDependencies{
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: spsCloneTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectSPSControllers,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         spsControllerReportError,
	})
	spsSystemTypeClone := appspscontroller.NewCloneSystemTypeHandler(
		appspscontroller.CloneSystemTypeDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: spsSystemTypeCloneTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Actor:               actor,
		},
	)
	spsSystemTypeDelete := appspscontroller.NewDeleteSystemTypeHandler(
		appspscontroller.DeleteSystemTypeDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: spsSystemTypeDeleteTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Actor:               actor,
		},
	)
	spsProjectClone := appspscontroller.NewCloneForProjectHandler(
		appspscontroller.CloneForProjectDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: spsProjectCloneTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         spsControllerReportError,
		},
	)
	spsProjectAssignment := appspscontroller.NewAssignToProjectHandler(
		appspscontroller.AssignToProjectDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: spsProjectAssignmentTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         spsControllerReportError,
		},
	)
	spsProjectReassignment := appspscontroller.NewReassignProjectLinkHandler(
		appspscontroller.ReassignProjectLinkDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: spsProjectReassignmentTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         spsControllerReportError,
		},
	)
	spsProjectSystemTypeClone := appspscontroller.NewCloneSystemTypeForProjectHandler(
		appspscontroller.CloneSystemTypeForProjectDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: spsProjectSystemTypeCloneTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         spsControllerReportError,
		},
	)
	spsUpdate := appspscontroller.NewUpdateHandler(appspscontroller.UpdateDependencies{
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: spsTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectSPSControllers,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         spsControllerReportError,
	})
	spsDelete := appspscontroller.NewDeleteHandler(appspscontroller.DeleteDependencies{
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: spsDeleteTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectSPSControllers,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         spsControllerReportError,
	})

	return &appfacility.Services{
		BacnetObject: &appfacility.BacnetObjectModule{
			Create:             bacnetObjectCreate,
			Update:             bacnetObjectUpdate,
			ReplaceAlarmValues: bacnetAlarmValuesReplace,
		},
		ControlCabinet: &appfacility.ControlCabinetModule{
			Create:              controlCabinetCreate,
			AssignToProject:     controlCabinetProjectAssignment,
			ReassignProjectLink: controlCabinetProjectReassignment,
			Clone:               controlCabinetClone,
			CloneForProject:     controlCabinetProjectClone,
			RestoreForProject:   controlCabinetProjectRestore,
			Update:              controlCabinetUpdate,
			Delete:              controlCabinetDelete,
		},
		FieldDevice: &appfacility.FieldDeviceModule{
			MultiCreate:           multiCreate,
			MultiCreateForProject: projectMultiCreate,
			AssignToProject:       projectAssignment,
			BulkAssignToProject:   projectBulkAssignment,
			ReassignProjectLink:   projectLinkReassignment,
			Update:                update,
			Delete:                deleteHandler,
			BulkUpdate:            bulkUpdate,
			BulkDelete:            bulkDelete,
		},
		ObjectData: &appfacility.ObjectDataModule{
			ProjectAssociation: objectDataProjectAssociation,
		},
		SPSController: &appfacility.SPSControllerModule{
			Create:                    spsCreate,
			AssignToProject:           spsProjectAssignment,
			ReassignProjectLink:       spsProjectReassignment,
			Clone:                     spsClone,
			CloneSystemType:           spsSystemTypeClone,
			DeleteSystemType:          spsSystemTypeDelete,
			CloneForProject:           spsProjectClone,
			CloneSystemTypeForProject: spsProjectSystemTypeClone,
			Update:                    spsUpdate,
			Delete:                    spsDelete,
		},
	}
}

func newHistoryApplicationServices(
	repos *Repositories,
	projectAccess projectRestoreAccessPolicy,
) *apphistory.Services {
	actor := func(ctx context.Context) *uuid.UUID {
		actorID, _ := auditctx.ActorID(ctx)
		return actorID
	}
	return &apphistory.Services{
		ProjectTimeline: apphistory.NewProjectTimelineHandler(
			apphistory.ProjectTimelineDependencies{
				Access:   &projectHistoryAccessScope{access: projectAccess},
				Timeline: repos.History,
				Actor:    actor,
			},
		),
	}
}
