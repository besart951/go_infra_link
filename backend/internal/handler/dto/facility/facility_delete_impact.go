package facility

import "github.com/google/uuid"

type DeleteImpactBlockerResponse struct {
	Resource string `json:"resource"`
	Count    int64  `json:"count"`
}

type DeleteImpactResponse struct {
	Resource string                        `json:"resource"`
	ID       uuid.UUID                     `json:"id"`
	Blockers []DeleteImpactBlockerResponse `json:"blockers"`
}

type DeleteImpactListResponse struct {
	Items []DeleteImpactResponse `json:"items"`
}
