package user

import (
	"fmt"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domainUser.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) GetByIds(ids []uuid.UUID) ([]*domainUser.User, error) {
	if len(ids) == 0 {
		return []*domainUser.User{}, nil
	}
	var items []*domainUser.User
	err := r.db.Where("deleted_at IS NULL").Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *userRepo) GetByEmail(email string) (*domainUser.User, error) {
	var user domainUser.User
	err := r.db.Where("deleted_at IS NULL").Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) Create(entity *domainUser.User) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}
	return r.db.Create(entity).Error
}

func (r *userRepo) Update(entity *domainUser.User) error {
	entity.Base.TouchForUpdate(time.Now().UTC())

	updates := map[string]any{
		"updated_at":            entity.UpdatedAt,
		"first_name":            entity.FirstName,
		"last_name":             entity.LastName,
		"email":                 entity.Email,
		"is_active":             entity.IsActive,
		"role":                  entity.Role,
		"disabled_at":           entity.DisabledAt,
		"locked_until":          entity.LockedUntil,
		"failed_login_attempts": entity.FailedLoginAttempts,
		"last_login_at":         entity.LastLoginAt,
		"created_by_id":         entity.CreatedByID,
	}
	if strings.TrimSpace(entity.Password) != "" {
		updates["password"] = entity.Password
	}

	return r.db.Model(&domainUser.User{}).
		Where("deleted_at IS NULL AND id = ?", entity.ID).
		Updates(updates).Error
}

func (r *userRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.Model(&domainUser.User{}).
		Where("id IN ?", ids).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}

func (r *userRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainUser.User], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainUser.User{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?", pattern, pattern, pattern)
	}

	allowedColumns := map[string]string{
		"last_login_at": "last_login_at",
		"created_at":    "created_at",
		"first_name":    "first_name",
		"last_name":     "last_name",
		"email":         "email",
		"role":          "role",
	}
	orderBy := "last_login_at"
	if col, ok := allowedColumns[params.OrderBy]; ok {
		orderBy = col
	}
	order := "DESC"
	if strings.EqualFold(params.Order, "asc") {
		order = "ASC"
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainUser.User
	if err := query.Order(fmt.Sprintf("%s %s", orderBy, order)).Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainUser.User]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

var _ domainUser.UserEmailRepository = (*userRepo)(nil)
