package team

import (
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

func (s *Service) Create(team *domainTeam.Team) error {
	return s.repo.Create(team)
}

func (s *Service) GetByID(id uuid.UUID) (*domainTeam.Team, error) {
	teams, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, domain.ErrNotFound
	}
	return teams[0], nil
}

func (s *Service) List(page, limit int, search string) (*domain.PaginatedList[domainTeam.Team], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{Page: page, Limit: limit, Search: search})
}

func (s *Service) Update(team *domainTeam.Team) error {
	return s.repo.Update(team)
}

func (s *Service) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}

func (s *Service) AddMember(teamID, userID uuid.UUID, role domainTeam.MemberRole) error {
	m := &domainTeam.TeamMember{TeamID: teamID, UserID: userID, Role: role}
	return s.memberRepo.Upsert(m)
}

func (s *Service) RemoveMember(teamID, userID uuid.UUID) error {
	return s.memberRepo.Delete(teamID, userID)
}

func (s *Service) ListMembers(teamID uuid.UUID, page, limit int) (*domain.PaginatedList[domainTeam.TeamMember], error) {
	page, limit = domain.NormalizePagination(page, limit, 20)
	return s.memberRepo.ListByTeam(teamID, domain.PaginationParams{Page: page, Limit: limit})
}
