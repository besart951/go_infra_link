package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

func collectProjectFieldDeviceLinkIDsByFieldDeviceIDs(repo domainProject.ProjectFieldDeviceRepository, fieldDeviceIDs []uuid.UUID) ([]uuid.UUID, error) {
	if repo == nil || len(fieldDeviceIDs) == 0 {
		return nil, nil
	}

	idSet := toUUIDSet(fieldDeviceIDs)
	result := make([]uuid.UUID, 0)
	page := 1

	for {
		items, err := repo.GetPaginatedList(domain.PaginationParams{Page: page, Limit: 500})
		if err != nil {
			return nil, err
		}

		for _, item := range items.Items {
			if _, ok := idSet[item.FieldDeviceID]; ok {
				result = append(result, item.ID)
			}
		}

		if page >= items.TotalPages || len(items.Items) == 0 {
			break
		}
		page++
	}

	return result, nil
}

func collectProjectSPSControllerLinkIDsBySPSControllerIDs(repo domainProject.ProjectSPSControllerRepository, spsControllerIDs []uuid.UUID) ([]uuid.UUID, error) {
	if repo == nil || len(spsControllerIDs) == 0 {
		return nil, nil
	}

	idSet := toUUIDSet(spsControllerIDs)
	result := make([]uuid.UUID, 0)
	page := 1

	for {
		items, err := repo.GetPaginatedList(domain.PaginationParams{Page: page, Limit: 500})
		if err != nil {
			return nil, err
		}

		for _, item := range items.Items {
			if _, ok := idSet[item.SPSControllerID]; ok {
				result = append(result, item.ID)
			}
		}

		if page >= items.TotalPages || len(items.Items) == 0 {
			break
		}
		page++
	}

	return result, nil
}

func collectProjectControlCabinetLinkIDsByControlCabinetIDs(repo domainProject.ProjectControlCabinetRepository, controlCabinetIDs []uuid.UUID) ([]uuid.UUID, error) {
	if repo == nil || len(controlCabinetIDs) == 0 {
		return nil, nil
	}

	idSet := toUUIDSet(controlCabinetIDs)
	result := make([]uuid.UUID, 0)
	page := 1

	for {
		items, err := repo.GetPaginatedList(domain.PaginationParams{Page: page, Limit: 500})
		if err != nil {
			return nil, err
		}

		for _, item := range items.Items {
			if _, ok := idSet[item.ControlCabinetID]; ok {
				result = append(result, item.ID)
			}
		}

		if page >= items.TotalPages || len(items.Items) == 0 {
			break
		}
		page++
	}

	return result, nil
}

func toUUIDSet(ids []uuid.UUID) map[uuid.UUID]struct{} {
	result := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		result[id] = struct{}{}
	}
	return result
}
