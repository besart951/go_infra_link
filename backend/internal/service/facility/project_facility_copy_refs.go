package facility

import (
	"context"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (c projectFacilityCopy) remapSoftwareReferences(
	ctx context.Context,
	clonesByOriginalID map[uuid.UUID]*domainFacility.BacnetObject,
	refsByOriginalID map[uuid.UUID]*uuid.UUID,
) error {
	if len(clonesByOriginalID) == 0 || len(refsByOriginalID) == 0 {
		return nil
	}

	newIDsByOriginalID := make(map[uuid.UUID]uuid.UUID, len(clonesByOriginalID))
	for originalID, clone := range clonesByOriginalID {
		if clone == nil {
			continue
		}
		newIDsByOriginalID[originalID] = clone.ID
	}

	for originalID, referenceID := range refsByOriginalID {
		if referenceID == nil {
			continue
		}

		mappedRefID, ok := newIDsByOriginalID[*referenceID]
		if !ok {
			continue
		}

		clone := clonesByOriginalID[originalID]
		if clone == nil {
			continue
		}

		clone.SoftwareReferenceID = &mappedRefID
		if err := c.bacnetObjectRepo.Update(ctx, clone); err != nil {
			return err
		}
	}

	return nil
}
