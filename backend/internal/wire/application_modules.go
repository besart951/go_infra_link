package wire

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	appfacility "github.com/besart951/go_infra_link/backend/internal/application/facility"
	appbacnetobject "github.com/besart951/go_infra_link/backend/internal/application/facility/bacnetobject"
	appcontrolcabinet "github.com/besart951/go_infra_link/backend/internal/application/facility/controlcabinet"
	appfielddevice "github.com/besart951/go_infra_link/backend/internal/application/facility/fielddevice"
	appobjectdata "github.com/besart951/go_infra_link/backend/internal/application/facility/objectdata"
	appprojectlink "github.com/besart951/go_infra_link/backend/internal/application/facility/projectlink"
	appspscontroller "github.com/besart951/go_infra_link/backend/internal/application/facility/spscontroller"
	apphistory "github.com/besart951/go_infra_link/backend/internal/application/history"
	appproject "github.com/besart951/go_infra_link/backend/internal/application/project"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	infraltime "github.com/besart951/go_infra_link/backend/internal/infrastructure/realtime"
	infratransaction "github.com/besart951/go_infra_link/backend/internal/infrastructure/transaction"
	facilityrepo "github.com/besart951/go_infra_link/backend/internal/repository/facilitysql"
	"github.com/besart951/go_infra_link/backend/internal/service/auditctx"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

type projectControlCabinetRestoreWorkflow struct {
	restorer appcontrolcabinet.ControlCabinetHistoryRestorer
	links    appcontrolcabinet.ProjectLinkReader
}

type projectFacilityUnlinkWorkflow struct {
	controlCabinets domainProject.ProjectControlCabinetRepository
	spsControllers  domainProject.ProjectSPSControllerRepository
	fieldDevices    domainProject.ProjectFieldDeviceRepository
}

type projectDeletionWorkflow struct {
	db         *gorm.DB
	projects   domainProject.ProjectRepository
	objectData domainObjectData.ObjectDataStore
}

type projectCloneSourceKind uint8

const (
	projectCloneControlCabinet projectCloneSourceKind = iota + 1
	projectCloneSPSController
	projectCloneSPSControllerSystemType
)

type projectFacilityCloneWorkflow struct {
	*projectservice.ProjectFacilityLinkService
	db         *gorm.DB
	sourceKind projectCloneSourceKind
}

type spsControllerSystemTypeCloneStore interface {
	CopyByID(
		context.Context,
		uuid.UUID,
	) (*domainFacility.SPSControllerSystemType, error)
	GetByID(
		context.Context,
		uuid.UUID,
	) (*domainFacility.SPSControllerSystemType, error)
}

type globalSPSControllerSystemTypeCloneWorkflow struct {
	clones spsControllerSystemTypeCloneStore
	db     *gorm.DB
}

type controlCabinetCloneStore interface {
	CopyByID(context.Context, uuid.UUID) (*domainFacility.ControlCabinet, error)
	GetByID(context.Context, uuid.UUID) (*domainFacility.ControlCabinet, error)
}

type globalControlCabinetCloneWorkflow struct {
	clones        controlCabinetCloneStore
	facilityLinks *projectservice.ProjectFacilityLinkService
	db            *gorm.DB
}

func (workflow *globalControlCabinetCloneWorkflow) CopyByID(
	ctx context.Context,
	id uuid.UUID,
) (*domainFacility.ControlCabinet, error) {
	return workflow.clones.CopyByID(ctx, id)
}

func (workflow *globalControlCabinetCloneWorkflow) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domainFacility.ControlCabinet, error) {
	return workflow.clones.GetByID(ctx, id)
}

func (workflow *globalControlCabinetCloneWorkflow) GetSourceProjectIDs(
	ctx context.Context,
	controlCabinetID uuid.UUID,
) ([]uuid.UUID, error) {
	var projectIDs []uuid.UUID
	result := workflow.db.WithContext(ctx).Raw(`
		SELECT project_link.project_id
		FROM project_control_cabinets AS project_link
		WHERE project_link.control_cabinet_id = ?
		ORDER BY project_link.project_id
		FOR UPDATE OF project_link
	`, controlCabinetID).Scan(&projectIDs)
	return projectIDs, result.Error
}

func (workflow *globalControlCabinetCloneWorkflow) AssignCopyToProject(
	ctx context.Context,
	projectID, controlCabinetID uuid.UUID,
) error {
	_, err := workflow.facilityLinks.CreateControlCabinet(
		ctx,
		projectID,
		controlCabinetID,
	)
	return err
}

func (workflow *globalSPSControllerSystemTypeCloneWorkflow) CopyByID(
	ctx context.Context,
	id uuid.UUID,
) (*domainFacility.SPSControllerSystemType, error) {
	return workflow.clones.CopyByID(ctx, id)
}

func (workflow *globalSPSControllerSystemTypeCloneWorkflow) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domainFacility.SPSControllerSystemType, error) {
	return workflow.clones.GetByID(ctx, id)
}

func (workflow *globalSPSControllerSystemTypeCloneWorkflow) GetOwningProjectIDs(
	ctx context.Context,
	spsControllerID uuid.UUID,
) ([]uuid.UUID, error) {
	var projectIDs []uuid.UUID
	result := workflow.db.WithContext(ctx).Raw(`
		SELECT project_link.project_id
		FROM project_sps_controllers AS project_link
		WHERE project_link.sps_controller_id = ?
		ORDER BY project_link.project_id
		FOR UPDATE OF project_link
	`, spsControllerID).Scan(&projectIDs)
	return projectIDs, result.Error
}

