package facility

import (
	"context"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
)

type UnitService struct {
	baseService[domainFacility.Unit]
}

func NewUnitService(repo domainFacility.UnitRepository) *UnitService {
	return &UnitService{baseService: newBase(repo, 20)}
}

func (s *UnitService) Create(ctx context.Context, unit *domainFacility.Unit) error {
	if err := validateUnit(unit); err != nil {
		return err
	}
	return s.repo.Create(ctx, unit)
}

func (s *UnitService) Update(ctx context.Context, unit *domainFacility.Unit) error {
	if err := validateUnit(unit); err != nil {
		return err
	}
	return s.repo.Update(ctx, unit)
}

func validateUnit(unit *domainFacility.Unit) error {
	ve := domain.NewValidationError()
	if strings.TrimSpace(unit.Code) == "" {
		ve = ve.Add("code", "required")
	}
	if strings.TrimSpace(unit.Symbol) == "" {
		ve = ve.Add("symbol", "required")
	}
	if strings.TrimSpace(unit.Name) == "" {
		ve = ve.Add("name", "required")
	}
	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
