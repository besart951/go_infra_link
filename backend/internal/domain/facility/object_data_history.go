package facility

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONMap is a helper for JSON fields
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return gorm.ErrInvalidField
	}
	return json.Unmarshal(bytes, j)
}

type ObjectDataHistory struct {
	domain.Base
	ObjectDataID *uuid.UUID
	ObjectData   *ObjectData `gorm:"foreignKey:ObjectDataID;constraint:OnDelete:SET NULL"`
	UserID       *uuid.UUID
	User         *domain.User `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
	Action       string
	Changes      JSONMap `gorm:"type:json"`
}
