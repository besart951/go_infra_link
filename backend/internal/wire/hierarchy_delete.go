package wire

import (
	"context"
	"fmt"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const hierarchyDeletePageSize = 100

// hierarchyDeleteCleaner explicitly removes one owned hierarchy branch through
// decorated repositories. Every page is bounded, so history capture never
// materializes an unbounded facility graph and database cascades do not bypass
// descendant audit records.
type hierarchyDeleteCleaner struct {
	db    *gorm.DB
	repos *Repositories
}

func (c *hierarchyDeleteCleaner) deleteControlCabinet(
	ctx context.Context,
	controlCabinetID uuid.UUID,
) error {
	var after *uuid.UUID
	for {
		ids, err := c.listSPSControllerIDs(ctx, controlCabinetID, after)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			break
		}
		if err := c.deleteSPSControllers(ctx, ids); err != nil {
			return err
		}
		last := ids[len(ids)-1]
		after = &last
	}
	if err := c.repos.ProjectControlCabinets.DeleteByControlCabinetIDs(
		ctx,
		[]uuid.UUID{controlCabinetID},
	); err != nil {
		return fmt.Errorf("delete ControlCabinet project links: %w", err)
	}
	if err := c.repos.FacilityControlCabinet.DeleteByIds(
		ctx,
		[]uuid.UUID{controlCabinetID},
	); err != nil {
		return fmt.Errorf("delete ControlCabinet root: %w", err)
	}
	return nil
}

func (c *hierarchyDeleteCleaner) deleteSPSControllers(
	ctx context.Context,
	spsControllerIDs []uuid.UUID,
) error {
	for _, chunk := range uuidPages(spsControllerIDs, hierarchyDeletePageSize) {
		var after *uuid.UUID
		for {
			systemTypeIDs, err := c.listSPSControllerSystemTypeIDs(ctx, chunk, after)
			if err != nil {
				return err
			}
			if len(systemTypeIDs) == 0 {
				break
			}
			if err := c.deleteSPSControllerSystemTypes(ctx, systemTypeIDs); err != nil {
				return err
			}
			last := systemTypeIDs[len(systemTypeIDs)-1]
			after = &last
		}
		if err := c.repos.ProjectSPSControllers.DeleteBySPSControllerIDs(ctx, chunk); err != nil {
			return fmt.Errorf("delete SPSController project links: %w", err)
		}
		if err := c.repos.FacilitySPSControllers.DeleteByIds(ctx, chunk); err != nil {
			return fmt.Errorf("delete SPSController roots: %w", err)
		}
	}
	return nil
}

func (c *hierarchyDeleteCleaner) deleteSPSControllerSystemTypes(
	ctx context.Context,
	systemTypeIDs []uuid.UUID,
) error {
	for _, chunk := range uuidPages(systemTypeIDs, hierarchyDeletePageSize) {
		var after *uuid.UUID
		for {
			fieldDeviceIDs, err := c.listFieldDeviceIDs(ctx, chunk, after)
			if err != nil {
				return err
			}
			if len(fieldDeviceIDs) == 0 {
				break
			}
			if err := c.deleteFieldDevices(ctx, fieldDeviceIDs); err != nil {
				return err
			}
			last := fieldDeviceIDs[len(fieldDeviceIDs)-1]
			after = &last
		}
		if err := c.repos.FacilitySPSControllerSystemTypes.DeleteByIds(ctx, chunk); err != nil {
			return fmt.Errorf("delete SPSControllerSystemType roots: %w", err)
		}
	}
	return nil
}

type hierarchyDeleteBacnetRow struct {
	ID          uuid.UUID `gorm:"column:id"`
	HasTemplate bool      `gorm:"column:has_template"`
}

