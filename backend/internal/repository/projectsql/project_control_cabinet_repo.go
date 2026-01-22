package projectsql

import (
"github.com/besart951/go_infra_link/backend/internal/domain"
"github.com/besart951/go_infra_link/backend/internal/domain/project"
"github.com/google/uuid"
"gorm.io/gorm"
)

type projectControlCabinetRepo struct {
db *gorm.DB
}

func NewProjectControlCabinetRepository(db *gorm.DB) project.ProjectControlCabinetRepository {
return &projectControlCabinetRepo{db: db}
}

func (r *projectControlCabinetRepo) GetByIds(ids []uuid.UUID) ([]*project.ProjectControlCabinet, error) {
if len(ids) == 0 {
return []*project.ProjectControlCabinet{}, nil
}
var items []*project.ProjectControlCabinet
err := r.db.Where("id IN ?", ids).Find(&items).Error
return items, err
}

func (r *projectControlCabinetRepo) Create(entity *project.ProjectControlCabinet) error {
return r.db.Create(entity).Error
}

func (r *projectControlCabinetRepo) Update(entity *project.ProjectControlCabinet) error {
return r.db.Save(entity).Error
}

func (r *projectControlCabinetRepo) DeleteByIds(ids []uuid.UUID) error {
if len(ids) == 0 {
return nil
}
return r.db.Where("id IN ?", ids).Delete(&project.ProjectControlCabinet{}).Error
}

func (r *projectControlCabinetRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[project.ProjectControlCabinet], error) {
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
if err := r.db.Model(&project.ProjectControlCabinet{}).Count(&total).Error; err != nil {
return nil, err
}

var items []project.ProjectControlCabinet
err := r.db.Offset(offset).Limit(limit).Find(&items).Error
if err != nil {
return nil, err
}

return &domain.PaginatedList[project.ProjectControlCabinet]{
Items:      items,
Total:      total,
Page:       page,
TotalPages: domain.CalculateTotalPages(total, limit),
}, nil
}

// GetByProjectID retrieves all control cabinets associated with a project
func (r *projectControlCabinetRepo) GetByProjectID(projectID uuid.UUID) ([]*project.ProjectControlCabinet, error) {
var items []*project.ProjectControlCabinet
err := r.db.Where("project_id = ?", projectID).Find(&items).Error
return items, err
}

// GetByControlCabinetID retrieves all projects associated with a control cabinet
func (r *projectControlCabinetRepo) GetByControlCabinetID(controlCabinetID uuid.UUID) ([]*project.ProjectControlCabinet, error) {
var items []*project.ProjectControlCabinet
err := r.db.Where("control_cabinet_id = ?", controlCabinetID).Find(&items).Error
return items, err
}

// DeleteByProjectAndControlCabinet deletes a specific association
func (r *projectControlCabinetRepo) DeleteByProjectAndControlCabinet(projectID, controlCabinetID uuid.UUID) error {
return r.db.Where("project_id = ? AND control_cabinet_id = ?", projectID, controlCabinetID).
Delete(&project.ProjectControlCabinet{}).Error
}
