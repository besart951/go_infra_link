package team

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/google/uuid"
)

type Service struct {
	repo       domainTeam.TeamRepository
	memberRepo domainTeam.TeamMemberRepository
}

func New(repo domainTeam.TeamRepository, memberRepo domainTeam.TeamMemberRepository) *Service {
	return &Service{repo: repo, memberRepo: memberRepo}
}

func (s *Service) Create(ctx context.Context, team *domainTeam.Team) error {
	return s.repo.Create(ctx, team)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domainTeam.Team, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *Service) List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainTeam.Team], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(ctx, domain.PaginationParams{Page: page, Limit: limit, Search: search})
}

func (s *Service) Update(ctx context.Context, team *domainTeam.Team) error {
	return s.repo.Update(ctx, team)
}

func (s *Service) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteByIds(ctx, []uuid.UUID{id})
}

func (s *Service) AddMember(ctx context.Context, teamID, userID uuid.UUID, role domainTeam.MemberRole) error {
	m := &domainTeam.TeamMember{TeamID: teamID, UserID: userID, Role: role}
	return s.memberRepo.Upsert(ctx, m)
}

func (s *Service) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	return s.memberRepo.Delete(ctx, teamID, userID)
}

func (s *Service) ListMembers(ctx context.Context, teamID uuid.UUID, page, limit int) (*domain.PaginatedList[domainTeam.TeamMember], error) {
	page, limit = domain.NormalizePagination(page, limit, 20)
	return s.memberRepo.ListByTeam(ctx, teamID, domain.PaginationParams{Page: page, Limit: limit})
}