func (workflow *projectFacilityCloneWorkflow) RequireSourceAccess(
	ctx context.Context,
	projectID, sourceID uuid.UUID,
) error {
	var query string
	switch workflow.sourceKind {
	case projectCloneControlCabinet:
		query = `
			SELECT project_link.id
			FROM project_control_cabinets AS project_link
			WHERE project_link.project_id = ?
			  AND project_link.control_cabinet_id = ?
			LIMIT 1
			FOR UPDATE OF project_link
		`
	case projectCloneSPSController:
		query = `
			SELECT project_link.id
			FROM project_sps_controllers AS project_link
			WHERE project_link.project_id = ?
			  AND project_link.sps_controller_id = ?
			LIMIT 1
			FOR UPDATE OF project_link
		`
	case projectCloneSPSControllerSystemType:
		query = `
			SELECT project_link.id
			FROM sps_controller_system_types AS source
			INNER JOIN project_sps_controllers AS project_link
				ON project_link.sps_controller_id = source.sps_controller_id
			WHERE project_link.project_id = ?
			  AND source.id = ?
			LIMIT 1
			FOR UPDATE OF project_link
		`
	default:
		return domain.ErrInvalidArgument
	}

	var linkID uuid.UUID
	result := workflow.db.WithContext(ctx).Raw(query, projectID, sourceID).Scan(&linkID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 || linkID == uuid.Nil {
		return domain.ErrNotFound
	}
	return nil
}

func (workflow *projectDeletionWorkflow) GetProjectForDeletion(
	ctx context.Context,
	projectID uuid.UUID,
) (*appproject.Snapshot, error) {
	var snapshot appproject.Snapshot
	err := workflow.db.WithContext(ctx).
		Table("projects").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id, status").
		Where("id = ?", projectID).
		Take(&snapshot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (workflow *projectDeletionWorkflow) GetActiveUserRole(
	ctx context.Context,
	userID uuid.UUID,
) (domainUser.Role, error) {
	var row struct {
		Role domainUser.Role
	}
	err := workflow.db.WithContext(ctx).
		Table("users").
		Select("role").
		Where("id = ?", userID).
		Where("is_active = ?", true).
		Where("disabled_at IS NULL").
		Where("deleted_at IS NULL").
		Where("anonymized_at IS NULL").
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return row.Role, nil
}

func (workflow *projectDeletionWorkflow) HasHierarchyLinks(
	ctx context.Context,
	projectID uuid.UUID,
) (bool, error) {
	var linked bool
	err := workflow.db.WithContext(ctx).Raw(`
		SELECT EXISTS (
			SELECT 1 FROM project_control_cabinets WHERE project_id = ?
			UNION ALL
			SELECT 1 FROM project_sps_controllers WHERE project_id = ?
			UNION ALL
			SELECT 1 FROM project_field_devices WHERE project_id = ?
		)
	`, projectID, projectID, projectID).Scan(&linked).Error
	return linked, err
}

func (workflow *projectDeletionWorkflow) ListProjectObjectDataIDs(
	ctx context.Context,
	projectID, after uuid.UUID,
	limit int,
) ([]uuid.UUID, error) {
	query := workflow.db.WithContext(ctx).
		Table("object_data").
		Select("id").
		Where("project_id = ?", projectID)
	if after != uuid.Nil {
		query = query.Where("id > ?", after)
	}
	var ids []uuid.UUID
	err := query.Order("id ASC").Limit(limit).Scan(&ids).Error
	return ids, err
}

func (workflow *projectDeletionWorkflow) DeleteObjectData(
	ctx context.Context,
	ids []uuid.UUID,
) error {
	return workflow.objectData.DeleteByIds(ctx, ids)
}

func (workflow *projectDeletionWorkflow) DeleteProjectMemberships(
	ctx context.Context,
	projectID uuid.UUID,
) error {
	return workflow.db.WithContext(ctx).
		Table("project_users").
		Where("project_id = ?", projectID).
		Delete(nil).Error
}

func (workflow *projectDeletionWorkflow) DeleteProject(
	ctx context.Context,
	projectID uuid.UUID,
) error {
	return workflow.projects.DeleteByIds(ctx, []uuid.UUID{projectID})
}

func (workflow *projectFacilityUnlinkWorkflow) GetProjectFacilityLink(
	ctx context.Context,
	kind appprojectlink.Kind,
	linkID uuid.UUID,
) (*appprojectlink.Link, error) {
	switch kind {
	case appprojectlink.KindControlCabinet:
		item, err := domain.GetByID(ctx, workflow.controlCabinets, linkID)
		if err != nil {
			return nil, err
		}
		return &appprojectlink.Link{
			ID: item.ID, ProjectID: item.ProjectID, EntityID: item.ControlCabinetID,
		}, nil
	case appprojectlink.KindSPSController:
		item, err := domain.GetByID(ctx, workflow.spsControllers, linkID)
		if err != nil {
			return nil, err
		}
		return &appprojectlink.Link{
			ID: item.ID, ProjectID: item.ProjectID, EntityID: item.SPSControllerID,
		}, nil
	case appprojectlink.KindFieldDevice:
		item, err := domain.GetByID(ctx, workflow.fieldDevices, linkID)
		if err != nil {
			return nil, err
		}
		return &appprojectlink.Link{
			ID: item.ID, ProjectID: item.ProjectID, EntityID: item.FieldDeviceID,
		}, nil
	default:
		return nil, domain.ErrInvalidArgument
	}
}

func (workflow *projectFacilityUnlinkWorkflow) DeleteProjectFacilityLink(
	ctx context.Context,
	kind appprojectlink.Kind,
	linkID uuid.UUID,
) error {
	switch kind {
	case appprojectlink.KindControlCabinet:
		link, err := domain.GetByID(ctx, workflow.controlCabinets, linkID)
		if err != nil {
			return err
		}
		source := domainProject.AssignmentSource{
			Kind:           domainProject.AssignmentSourceControlCabinet,
			SourceEntityID: link.ControlCabinetID,
		}
		if pruner, ok := workflow.fieldDevices.(interface {
			RemoveAssignmentSourceAndPrune(
				context.Context,
				uuid.UUID,
				domainProject.AssignmentSource,
			) (bool, error)
		}); ok {
			if _, err := pruner.RemoveAssignmentSourceAndPrune(
				ctx,
				link.ProjectID,
				source,
			); err != nil {
				return err
			}
		}
		if pruner, ok := workflow.spsControllers.(interface {
			RemoveAssignmentSourceAndPrune(
				context.Context,
				uuid.UUID,
				domainProject.AssignmentSource,
			) (bool, error)
		}); ok {
			if _, err := pruner.RemoveAssignmentSourceAndPrune(
				ctx,
				link.ProjectID,
				source,
			); err != nil {
				return err
			}
		}
		return workflow.controlCabinets.DeleteByIds(ctx, []uuid.UUID{linkID})
	case appprojectlink.KindSPSController:
		link, err := domain.GetByID(ctx, workflow.spsControllers, linkID)
		if err != nil {
			return err
		}
		if pruner, ok := workflow.fieldDevices.(interface {
			RemoveAssignmentSourceAndPrune(
				context.Context,
				uuid.UUID,
				domainProject.AssignmentSource,
			) (bool, error)
		}); ok {
			if _, err := pruner.RemoveAssignmentSourceAndPrune(
				ctx,
				link.ProjectID,
				domainProject.AssignmentSource{
					Kind:           domainProject.AssignmentSourceSPSController,
					SourceEntityID: link.SPSControllerID,
				},
			); err != nil {
				return err
			}
		}
		if pruner, ok := workflow.spsControllers.(interface {
			RemoveAssignmentSourceAndPrune(
				context.Context,
				uuid.UUID,
				domainProject.AssignmentSource,
			) (bool, error)
		}); ok {
			handled, err := pruner.RemoveAssignmentSourceAndPrune(
				ctx,
				link.ProjectID,
				domainProject.AssignmentSource{
					Kind:           domainProject.AssignmentSourceExplicit,
					SourceEntityID: link.SPSControllerID,
				},
			)
			if err != nil {
				return err
			}
			if handled {
				return nil
			}
		}
		return workflow.spsControllers.DeleteByIds(ctx, []uuid.UUID{linkID})
	case appprojectlink.KindFieldDevice:
		link, err := domain.GetByID(ctx, workflow.fieldDevices, linkID)
		if err != nil {
			return err
		}
		if pruner, ok := workflow.fieldDevices.(interface {
			RemoveAssignmentSourceAndPrune(
				context.Context,
				uuid.UUID,
				domainProject.AssignmentSource,
			) (bool, error)
		}); ok {
			handled, err := pruner.RemoveAssignmentSourceAndPrune(
				ctx,
				link.ProjectID,
				domainProject.AssignmentSource{
					Kind:           domainProject.AssignmentSourceExplicit,
					SourceEntityID: link.FieldDeviceID,
				},
			)
			if err != nil {
				return err
			}
			if handled {
				return nil
			}
		}
		return workflow.fieldDevices.DeleteByIds(ctx, []uuid.UUID{linkID})
	default:
		return domain.ErrInvalidArgument
	}
}

func (w *projectControlCabinetRestoreWorkflow) RestoreControlCabinet(
	ctx context.Context,
	controlCabinetID uuid.UUID,
	request domainHistory.RestoreControlCabinetRequest,
) (*domainHistory.RestoreResult, error) {
	return w.restorer.RestoreControlCabinet(ctx, controlCabinetID, request)
}

func (w *projectControlCabinetRestoreWorkflow) GetByControlCabinetIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectControlCabinet, error) {
	return w.links.GetByControlCabinetIDs(ctx, ids)
}

type bacnetObjectCreateOutboxWorkflow struct {
	appbacnetobject.CreateWorkflow
	links appbacnetobject.ProjectLinkReader
}

type bacnetObjectUpdateOutboxWorkflow struct {
	appbacnetobject.UpdateWorkflow
	links  appbacnetobject.ProjectLinkReader
	owners appbacnetobject.ObjectDataOwnerReader
}

func (w *bacnetObjectUpdateOutboxWorkflow) GetByFieldDeviceIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	return w.links.GetByFieldDeviceIDs(ctx, ids)
}

func (w *bacnetObjectUpdateOutboxWorkflow) GetByBacnetObjectIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]domainObjectData.BacnetObjectOwner, error) {
	return w.owners.GetByBacnetObjectIDs(ctx, ids)
}

