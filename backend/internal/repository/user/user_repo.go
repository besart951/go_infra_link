package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/besart951/go_infra_link/backend/internal/repository/searchspec"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepo struct {
	*gormbase.BaseRepository[*domainUser.User]
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domainUser.UserRepository {
	baseRepo := gormbase.NewBaseRepository(db, userSearchCallback())
	return &userRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	var user domainUser.User
	err := r.db.WithContext(ctx).
		Where("email IS NOT NULL").
		Where("LOWER(email) = ?", domainUser.NormalizeEmail(email)).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) Update(ctx context.Context, entity *domainUser.User) error {
	entity.Base.TouchForUpdate(time.Now().UTC())

	updates := map[string]any{
		"updated_at":            entity.UpdatedAt,
		"first_name":            entity.FirstName,
		"last_name":             entity.LastName,
		"email":                 entity.Email,
		"is_active":             entity.IsActive,
		"role":                  entity.Role,
		"disabled_at":           entity.DisabledAt,
		"deleted_at":            entity.DeletedAt,
		"deleted_by_id":         entity.DeletedByID,
		"restore_until":         entity.RestoreUntil,
		"scheduled_purge_at":    entity.ScheduledPurgeAt,
		"anonymized_at":         entity.AnonymizedAt,
		"deleted_email_hash":    entity.DeletedEmailHash,
		"locked_until":          entity.LockedUntil,
		"failed_login_attempts": entity.FailedLoginAttempts,
		"last_login_at":         entity.LastLoginAt,
		"created_by_id":         entity.CreatedByID,
	}
	if strings.TrimSpace(entity.Password) != "" {
		updates["password"] = entity.Password
	}

	return r.db.WithContext(ctx).Model(&domainUser.User{}).
		Where("id = ?", entity.ID).
		Updates(updates).Error
}

func (r *userRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domainUser.User{}).
			Where("created_by_id IN ?", ids).
			Update("created_by_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Model(&domainUser.UserInvitation{}).
			Where("created_by_id IN ?", ids).
			Update("created_by_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Model(&domainNotification.SystemNotification{}).
			Where("actor_id IN ?", ids).
			Update("actor_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Model(&domainNotification.NotificationRule{}).
			Where("created_by_id IN ?", ids).
			Update("created_by_id", nil).Error; err != nil {
			return err
		}

		deletes := []struct {
			model  any
			column string
		}{
			{&domainAuth.RefreshToken{}, "user_id"},
			{&domainUser.BusinessDetails{}, "user_id"},
			{&domainUser.UserTeam{}, "user_id"},
			{&domainTeam.TeamMember{}, "user_id"},
			{&domainUser.UserInvitation{}, "user_id"},
			{&domainNotification.UserPreference{}, "user_id"},
			{&domainNotification.SystemNotification{}, "recipient_id"},
			{&domainNotification.EmailOutbox{}, "recipient_id"},
		}
		for _, item := range deletes {
			if err := tx.Where(fmt.Sprintf("%s IN ?", item.column), ids).Delete(item.model).Error; err != nil {
				return err
			}
		}
		if err := tx.Exec("DELETE FROM project_users WHERE user_id IN ?", ids).Error; err != nil {
			return err
		}

		return tx.Where("id IN ?", ids).Delete(&domainUser.User{}).Error
	})
}

func (r *userRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainUser.User], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&domainUser.User{})
	if !params.IncludeDeleted {
		query = query.Where("deleted_at IS NULL").Where("anonymized_at IS NULL")
	}
	if strings.TrimSpace(params.Search) != "" {
		query = userSearchCallback()(query, params.Search)
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

func (r *userRepo) ListByRoles(ctx context.Context, roles []domainUser.Role) ([]domainUser.User, error) {
	if len(roles) == 0 {
		return []domainUser.User{}, nil
	}

	var users []domainUser.User
	if err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("anonymized_at IS NULL").
		Where("role IN ?", roles).
		Order("created_at ASC").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepo) ListDueForAnonymization(ctx context.Context, now time.Time, limit int) ([]*domainUser.User, error) {
	if limit <= 0 {
		limit = 100
	}
	items := make([]*domainUser.User, 0, limit)
	err := r.db.WithContext(ctx).
		Model(&domainUser.User{}).
		Where("deleted_at IS NOT NULL").
		Where("anonymized_at IS NULL").
		Where("scheduled_purge_at IS NOT NULL").
		Where("scheduled_purge_at <= ?", now).
		Order("scheduled_purge_at ASC").
		Limit(limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

var _ domainUser.UserEmailRepository = (*userRepo)(nil)
var _ domainUser.UserLifecycleRepository = (*userRepo)(nil)
var _ domainUser.UserRoleRepository = (*userRepo)(nil)

func userSearchCallback() gormbase.SearchCallback[*domainUser.User] {
	return gormbase.TrigramSearchCallback[*domainUser.User](searchspec.Users.SearchColumns("")...)
}