func (c *hierarchyDeleteCleaner) deleteFieldDevices(
	ctx context.Context,
	fieldDeviceIDs []uuid.UUID,
) error {
	for _, chunk := range uuidPages(fieldDeviceIDs, hierarchyDeletePageSize) {
		specifications, err := c.repos.FacilitySpecifications.GetByFieldDeviceIDs(ctx, chunk)
		if err != nil {
			return fmt.Errorf("load descendant Specifications: %w", err)
		}
		specificationIDs := make([]uuid.UUID, 0, len(specifications))
		for _, specification := range specifications {
			if specification != nil && specification.ID != uuid.Nil {
				specificationIDs = append(specificationIDs, specification.ID)
			}
		}
		if err := c.repos.FacilitySpecifications.DeleteByIds(ctx, specificationIDs); err != nil {
			return fmt.Errorf("delete descendant Specifications: %w", err)
		}

		if err := c.cleanupFieldDeviceBacnetObjects(ctx, chunk); err != nil {
			return err
		}
		if err := c.repos.ProjectFieldDevices.DeleteByFieldDeviceIDs(ctx, chunk); err != nil {
			return fmt.Errorf("delete FieldDevice project links: %w", err)
		}
		if err := c.repos.FacilityFieldDevices.DeleteByIds(ctx, chunk); err != nil {
			return fmt.Errorf("delete FieldDevice descendants: %w", err)
		}
	}
	return nil
}

func (c *hierarchyDeleteCleaner) cleanupFieldDeviceBacnetObjects(
	ctx context.Context,
	fieldDeviceIDs []uuid.UUID,
) error {
	var rows []hierarchyDeleteBacnetRow
	if err := c.db.WithContext(ctx).Raw(`
		SELECT
			object.id,
			EXISTS (
				SELECT 1
				FROM object_data_bacnet_objects AS link
				WHERE link.bacnet_object_id = object.id
			) AS has_template
		FROM bacnet_objects AS object
		WHERE object.field_device_id IN ?
		ORDER BY object.id
	`, fieldDeviceIDs).Scan(&rows).Error; err != nil {
		return fmt.Errorf("load descendant BACnet ownership: %w", err)
	}

	exclusiveIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		if !row.HasTemplate {
			exclusiveIDs = append(exclusiveIDs, row.ID)
			continue
		}
		object, err := domain.GetByID(ctx, c.repos.FacilityBacnetObjects, row.ID)
		if err != nil {
			return fmt.Errorf("load shared BACnet object %s: %w", row.ID, err)
		}
		object.FieldDeviceID = nil
		if err := c.repos.FacilityBacnetObjects.Update(ctx, object); err != nil {
			return fmt.Errorf("detach shared BACnet object %s: %w", row.ID, err)
		}
	}

	for _, chunk := range uuidPages(exclusiveIDs, hierarchyDeletePageSize) {
		if err := c.clearBacnetSoftwareReferences(ctx, chunk); err != nil {
			return err
		}
		if err := c.deleteBacnetAlarmValues(ctx, chunk); err != nil {
			return err
		}
		if err := c.repos.FacilityBacnetObjects.DeleteByIds(ctx, chunk); err != nil {
			return fmt.Errorf("delete descendant BACnet objects: %w", err)
		}
	}
	return nil
}

func (c *hierarchyDeleteCleaner) clearBacnetSoftwareReferences(
	ctx context.Context,
	targetIDs []uuid.UUID,
) error {
	var after *uuid.UUID
	for {
		query := c.db.WithContext(ctx).
			Model(&domainFacility.BacnetObject{}).
			Where("software_reference_id IN ?", targetIDs)
		if after != nil {
			query = query.Where("id > ?", *after)
		}
		var ids []uuid.UUID
		if err := query.Order("id ASC").
			Limit(hierarchyDeletePageSize).
			Pluck("id", &ids).Error; err != nil {
			return fmt.Errorf("load BACnet software-reference dependants: %w", err)
		}
		if len(ids) == 0 {
			return nil
		}
		objects, err := c.repos.FacilityBacnetObjects.GetByIds(ctx, ids)
		if err != nil {
			return fmt.Errorf("load BACnet software-reference rows: %w", err)
		}
		for _, object := range objects {
			if object == nil {
				continue
			}
			object.SoftwareReferenceID = nil
			if err := c.repos.FacilityBacnetObjects.Update(ctx, object); err != nil {
				return fmt.Errorf("clear BACnet software reference %s: %w", object.ID, err)
			}
		}
		last := ids[len(ids)-1]
		after = &last
	}
}