type bacnetAlarmValuesOutboxWorkflow struct {
	appbacnetobject.AlarmValueStore
	db      *gorm.DB
	objects appbacnetobject.BacnetObjectStateReader
	links   appbacnetobject.ProjectLinkReader
	owners  appbacnetobject.ObjectDataOwnerReader
}

func (w *bacnetAlarmValuesOutboxWorkflow) GetAlarmSelection(
	ctx context.Context,
	bacnetObjectID uuid.UUID,
) (*appbacnetobject.AlarmSelection, error) {
	var selection appbacnetobject.AlarmSelection
	err := w.db.WithContext(ctx).
		Table("bacnet_objects").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id AS bacnet_object_id, alarm_type_id").
		Where("id = ?", bacnetObjectID).
		Take(&selection).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &selection, nil
}

func (w *bacnetAlarmValuesOutboxWorkflow) GetAlarmTypeFields(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainFacility.AlarmTypeField, error) {
	if len(ids) == 0 {
		return []*domainFacility.AlarmTypeField{}, nil
	}
	var fields []*domainFacility.AlarmTypeField
	err := w.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id IN ?", ids).
		Find(&fields).Error
	return fields, err
}

func (w *bacnetAlarmValuesOutboxWorkflow) GetByIds(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainFacility.BacnetObject, error) {
	return w.objects.GetByIds(ctx, ids)
}

