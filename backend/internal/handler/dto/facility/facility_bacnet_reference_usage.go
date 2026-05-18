package facility

import "github.com/google/uuid"

type BacnetReferenceUsageResponse struct {
	Resource          string    `json:"resource"`
	ID                uuid.UUID `json:"id"`
	BacnetObjectCount int64     `json:"bacnet_object_count"`
}

type BacnetReferenceUsageListResponse struct {
	Items []BacnetReferenceUsageResponse `json:"items"`
}
