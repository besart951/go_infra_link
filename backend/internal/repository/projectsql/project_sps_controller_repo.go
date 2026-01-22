package projectsql

import (
"github.com/besart951/go_infra_link/backend/internal/domain"
"github.com/besart951/go_infra_link/backend/internal/domain/project"
"github.com/google/uuid"
"gorm.io/gorm"
)

type projectSPSControllerRepo struct {
db *gorm.DB
}

func NewProjectSPSControllerRepository(db *gorm.DB) project.ProjectSPSControllerRepository {
return &projectSPSControllerRepo{db: db}
}

func (r *projectSPSControllerRepo) GetByIds(ids []uuid.UUID) ([]*project.ProjectSPSController, error) {
if len(ids) == 0 {
return []*project.ProjectSPSController{}, nil
}
var items []*project.ProjectSPSController
err := r.db.Where("id IN ?", ids).Find(&items).Error
return items, err
}

func (r *projectSPSControllerRepo) Create(entity *project.ProjectSPSController) error {
return r.db.Create(entity).Error
}

func (r *projectSPSControllerRepo) Update(entity *project.ProjectSPSController) error {
return r.db.Save(entity).Error
}

func (r *projectSPSControllerRepo) DeleteByIds(ids []uuid.UUID) error {
if len(ids) == 0 {
return nil
}
return r.db.Where("id IN ?", ids).Delete(&project.ProjectSPSController{}).Error
}

func (r *projectSPSControllerRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[project.ProjectSPSController], error) {
page := params.Page
limit := params.Limit
if page <= 0 {
page = 1
}
if limit <= 0 {
limit = 10
}
offset := (page - 1) * limit

var total int64
if err := r.db.Model(&project.ProjectSPSController{}).Count(&total).Error; err != nil {
return nil, err
}

var items []project.ProjectSPSController
err := r.db.Offset(offset).Limit(limit).Find(&items).Error
if err != nil {
return nil, err
}

return &domain.PaginatedList[project.ProjectSPSController]{
Items:      items,
Total:      total,
Page:       page,
TotalPages: domain.CalculateTotalPages(total, limit),
}, nil
}

// GetByProjectID retrieves all SPS controllers associated with a project
func (r *projectSPSControllerRepo) GetByProjectID(projectID uuid.UUID) ([]*project.ProjectSPSController, error) {
var items []*project.ProjectSPSController
err := r.db.Where("project_id = ?", projectID).Find(&items).Error
return items, err
}

// GetBySPSControllerID retrieves all projects associated with an SPS controller
func (r *projectSPSControllerRepo) GetBySPSControllerID(spsControllerID uuid.UUID) ([]*project.ProjectSPSController, error) {
var items []*project.ProjectSPSController
err := r.db.Where("sps_controller_id = ?", spsControllerID).Find(&items).Error
return items, err
}

// DeleteByProjectAndSPSController deletes a specific association
func (r *projectSPSControllerRepo) DeleteByProjectAndSPSController(projectID, spsControllerID uuid.UUID) error {
return r.db.Where("project_id = ? AND sps_controller_id = ?", projectID, spsControllerID).
Delete(&project.ProjectSPSController{}).Error
}