func (w *bacnetAlarmValuesOutboxWorkflow) GetByFieldDeviceIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	return w.links.GetByFieldDeviceIDs(ctx, ids)
}

func (w *bacnetAlarmValuesOutboxWorkflow) GetByBacnetObjectIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]domainObjectData.BacnetObjectOwner, error) {
	return w.owners.GetByBacnetObjectIDs(ctx, ids)
}

func (w *bacnetObjectCreateOutboxWorkflow) GetByFieldDeviceIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	return w.links.GetByFieldDeviceIDs(ctx, ids)
}

type fieldDeviceUpdateOutboxWorkflow struct {
	appfielddevice.UpdateWorkflow
	links       appfielddevice.ProjectLinkReader
	assignments *projectservice.ProjectFacilityLinkService
}

type fieldDeviceBulkUpdateOutboxWorkflow struct {
	appfielddevice.BulkUpdateExecutor
	links appfielddevice.ProjectLinkReader
}

type fieldDeviceDeleteOutboxWorkflow struct {
	appfielddevice.DeleteWorkflow
	links appfielddevice.ProjectLinkReader
}

type fieldDeviceBulkDeleteOutboxWorkflow struct {
	appfielddevice.BulkDeleteWorkflow
	links appfielddevice.ProjectLinkReader
}

func (w *fieldDeviceUpdateOutboxWorkflow) ReconcileFieldDeviceMove(
	ctx context.Context,
	fieldDeviceID uuid.UUID,
	fromSystemTypeID uuid.UUID,
	toSystemTypeID uuid.UUID,
) ([]uuid.UUID, error) {
	if w.assignments == nil {
		return nil, nil
	}
	return w.assignments.ReconcileFieldDeviceMove(
		ctx,
		fieldDeviceID,
		fromSystemTypeID,
		toSystemTypeID,
	)
}

func (w *fieldDeviceBulkUpdateOutboxWorkflow) GetByFieldDeviceIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	return w.links.GetByFieldDeviceIDs(ctx, ids)
}

type controlCabinetUpdateOutboxWorkflow struct {
	appcontrolcabinet.UpdateWorkflow
	links appcontrolcabinet.ProjectLinkReader
	db    *gorm.DB
}

func (w *controlCabinetUpdateOutboxWorkflow) CountSPSControllers(
	ctx context.Context,
	controlCabinetID uuid.UUID,
) (int64, error) {
	var count int64
	err := w.db.WithContext(ctx).
		Model(&domainFacility.SPSController{}).
		Where("control_cabinet_id = ?", controlCabinetID).
		Count(&count).Error
	return count, err
}

type controlCabinetDeleteOutboxWorkflow struct {
	appcontrolcabinet.DeleteWorkflow
	cleaner *hierarchyDeleteCleaner
	db      *gorm.DB
}

