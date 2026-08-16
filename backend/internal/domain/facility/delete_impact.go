package facility

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type DeleteImpactResource string

const (
	DeleteImpactResourceApparat    DeleteImpactResource = "apparat"
	DeleteImpactResourceSystemPart DeleteImpactResource = "system_part"
)

var ErrReferenceInUse = errors.New("facility reference in use")

type DeleteImpactBlocker struct {
	Resource string
	Count    int64
}

type DeleteImpact struct {
	Resource DeleteImpactResource
	ID       uuid.UUID
	Blockers []DeleteImpactBlocker
}

type DeleteImpactRepository interface {
	DeleteImpacts(ctx context.Context, resource DeleteImpactResource, ids []uuid.UUID) ([]DeleteImpact, error)
}

func ParseDeleteImpactResource(value string) (DeleteImpactResource, bool) {
	resource := DeleteImpactResource(value)
	switch resource {
	case DeleteImpactResourceApparat, DeleteImpactResourceSystemPart:
		return resource, true
	default:
		return "", false
	}
}
