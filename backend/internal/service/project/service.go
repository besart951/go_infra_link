package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type Service struct {
	repo             domainProject.ProjectRepository
	objectDataRepo   domainFacility.ObjectDataStore
	bacnetObjectRepo domainFacility.BacnetObjectStore
}

func New(repo domainProject.ProjectRepository, objectDataRepo domainFacility.ObjectDataStore, bacnetObjectRepo domainFacility.BacnetObjectStore) *Service {
	return &Service{repo: repo, objectDataRepo: objectDataRepo, bacnetObjectRepo: bacnetObjectRepo}
}

func (s *Service) Create(project *domainProject.Project) error {
	if project.Status == "" {
		project.Status = domainProject.StatusPlanned
	}

	if err := s.repo.Create(project); err != nil {
		return err
	}

	// Copy ObjectData templates
	templates, err := s.objectDataRepo.GetTemplates()
	if err != nil {
		return err
	}

	for _, tmpl := range templates {
		copy := *tmpl
		copy.ID = uuid.Nil
		copy.ProjectID = &project.ID
		copy.BacnetObjects = nil // clear for now, we will rebuild them

		if err := s.objectDataRepo.Create(&copy); err != nil {
			return err
		}

		// Deep copy BacnetObjects
		if len(tmpl.BacnetObjects) == 0 {
			continue
		}

		// Map old ID -> new Instance
		oldToNew := make(map[uuid.UUID]*domainFacility.BacnetObject)
		// Map old ID -> old SoftwareReferenceID (for second pass)
		oldRefs := make(map[uuid.UUID]*uuid.UUID)

		// 1. Create clones
		for _, bo := range tmpl.BacnetObjects {
			newBO := &domainFacility.BacnetObject{
				TextFix:             bo.TextFix,
				Description:         bo.Description,
				GMSVisible:          bo.GMSVisible,
				Optional:            bo.Optional,
				TextIndividual:      bo.TextIndividual,
				SoftwareType:        bo.SoftwareType,
				SoftwareNumber:      bo.SoftwareNumber,
				HardwareType:        bo.HardwareType,
				HardwareQuantity:    bo.HardwareQuantity,
				StateTextID:         bo.StateTextID,
				NotificationClassID: bo.NotificationClassID,
				AlarmDefinitionID:   bo.AlarmDefinitionID,
				// FieldDeviceID is NULL for ObjectData templates
			}
			if err := s.bacnetObjectRepo.Create(newBO); err != nil {
				return err
			}
			oldToNew[bo.ID] = newBO
			oldRefs[bo.ID] = bo.SoftwareReferenceID
		}

		// 2. Fix references and link to new ObjectData
		newBacnetObjects := make([]*domainFacility.BacnetObject, 0, len(tmpl.BacnetObjects))
		for oldID, newBO := range oldToNew {
			// Fix reference
			if refID := oldRefs[oldID]; refID != nil {
				if target, ok := oldToNew[*refID]; ok {
					id := target.ID
					newBO.SoftwareReferenceID = &id
					if err := s.bacnetObjectRepo.Update(newBO); err != nil {
						return err
					}
				}
			}
			newBacnetObjects = append(newBacnetObjects, newBO)
		}

		// 3. Associate with ObjectData
		copy.BacnetObjects = newBacnetObjects
		if err := s.objectDataRepo.Update(&copy); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GetByIds(ids []uuid.UUID) ([]*domainProject.Project, error) {
	return s.repo.GetByIds(ids)
}

func (s *Service) GetByID(id uuid.UUID) (*domainProject.Project, error) {
	projects, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return nil, domain.ErrNotFound
	}
	return projects[0], nil
}

func (s *Service) Update(project *domainProject.Project) error {
	return s.repo.Update(project)
}

func (s *Service) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}

func (s *Service) List(page, limit int, search string) (*domain.PaginatedList[domainProject.Project], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}