func (w *controlCabinetDeleteOutboxWorkflow) GetByControlCabinetIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectControlCabinet, error) {
	if len(ids) == 0 {
		return []*domainProject.ProjectControlCabinet{}, nil
	}
	var projectIDs []uuid.UUID
	if err := w.db.WithContext(ctx).Raw(`
		WITH descendant_fields AS (
			SELECT field.id
			FROM field_devices AS field
			JOIN sps_controller_system_types AS assignment
				ON assignment.id = field.sps_controller_system_type_id
			JOIN sps_controllers AS controller
				ON controller.id = assignment.sps_controller_id
			WHERE controller.control_cabinet_id IN ?
		),
		descendant_bacnet AS (
			SELECT object.id
			FROM bacnet_objects AS object
			WHERE object.field_device_id IN (SELECT id FROM descendant_fields)
		),
		affected_bacnet AS (
			SELECT id FROM descendant_bacnet
			UNION
			SELECT object.id
			FROM bacnet_objects AS object
			WHERE object.software_reference_id IN (SELECT id FROM descendant_bacnet)
		)
		SELECT project_id
		FROM project_control_cabinets
		WHERE control_cabinet_id IN ?
		UNION
		SELECT link.project_id
		FROM project_sps_controllers AS link
		JOIN sps_controllers AS controller
			ON controller.id = link.sps_controller_id
		WHERE controller.control_cabinet_id IN ?
		UNION
		SELECT project_id
		FROM project_field_devices
		WHERE field_device_id IN (SELECT id FROM descendant_fields)
		UNION
		SELECT link.project_id
		FROM project_field_devices AS link
		JOIN bacnet_objects AS object ON object.field_device_id = link.field_device_id
		WHERE object.id IN (SELECT id FROM affected_bacnet)
		UNION
		SELECT object_data.project_id
		FROM object_data
		JOIN object_data_bacnet_objects AS link
			ON link.object_data_id = object_data.id
		WHERE object_data.project_id IS NOT NULL
		  AND link.bacnet_object_id IN (SELECT id FROM affected_bacnet)
		ORDER BY project_id
	`, ids, ids, ids).Scan(&projectIDs).Error; err != nil {
		return nil, err
	}
	links := make([]*domainProject.ProjectControlCabinet, len(projectIDs))
	for index, projectID := range projectIDs {
		links[index] = &domainProject.ProjectControlCabinet{
			ProjectID:        projectID,
			ControlCabinetID: ids[0],
		}
	}
	return links, nil
}

func (w *controlCabinetDeleteOutboxWorkflow) DeleteByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	return w.cleaner.deleteControlCabinet(ctx, id)
}

func (w *controlCabinetUpdateOutboxWorkflow) GetByControlCabinetIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectControlCabinet, error) {
	return w.links.GetByControlCabinetIDs(ctx, ids)
}

type spsControllerUpdateOutboxWorkflow struct {
	appspscontroller.UpdateWorkflow
	links       appspscontroller.ProjectLinkReader
	assignments *projectservice.ProjectFacilityLinkService
}

type spsControllerDeleteOutboxWorkflow struct {
	appspscontroller.DeleteWorkflow
	cleaner *hierarchyDeleteCleaner
	db      *gorm.DB
}

type spsControllerSystemTypeDeleteWorkflow struct {
	appspscontroller.DeleteSystemTypeWorkflow
	cleaner *hierarchyDeleteCleaner
	db      *gorm.DB
}

func (w *spsControllerUpdateOutboxWorkflow) ReconcileSPSControllerMove(
	ctx context.Context,
	spsControllerID uuid.UUID,
	fromControlCabinetID uuid.UUID,
	toControlCabinetID uuid.UUID,
) ([]uuid.UUID, error) {
	if w.assignments == nil {
		return nil, nil
	}
	return w.assignments.ReconcileSPSControllerMove(
		ctx,
		spsControllerID,
		fromControlCabinetID,
		toControlCabinetID,
	)
}

func (w *spsControllerSystemTypeDeleteWorkflow) DeleteByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	return w.cleaner.deleteSPSControllerSystemTypes(ctx, []uuid.UUID{id})
}

func (w *spsControllerSystemTypeDeleteWorkflow) GetDeleteProjectScope(
	ctx context.Context,
	systemTypeID uuid.UUID,
) (uuid.UUID, []uuid.UUID, error) {
	var ownerID uuid.UUID
	if err := w.db.WithContext(ctx).
		Model(&domainFacility.SPSControllerSystemType{}).
		Where("id = ?", systemTypeID).
		Pluck("sps_controller_id", &ownerID).Error; err != nil {
		return uuid.Nil, nil, err
	}
	if ownerID == uuid.Nil {
		return uuid.Nil, nil, nil
	}

	var projectIDs []uuid.UUID
	if err := w.db.WithContext(ctx).Raw(`
		WITH descendant_fields AS (
			SELECT id
			FROM field_devices
			WHERE sps_controller_system_type_id = ?
		),
		descendant_bacnet AS (
			SELECT object.id
			FROM bacnet_objects AS object
			WHERE object.field_device_id IN (SELECT id FROM descendant_fields)
		),
		affected_bacnet AS (
			SELECT id FROM descendant_bacnet
			UNION
			SELECT object.id
			FROM bacnet_objects AS object
			WHERE object.software_reference_id IN (SELECT id FROM descendant_bacnet)
		)
		SELECT project_id
		FROM project_sps_controllers
		WHERE sps_controller_id = ?
		UNION
		SELECT project_id
		FROM project_field_devices
		WHERE field_device_id IN (SELECT id FROM descendant_fields)
		UNION
		SELECT link.project_id
		FROM project_field_devices AS link
		JOIN bacnet_objects AS object ON object.field_device_id = link.field_device_id
		WHERE object.id IN (SELECT id FROM affected_bacnet)
		UNION
		SELECT object_data.project_id
		FROM object_data
		JOIN object_data_bacnet_objects AS link
			ON link.object_data_id = object_data.id
		WHERE object_data.project_id IS NOT NULL
		  AND link.bacnet_object_id IN (SELECT id FROM affected_bacnet)
		ORDER BY project_id
	`, systemTypeID, ownerID).Scan(&projectIDs).Error; err != nil {
		return uuid.Nil, nil, err
	}
	return ownerID, projectIDs, nil
}

