package projectsql

import (
"github.com/besart951/go_infra_link/backend/internal/domain"
"github.com/besart951/go_infra_link/backend/internal/domain/project"
"github.com/google/uuid"
"gorm.io/gorm"
)

type projectFieldDeviceRepo struct {
db *gorm.DB
}

func NewProjectFieldDeviceRepository(db *gorm.DB) project.ProjectFieldDeviceRepository {
return &projectFieldDeviceRepo{db: db}
}

func (r *projectFieldDeviceRepo) GetByIds(ids []uuid.UUID) ([]*project.ProjectFieldDevice, error) {
if len(ids) == 0 {
return []*project.ProjectFieldDevice{}, nil
}
var items []*project.ProjectFieldDevice
err := r.db.Where("id IN ?", ids).Find(&items).Error
return items, err
}

func (r *projectFieldDeviceRepo) Create(entity *project.ProjectFieldDevice) error {
return r.db.Create(entity).Error
}

func (r *projectFieldDeviceRepo) Update(entity *project.ProjectFieldDevice) error {
return r.db.Save(entity).Error
}

func (r *projectFieldDeviceRepo) DeleteByIds(ids []uuid.UUID) error {
if len(ids) == 0 {
return nil
}
return r.db.Where("id IN ?", ids).Delete(&project.ProjectFieldDevice{}).Error
}

func (r *projectFieldDeviceRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[project.ProjectFieldDevice], error) {
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
if err := r.db.Model(&project.ProjectFieldDevice{}).Count(&total).Error; err != nil {
return nil, err
}

var items []project.ProjectFieldDevice
err := r.db.Offset(offset).Limit(limit).Find(&items).Error
if err != nil {
return nil, err
}

return &domain.PaginatedList[project.ProjectFieldDevice]{
Items:      items,
Total:      total,
Page:       page,
TotalPages: domain.CalculateTotalPages(total, limit),
}, nil
}

// GetByProjectID retrieves all field devices associated with a project
func (r *projectFieldDeviceRepo) GetByProjectID(projectID uuid.UUID) ([]*project.ProjectFieldDevice, error) {
var items []*project.ProjectFieldDevice
err := r.db.Where("project_id = ?", projectID).Find(&items).Error
return items, err
}

// GetByFieldDeviceID retrieves all projects associated with a field device
func (r *projectFieldDeviceRepo) GetByFieldDeviceID(fieldDeviceID uuid.UUID) ([]*project.ProjectFieldDevice, error) {
var items []*project.ProjectFieldDevice
err := r.db.Where("field_device_id = ?", fieldDeviceID).Find(&items).Error
return items, err
}

// DeleteByProjectAndFieldDevice deletes a specific association
func (r *projectFieldDeviceRepo) DeleteByProjectAndFieldDevice(projectID, fieldDeviceID uuid.UUID) error {
return r.db.Where("project_id = ? AND field_device_id = ?", projectID, fieldDeviceID).
Delete(&project.ProjectFieldDevice{}).Error
}
