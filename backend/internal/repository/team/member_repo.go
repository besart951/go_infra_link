package team

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type memberRepo struct {
	db *gorm.DB
}

func NewTeamMemberRepository(db *gorm.DB) domainTeam.TeamMemberRepository {
	return &memberRepo{db: db}
}

func (r *memberRepo) GetUserRole(ctx context.Context, teamID, userID uuid.UUID) (*domainTeam.MemberRole, error) {
	var member domainTeam.TeamMember
	err := r.db.WithContext(ctx).Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	role := member.Role
	return &role, nil
}

func (r *memberRepo) Upsert(ctx context.Context, member *domainTeam.TeamMember) error {
	now := time.Now().UTC()
	if err := member.Base.InitForCreate(now); err != nil {
		return err
	}
	if member.JoinedAt.IsZero() {
		member.JoinedAt = now
	}

	result := r.db.WithContext(ctx).Model(&domainTeam.TeamMember{}).
		Where("team_id = ? AND user_id = ?", member.TeamID, member.UserID).
		Updates(map[string]any{
			"updated_at": now,
			"role":       member.Role,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	return r.db.WithContext(ctx).Create(member).Error
}

func (r *memberRepo) Delete(ctx context.Context, teamID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Delete(&domainTeam.TeamMember{}).Error
}

func (r *memberRepo) ListByTeam(ctx context.Context, teamID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainTeam.TeamMember], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 20)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&domainTeam.TeamMember{}).Where("team_id = ?", teamID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainTeam.TeamMember
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainTeam.TeamMember]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *memberRepo) ListByUser(ctx context.Context, userID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainTeam.TeamMember], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 20)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&domainTeam.TeamMember{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainTeam.TeamMember
	if err := query.Order("joined_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainTeam.TeamMember]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