func (w *spsControllerUpdateOutboxWorkflow) GetBySPSControllerIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectSPSController, error) {
	return w.links.GetBySPSControllerIDs(ctx, ids)
}

func (w *spsControllerDeleteOutboxWorkflow) GetBySPSControllerIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectSPSController, error) {
	if len(ids) == 0 {
		return []*domainProject.ProjectSPSController{}, nil
	}
	var projectIDs []uuid.UUID
	if err := w.db.WithContext(ctx).Raw(`
		WITH descendant_fields AS (
			SELECT field.id
			FROM field_devices AS field
			JOIN sps_controller_system_types AS assignment
				ON assignment.id = field.sps_controller_system_type_id
			WHERE assignment.sps_controller_id IN ?
		),
		descendant_bacnet AS (
			SELECT object.id
			FROM bacnet_objects AS object
			WHERE object.field_device_id IN (SELECT id FROM descendant_fields)
		),
		affected_bacnet AS (
			SELECT id FROM descendant_bacnet
			UNION
			SELECT object.id
			FROM bacnet_objects AS object
			WHERE object.software_reference_id IN (SELECT id FROM descendant_bacnet)
		)
		SELECT project_id
		FROM project_sps_controllers
		WHERE sps_controller_id IN ?
		UNION
		SELECT project_id
		FROM project_field_devices
		WHERE field_device_id IN (SELECT id FROM descendant_fields)
		UNION
		SELECT link.project_id
		FROM project_field_devices AS link
		JOIN bacnet_objects AS object ON object.field_device_id = link.field_device_id
		WHERE object.id IN (SELECT id FROM affected_bacnet)
		UNION
		SELECT object_data.project_id
		FROM object_data
		JOIN object_data_bacnet_objects AS link
			ON link.object_data_id = object_data.id
		WHERE object_data.project_id IS NOT NULL
		  AND link.bacnet_object_id IN (SELECT id FROM affected_bacnet)
		ORDER BY project_id
	`, ids, ids).Scan(&projectIDs).Error; err != nil {
		return nil, err
	}
	links := make([]*domainProject.ProjectSPSController, len(projectIDs))
	for index, projectID := range projectIDs {
		links[index] = &domainProject.ProjectSPSController{
			ProjectID:       projectID,
			SPSControllerID: ids[0],
		}
	}
	return links, nil
}

func (w *spsControllerDeleteOutboxWorkflow) DeleteByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	return w.cleaner.deleteSPSControllers(ctx, []uuid.UUID{id})
}

func (w *fieldDeviceDeleteOutboxWorkflow) GetByFieldDeviceIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	return w.links.GetByFieldDeviceIDs(ctx, ids)
}

func (w *fieldDeviceBulkDeleteOutboxWorkflow) GetByFieldDeviceIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	return w.links.GetByFieldDeviceIDs(ctx, ids)
}

func (w *fieldDeviceUpdateOutboxWorkflow) GetByFieldDeviceIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	return w.links.GetByFieldDeviceIDs(ctx, ids)
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