func (c *hierarchyDeleteCleaner) deleteBacnetAlarmValues(
	ctx context.Context,
	bacnetObjectIDs []uuid.UUID,
) error {
	var after *uuid.UUID
	for {
		query := c.db.WithContext(ctx).
			Model(&domainFacility.BacnetObjectAlarmValue{}).
			Where("bacnet_object_id IN ?", bacnetObjectIDs)
		if after != nil {
			query = query.Where("id > ?", *after)
		}
		var ids []uuid.UUID
		if err := query.Order("id ASC").
			Limit(hierarchyDeletePageSize).
			Pluck("id", &ids).Error; err != nil {
			return fmt.Errorf("load descendant BACnet alarm values: %w", err)
		}
		if len(ids) == 0 {
			return nil
		}
		if err := c.repos.FacilityBacnetObjectAlarmValues.DeleteByIds(ctx, ids); err != nil {
			return fmt.Errorf("delete descendant BACnet alarm values: %w", err)
		}
		last := ids[len(ids)-1]
		after = &last
	}
}

func (c *hierarchyDeleteCleaner) listSPSControllerIDs(
	ctx context.Context,
	controlCabinetID uuid.UUID,
	after *uuid.UUID,
) ([]uuid.UUID, error) {
	query := c.db.WithContext(ctx).
		Model(&domainFacility.SPSController{}).
		Where("control_cabinet_id = ?", controlCabinetID)
	if after != nil {
		query = query.Where("id > ?", *after)
	}
	var ids []uuid.UUID
	err := query.Order("id ASC").Limit(hierarchyDeletePageSize).Pluck("id", &ids).Error
	return ids, err
}

func (c *hierarchyDeleteCleaner) listSPSControllerSystemTypeIDs(
	ctx context.Context,
	spsControllerIDs []uuid.UUID,
	after *uuid.UUID,
) ([]uuid.UUID, error) {
	query := c.db.WithContext(ctx).
		Model(&domainFacility.SPSControllerSystemType{}).
		Where("sps_controller_id IN ?", spsControllerIDs)
	if after != nil {
		query = query.Where("id > ?", *after)
	}
	var ids []uuid.UUID
	err := query.Order("id ASC").Limit(hierarchyDeletePageSize).Pluck("id", &ids).Error
	return ids, err
}

func (c *hierarchyDeleteCleaner) listFieldDeviceIDs(
	ctx context.Context,
	systemTypeIDs []uuid.UUID,
	after *uuid.UUID,
) ([]uuid.UUID, error) {
	query := c.db.WithContext(ctx).
		Model(&domainFacility.FieldDevice{}).
		Where("sps_controller_system_type_id IN ?", systemTypeIDs)
	if after != nil {
		query = query.Where("id > ?", *after)
	}
	var ids []uuid.UUID
	err := query.Order("id ASC").Limit(hierarchyDeletePageSize).Pluck("id", &ids).Error
	return ids, err
}

func uuidPages(ids []uuid.UUID, size int) [][]uuid.UUID {
	if size <= 0 {
		size = hierarchyDeletePageSize
	}
	pages := make([][]uuid.UUID, 0, (len(ids)+size-1)/size)
	for start := 0; start < len(ids); start += size {
		end := start + size
		if end > len(ids) {
			end = len(ids)
		}
		pages = append(pages, ids[start:end])
	}
	return pages
}
