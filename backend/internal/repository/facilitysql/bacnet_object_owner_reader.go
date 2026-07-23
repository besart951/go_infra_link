package facilitysql

import (
	"context"
	"sort"

	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bacnetObjectOwnerReader struct {
	db *gorm.DB
}

func NewBacnetObjectOwnerReader(db *gorm.DB) domainObjectData.BacnetObjectOwnerReader {
	return &bacnetObjectOwnerReader{db: db}
}

func (r *bacnetObjectOwnerReader) GetByBacnetObjectIDs(
	ctx context.Context,
	bacnetObjectIDs []uuid.UUID,
) ([]domainObjectData.BacnetObjectOwner, error) {
	if len(bacnetObjectIDs) == 0 {
		return []domainObjectData.BacnetObjectOwner{}, nil
	}

	owners := make([]domainObjectData.BacnetObjectOwner, 0)
	for _, chunk := range uuidFilterChunks(bacnetObjectIDs, uuidFilterChunkSize) {
		var rows []domainObjectData.BacnetObjectOwner
		if err := r.db.WithContext(ctx).
			Table("object_data_bacnet_objects AS odb").
			Select("odb.bacnet_object_id, odb.object_data_id, od.project_id").
			Joins("JOIN object_data AS od ON od.id = odb.object_data_id").
			Where("odb.bacnet_object_id IN ?", chunk).
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		owners = append(owners, rows...)
	}

	sort.Slice(owners, func(i, j int) bool {
		if owners[i].BacnetObjectID != owners[j].BacnetObjectID {
			return owners[i].BacnetObjectID.String() < owners[j].BacnetObjectID.String()
		}
		return owners[i].ObjectDataID.String() < owners[j].ObjectDataID.String()
	})
	return owners, nil
}

var _ domainObjectData.BacnetObjectOwnerReader = (*bacnetObjectOwnerReader)(nil)
