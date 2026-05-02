package historycapture

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/repository/historysql"
	"github.com/google/uuid"
)

type ProjectRepository struct {
	domainProject.ProjectRepository
	audit audit[domainProject.Project]
}

func WrapProject(next domainProject.ProjectRepository, store *historysql.Store) domainProject.ProjectRepository {
	return &ProjectRepository{ProjectRepository: next, audit: newAudit[domainProject.Project]("projects", store)}
}

func (r *ProjectRepository) Create(ctx context.Context, entity *domainProject.Project) error {
	return r.audit.create(ctx, r.ProjectRepository, entity)
}

func (r *ProjectRepository) Update(ctx context.Context, entity *domainProject.Project) error {
	return r.audit.update(ctx, r.ProjectRepository, entity)
}

func (r *ProjectRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.ProjectRepository, ids)
}

func (r *ProjectRepository) GetPaginatedListForUser(ctx context.Context, params domain.PaginationParams, userID uuid.UUID) (*domain.PaginatedList[domainProject.Project], error) {
	return r.ProjectRepository.GetPaginatedListForUser(ctx, params, userID)
}

func (r *ProjectRepository) GetPaginatedListWithStatus(ctx context.Context, params domain.PaginationParams, status *domainProject.ProjectStatus) (*domain.PaginatedList[domainProject.Project], error) {
	return r.ProjectRepository.GetPaginatedListWithStatus(ctx, params, status)
}

func (r *ProjectRepository) GetPaginatedListForUserWithStatus(ctx context.Context, params domain.PaginationParams, userID uuid.UUID, status *domainProject.ProjectStatus) (*domain.PaginatedList[domainProject.Project], error) {
	return r.ProjectRepository.GetPaginatedListForUserWithStatus(ctx, params, userID, status)
}

type ProjectControlCabinetRepository struct {
	domainProject.ProjectControlCabinetRepository
	audit audit[domainProject.ProjectControlCabinet]
	store *historysql.Store
}

func WrapProjectControlCabinet(next domainProject.ProjectControlCabinetRepository, store *historysql.Store) domainProject.ProjectControlCabinetRepository {
	return &ProjectControlCabinetRepository{ProjectControlCabinetRepository: next, audit: newAudit[domainProject.ProjectControlCabinet]("project_control_cabinets", store), store: store}
}

func (r *ProjectControlCabinetRepository) Create(ctx context.Context, entity *domainProject.ProjectControlCabinet) error {
	return r.audit.create(ctx, r.ProjectControlCabinetRepository, entity)
}
func (r *ProjectControlCabinetRepository) Update(ctx context.Context, entity *domainProject.ProjectControlCabinet) error {
	return r.audit.update(ctx, r.ProjectControlCabinetRepository, entity)
}
func (r *ProjectControlCabinetRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.ProjectControlCabinetRepository, ids)
}
func (r *ProjectControlCabinetRepository) DeleteByControlCabinetIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.audit.deleteRows(ctx,
		func(ctx context.Context) (map[uuid.UUID]domainHistory.JSONB, error) {
			return r.store.LoadRowsWhere(ctx, "project_control_cabinets", "control_cabinet_id IN ?", ids)
		},
		func(ctx context.Context) error {
			return r.ProjectControlCabinetRepository.DeleteByControlCabinetIDs(ctx, ids)
		},
	)
}

type ProjectSPSControllerRepository struct {
	domainProject.ProjectSPSControllerRepository
	audit audit[domainProject.ProjectSPSController]
	store *historysql.Store
}

func WrapProjectSPSController(next domainProject.ProjectSPSControllerRepository, store *historysql.Store) domainProject.ProjectSPSControllerRepository {
	return &ProjectSPSControllerRepository{ProjectSPSControllerRepository: next, audit: newAudit[domainProject.ProjectSPSController]("project_sps_controllers", store), store: store}
}

