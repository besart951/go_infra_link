package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type ObjectData struct {
	domain.Base
	Description string
	Version     string
	IsActive    bool
	ProjectID   *uuid.UUID

	Project *project.Project

	BacnetObjects []*BacnetObject
	Apparats      []*Apparat
}