func newProjectApplicationServices(
	gormDB *gorm.DB,
) *appproject.Services {
	actor := func(ctx context.Context) *uuid.UUID {
		actorID, _ := auditctx.ActorID(ctx)
		return actorID
	}
	transactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appproject.DeletionWorkflow, error) {
		txDB, err := infratransaction.GormDB(unit)
		if err != nil {
			return nil, err
		}
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, err
		}
		return &projectDeletionWorkflow{
			db:         txDB,
			projects:   txRepos.Project,
			objectData: txRepos.FacilityObjectData,
		}, nil
	}

	return &appproject.Services{
		Delete: appproject.NewDeleteHandler(appproject.DeleteDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: transactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Actor:               actor,
		}),
	}
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
	projectFacilityUnlink := appprojectlink.NewHandler(appprojectlink.Dependencies{
		TransactionRunner: infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: func(
			unit apptransaction.UnitOfWork,
		) (appprojectlink.Workflow, error) {
			txRepos, err := repositoriesFromUnit(unit)
			if err != nil {
				return nil, fmt.Errorf("project facility unlink transaction unit: %w", err)
			}
			return &projectFacilityUnlinkWorkflow{
				controlCabinets: txRepos.ProjectControlCabinets,
				spsControllers:  txRepos.ProjectSPSControllers,
				fieldDevices:    txRepos.ProjectFieldDevices,
			}, nil
		},
		HistoryBatch: auditctx.WithBatchID,
		Actor:        actor,
	})
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
		return &bacnetObjectUpdateOutboxWorkflow{
			UpdateWorkflow: txServices.BacnetObject,
			links:          txRepos.ProjectFieldDevices,
			owners:         txRepos.FacilityBacnetObjectOwners,
		}, nil
	}
	bacnetObjectCreateTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appbacnetobject.CreateWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("BACnetObject application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return &bacnetObjectCreateOutboxWorkflow{
			CreateWorkflow: txServices.BacnetObject,
			links:          txRepos.ProjectFieldDevices,
		}, nil
	}
	bacnetAlarmValuesTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appbacnetobject.ReplaceAlarmValuesWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("BACnet alarm-value application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		txDB, err := infratransaction.GormDB(unit)
		if err != nil {
			return nil, fmt.Errorf("BACnet alarm-value transaction database: %w", err)
		}
		return &bacnetAlarmValuesOutboxWorkflow{
			AlarmValueStore: txServices.BacnetAlarmValue,
			db:              txDB,
			objects:         txRepos.FacilityBacnetObjects,
			links:           txRepos.ProjectFieldDevices,
			owners:          txRepos.FacilityBacnetObjectOwners,
		}, nil
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
		txDB, err := infratransaction.GormDB(unit)
		if err != nil {
			return nil, fmt.Errorf("resolve ControlCabinet update transaction DB: %w", err)
		}
		return &controlCabinetUpdateOutboxWorkflow{
			UpdateWorkflow: txServices.ControlCabinet,
			links:          txRepos.ProjectControlCabinets,
			db:             txDB,
		}, nil
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
		txFacilityServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		txProjectServices := projectservice.NewServices(
			buildProjectDependencies(txRepos, txFacilityServices),
		)
		txDB, err := infratransaction.GormDB(unit)
		if err != nil {
			return nil, fmt.Errorf("resolve ControlCabinet clone transaction DB: %w", err)
		}
		return &globalControlCabinetCloneWorkflow{
			clones:        txFacilityServices.ControlCabinet,
			facilityLinks: txProjectServices.FacilityLink,
			db:            txDB,
		}, nil
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
		txDB, err := infratransaction.GormDB(unit)
		if err != nil {
			return nil, fmt.Errorf("resolve project ControlCabinet clone transaction DB: %w", err)
		}
		return &projectFacilityCloneWorkflow{
			ProjectFacilityLinkService: txProjectServices.FacilityLink,
			db:                         txDB,
			sourceKind:                 projectCloneControlCabinet,
		}, nil
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
		txDB, err := infratransaction.GormDB(unit)
		if err != nil {
			return nil, fmt.Errorf("resolve ControlCabinet delete transaction DB: %w", err)
		}
		cleaner := &hierarchyDeleteCleaner{db: txDB, repos: txRepos}
		return &controlCabinetDeleteOutboxWorkflow{
			DeleteWorkflow: txServices.ControlCabinet,
			cleaner:        cleaner,
			db:             txDB,
		}, nil
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
		TransactionRunner: infratransaction.NewGormRunnerWithIsolation(
			gormDB,
			sql.LevelRepeatableRead,
		),
		TransactionWorkflow: controlCabinetCloneTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         controlCabinetReportError,
	})
	controlCabinetProjectClone := appcontrolcabinet.NewCloneForProjectHandler(
		appcontrolcabinet.CloneForProjectDependencies{
			TransactionRunner: infratransaction.NewGormRunnerWithIsolation(
				gormDB,
				sql.LevelRepeatableRead,
			),
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
	controlCabinetProjectRestoreTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appcontrolcabinet.RestoreForProjectWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("project ControlCabinet restore transaction unit: %w", err)
		}
		return &projectControlCabinetRestoreWorkflow{
			restorer: txRepos.History,
			links:    txRepos.ProjectControlCabinets,
		}, nil
	}
	controlCabinetProjectRestore := appcontrolcabinet.NewRestoreForProjectHandler(
		appcontrolcabinet.RestoreForProjectDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: controlCabinetProjectRestoreTransactionWorkflow,
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
		txDB, err := infratransaction.GormDB(unit)
		if err != nil {
			return nil, fmt.Errorf("resolve project SPSController clone transaction DB: %w", err)
		}
		return &projectFacilityCloneWorkflow{
			ProjectFacilityLinkService: txProjectServices.FacilityLink,
			db:                         txDB,
			sourceKind:                 projectCloneSPSController,
		}, nil
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
	bulkUpdateTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appfielddevice.BulkUpdateWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("FieldDevice bulk-update transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return &fieldDeviceBulkUpdateOutboxWorkflow{
			BulkUpdateExecutor: txServices.FieldDevice,
			links:              txRepos.ProjectFieldDevices,
		}, nil
	}
	bulkUpdate := appfielddevice.NewBulkUpdateHandler(appfielddevice.BulkUpdateDependencies{
		Executor:            legacy.FieldDevice,
		TransactionRunner:   infratransaction.NewGormRunner(gormDB),
		TransactionWorkflow: bulkUpdateTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectFieldDevices,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         fieldDeviceReportError,
		MapTransactionError: facilityrepo.MapFieldDeviceWriteError,
	})

	transactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appfielddevice.UpdateWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("FieldDevice application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		txProjectServices := projectservice.NewServices(
			buildProjectDependencies(txRepos, txServices),
		)
		return &fieldDeviceUpdateOutboxWorkflow{
			UpdateWorkflow: txServices.FieldDevice,
			links:          txRepos.ProjectFieldDevices,
			assignments:    txProjectServices.FacilityLink,
		}, nil
	}
	deleteTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appfielddevice.DeleteWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("FieldDevice delete application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return &fieldDeviceDeleteOutboxWorkflow{
			DeleteWorkflow: txServices.FieldDevice,
			links:          txRepos.ProjectFieldDevices,
		}, nil
	}
	bulkDeleteTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appfielddevice.BulkDeleteWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("FieldDevice bulk-delete application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		return &fieldDeviceBulkDeleteOutboxWorkflow{
			BulkDeleteWorkflow: txServices.FieldDevice,
			links:              txRepos.ProjectFieldDevices,
		}, nil
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
		txProjectServices := projectservice.NewServices(
			buildProjectDependencies(txRepos, txServices),
		)
		return &spsControllerUpdateOutboxWorkflow{
			UpdateWorkflow: txServices.SPSController,
			links:          txRepos.ProjectSPSControllers,
			assignments:    txProjectServices.FacilityLink,
		}, nil
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
		txDB, err := infratransaction.GormDB(unit)
		if err != nil {
			return nil, fmt.Errorf("resolve SPSControllerSystemType clone transaction DB: %w", err)
		}
		return &globalSPSControllerSystemTypeCloneWorkflow{
			clones: txServices.SPSControllerSystemType,
			db:     txDB,
		}, nil
	}
	spsSystemTypeDeleteTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appspscontroller.DeleteSystemTypeWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("SPSControllerSystemType delete application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		txDB, err := infratransaction.GormDB(unit)
		if err != nil {
			return nil, fmt.Errorf("resolve SPSControllerSystemType delete transaction DB: %w", err)
		}
		return &spsControllerSystemTypeDeleteWorkflow{
			DeleteSystemTypeWorkflow: txServices.SPSControllerSystemType,
			cleaner:                  &hierarchyDeleteCleaner{db: txDB, repos: txRepos},
			db:                       txDB,
		}, nil
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
		txDB, err := infratransaction.GormDB(unit)
		if err != nil {
			return nil, fmt.Errorf("resolve project SPSController clone transaction DB: %w", err)
		}
		return &projectFacilityCloneWorkflow{
			ProjectFacilityLinkService: txProjectServices.FacilityLink,
			db:                         txDB,
			sourceKind:                 projectCloneSPSController,
		}, nil
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
		txDB, err := infratransaction.GormDB(unit)
		if err != nil {
			return nil, fmt.Errorf("resolve project SPSControllerSystemType clone transaction DB: %w", err)
		}
		return &projectFacilityCloneWorkflow{
			ProjectFacilityLinkService: txProjectServices.FacilityLink,
			db:                         txDB,
			sourceKind:                 projectCloneSPSControllerSystemType,
		}, nil
	}
	spsDeleteTransactionWorkflow := func(
		unit apptransaction.UnitOfWork,
	) (appspscontroller.DeleteWorkflow, error) {
		txRepos, err := repositoriesFromUnit(unit)
		if err != nil {
			return nil, fmt.Errorf("SPSController delete application transaction unit: %w", err)
		}
		txServices := facilityservice.NewServices(buildFacilityRepositories(txRepos))
		txDB, err := infratransaction.GormDB(unit)
		if err != nil {
			return nil, fmt.Errorf("resolve SPSController delete transaction DB: %w", err)
		}
		cleaner := &hierarchyDeleteCleaner{db: txDB, repos: txRepos}
		return &spsControllerDeleteOutboxWorkflow{
			DeleteWorkflow: txServices.SPSController,
			cleaner:        cleaner,
			db:             txDB,
		}, nil
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
		TransactionRunner: infratransaction.NewGormRunnerWithIsolation(
			gormDB,
			sql.LevelRepeatableRead,
		),
		TransactionWorkflow: spsCloneTransactionWorkflow,
		HistoryBatch:        auditctx.WithBatchID,
		ProjectLinks:        repos.ProjectSPSControllers,
		Dispatcher:          dispatcher,
		Actor:               actor,
		ReportError:         spsControllerReportError,
	})
	spsSystemTypeClone := appspscontroller.NewCloneSystemTypeHandler(
		appspscontroller.CloneSystemTypeDependencies{
			TransactionRunner: infratransaction.NewGormRunnerWithIsolation(
				gormDB,
				sql.LevelRepeatableRead,
			),
			TransactionWorkflow: spsSystemTypeCloneTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         spsControllerReportError,
		},
	)
	spsSystemTypeDelete := appspscontroller.NewDeleteSystemTypeHandler(
		appspscontroller.DeleteSystemTypeDependencies{
			TransactionRunner:   infratransaction.NewGormRunner(gormDB),
			TransactionWorkflow: spsSystemTypeDeleteTransactionWorkflow,
			HistoryBatch:        auditctx.WithBatchID,
			Dispatcher:          dispatcher,
			Actor:               actor,
			ReportError:         spsControllerReportError,
		},
	)
	spsProjectClone := appspscontroller.NewCloneForProjectHandler(
		appspscontroller.CloneForProjectDependencies{
			TransactionRunner: infratransaction.NewGormRunnerWithIsolation(
				gormDB,
				sql.LevelRepeatableRead,
			),
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
			TransactionRunner: infratransaction.NewGormRunnerWithIsolation(
				gormDB,
				sql.LevelRepeatableRead,
			),
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
		ProjectLink: projectFacilityUnlink,
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
		Global: apphistory.NewGlobalHistoryService(
			apphistory.GlobalHistoryDependencies{
				History: repos.History,
				Users:   repos.User,
				Actor:   actor,
			},
		),
		ProjectTimeline: apphistory.NewProjectTimelineHandler(
			apphistory.ProjectTimelineDependencies{
				Access:   &projectHistoryAccessScope{access: projectAccess},
				Timeline: repos.History,
				Actor:    actor,
			},
		),
	}
}