func (r *ProjectSPSControllerRepository) Create(ctx context.Context, entity *domainProject.ProjectSPSController) error {
	return r.audit.create(ctx, r.ProjectSPSControllerRepository, entity)
}
func (r *ProjectSPSControllerRepository) Update(ctx context.Context, entity *domainProject.ProjectSPSController) error {
	return r.audit.update(ctx, r.ProjectSPSControllerRepository, entity)
}
func (r *ProjectSPSControllerRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.ProjectSPSControllerRepository, ids)
}
func (r *ProjectSPSControllerRepository) DeleteBySPSControllerIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.audit.deleteRows(ctx,
		func(ctx context.Context) (map[uuid.UUID]domainHistory.JSONB, error) {
			return r.store.LoadRowsWhere(ctx, "project_sps_controllers", "sps_controller_id IN ?", ids)
		},
		func(ctx context.Context) error {
			return r.ProjectSPSControllerRepository.DeleteBySPSControllerIDs(ctx, ids)
		},
	)
}
func (r *ProjectSPSControllerRepository) BulkCreate(ctx context.Context, entities []*domainProject.ProjectSPSController, batchSize int) error {
	creator, ok := r.ProjectSPSControllerRepository.(interface {
		BulkCreate(context.Context, []*domainProject.ProjectSPSController, int) error
	})
	if !ok {
		for _, entity := range entities {
			if err := r.Create(ctx, entity); err != nil {
				return err
			}
		}
		return nil
	}
	return r.audit.bulkCreate(ctx,
		func(ctx context.Context) error { return creator.BulkCreate(ctx, entities, batchSize) },
		func() []uuid.UUID { return idsOf(entities) },
	)
}
func (r *ProjectSPSControllerRepository) BulkCreateBySPSControllerIDs(ctx context.Context, projectID uuid.UUID, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}
	creator, ok := r.ProjectSPSControllerRepository.(interface {
		BulkCreateBySPSControllerIDs(context.Context, uuid.UUID, []uuid.UUID) error
	})
	if !ok {
		return nil
	}
	if err := creator.BulkCreateBySPSControllerIDs(ctx, projectID, spsControllerIDs); err != nil {
		return err
	}
	rows, err := r.store.LoadRowsWhere(ctx, "project_sps_controllers", "project_id = ? AND sps_controller_id IN ?", projectID, spsControllerIDs)
	if err != nil {
		return err
	}
	for id := range rows {
		if err := r.audit.recordCreated(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

type ProjectFieldDeviceRepository struct {
	domainProject.ProjectFieldDeviceRepository
	audit audit[domainProject.ProjectFieldDevice]
	store *historysql.Store
}

func WrapProjectFieldDevice(next domainProject.ProjectFieldDeviceRepository, store *historysql.Store) domainProject.ProjectFieldDeviceRepository {
	return &ProjectFieldDeviceRepository{ProjectFieldDeviceRepository: next, audit: newAudit[domainProject.ProjectFieldDevice]("project_field_devices", store), store: store}
}

func (r *ProjectFieldDeviceRepository) Create(ctx context.Context, entity *domainProject.ProjectFieldDevice) error {
	return r.audit.create(ctx, r.ProjectFieldDeviceRepository, entity)
}
func (r *ProjectFieldDeviceRepository) Update(ctx context.Context, entity *domainProject.ProjectFieldDevice) error {
	return r.audit.update(ctx, r.ProjectFieldDeviceRepository, entity)
}
func (r *ProjectFieldDeviceRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	return r.audit.deleteByIds(ctx, r.ProjectFieldDeviceRepository, ids)
}
func (r *ProjectFieldDeviceRepository) DeleteByFieldDeviceIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.audit.deleteRows(ctx,
		func(ctx context.Context) (map[uuid.UUID]domainHistory.JSONB, error) {
			return r.store.LoadRowsWhere(ctx, "project_field_devices", "field_device_id IN ?", ids)
		},
		func(ctx context.Context) error {
			return r.ProjectFieldDeviceRepository.DeleteByFieldDeviceIDs(ctx, ids)
		},
	)
}
func (r *ProjectFieldDeviceRepository) BulkCreate(ctx context.Context, entities []*domainProject.ProjectFieldDevice, batchSize int) error {
	creator, ok := r.ProjectFieldDeviceRepository.(interface {
		BulkCreate(context.Context, []*domainProject.ProjectFieldDevice, int) error
	})
	if !ok {
		for _, entity := range entities {
			if err := r.Create(ctx, entity); err != nil {
				return err
			}
		}
		return nil
	}
	return r.audit.bulkCreate(ctx,
		func(ctx context.Context) error { return creator.BulkCreate(ctx, entities, batchSize) },
		func() []uuid.UUID { return idsOf(entities) },
	)
}
func (r *ProjectFieldDeviceRepository) BulkCreateByFieldDeviceIDs(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}
	creator, ok := r.ProjectFieldDeviceRepository.(interface {
		BulkCreateByFieldDeviceIDs(context.Context, uuid.UUID, []uuid.UUID) error
	})
	if !ok {
		return nil
	}
	if err := creator.BulkCreateByFieldDeviceIDs(ctx, projectID, fieldDeviceIDs); err != nil {
		return err
	}
	return r.recordRowsByFieldDeviceIDs(ctx, projectID, fieldDeviceIDs)
}
func (r *ProjectFieldDeviceRepository) BulkCreateBySPSControllerSystemTypeIDs(ctx context.Context, projectID uuid.UUID, systemTypeIDs []uuid.UUID) error {
	if len(systemTypeIDs) == 0 {
		return nil
	}
	creator, ok := r.ProjectFieldDeviceRepository.(interface {
		BulkCreateBySPSControllerSystemTypeIDs(context.Context, uuid.UUID, []uuid.UUID) error
	})
	if !ok {
		return nil
	}
	if err := creator.BulkCreateBySPSControllerSystemTypeIDs(ctx, projectID, systemTypeIDs); err != nil {
		return err
	}
	rows, err := r.store.LoadRowsWhere(ctx, "project_field_devices", `
		project_id = ?
		AND field_device_id IN (
			SELECT id FROM field_devices WHERE sps_controller_system_type_id IN ?
		)`, projectID, systemTypeIDs)
	if err != nil {
		return err
	}
	for id := range rows {
		if err := r.audit.recordCreated(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
func (r *ProjectFieldDeviceRepository) DeleteBySPSControllerSystemTypeIDs(ctx context.Context, systemTypeIDs []uuid.UUID) error {
	if len(systemTypeIDs) == 0 {
		return nil
	}
	return r.audit.deleteRows(ctx,
		func(ctx context.Context) (map[uuid.UUID]domainHistory.JSONB, error) {
			return r.store.LoadRowsWhere(ctx, "project_field_devices", `
				field_device_id IN (
					SELECT id FROM field_devices WHERE sps_controller_system_type_id IN ?
				)`, systemTypeIDs)
		},
		func(ctx context.Context) error {
			deleter, ok := r.ProjectFieldDeviceRepository.(interface {
				DeleteBySPSControllerSystemTypeIDs(context.Context, []uuid.UUID) error
			})
			if ok {
				return deleter.DeleteBySPSControllerSystemTypeIDs(ctx, systemTypeIDs)
			}
			return nil
		},
	)
}
func (r *ProjectFieldDeviceRepository) recordRowsByFieldDeviceIDs(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}
	rows, err := r.store.LoadRowsWhere(ctx, "project_field_devices", "project_id = ? AND field_device_id IN ?", projectID, fieldDeviceIDs)
	if err != nil {
		return err
	}
	for id := range rows {
		if err := r.audit.recordCreated(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
