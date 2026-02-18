package team

import (
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

func (r *memberRepo) GetUserRole(teamID, userID uuid.UUID) (*domainTeam.MemberRole, error) {
	var member domainTeam.TeamMember
	err := r.db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	role := member.Role
	return &role, nil
}

func (r *memberRepo) Upsert(member *domainTeam.TeamMember) error {
	now := time.Now().UTC()
	if err := member.Base.InitForCreate(now); err != nil {
		return err
	}
	if member.JoinedAt.IsZero() {
		member.JoinedAt = now
	}

	result := r.db.Model(&domainTeam.TeamMember{}).
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

	return r.db.Create(member).Error
}

func (r *memberRepo) Delete(teamID, userID uuid.UUID) error {
	return r.db.
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Delete(&domainTeam.TeamMember{}).Error
}

func (r *memberRepo) ListByTeam(teamID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainTeam.TeamMember], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 20)
	offset := (page - 1) * limit

	query := r.db.Model(&domainTeam.TeamMember{}).Where("team_id = ?", teamID)

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
