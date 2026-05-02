package historysql

import (
	"context"
	"encoding/json"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

const (
	scopeProject                 = "project"
	scopeBuilding                = "building"
	scopeControlCabinet          = "control_cabinet"
	scopeSPSController           = "sps_controller"
	scopeSPSControllerSystemType = "sps_controller_system_type"
	scopeFieldDevice             = "field_device"
	scopeSpecification           = "specification"
	scopeBacnetObject            = "bacnet_object"
	scopeObjectData              = "object_data"
)

type resolvedScope struct {
	Type string
	ID   uuid.UUID
}

type scopeSet map[string]resolvedScope

func (s scopeSet) add(scopeType string, id uuid.UUID) {
	if id == uuid.Nil {
		return
	}
	s[scopeType+":"+id.String()] = resolvedScope{Type: scopeType, ID: id}
}

func (s scopeSet) values() []resolvedScope {
	out := make([]resolvedScope, 0, len(s))
	for _, scope := range s {
		out = append(out, scope)
	}
	return out
}

func (s *Store) resolveScopes(ctx context.Context, table string, id uuid.UUID, snapshot domainHistory.JSONB) ([]resolvedScope, error) {
	scopes := scopeSet{}
	var row map[string]any
	if len(snapshot) > 0 {
		_ = json.Unmarshal(snapshot, &row)
	}

	switch table {
	case "projects":
		scopes.add(scopeProject, id)
	case "buildings":
		scopes.add(scopeBuilding, id)
	case "control_cabinets":
		scopes.add(scopeControlCabinet, id)
		scopes.add(scopeBuilding, uuidField(row, "building_id"))
		if err := s.addProjectsForControlCabinet(ctx, scopes, id); err != nil {
			return nil, err
		}
	case "sps_controllers":
		scopes.add(scopeSPSController, id)
		controlCabinetID := uuidField(row, "control_cabinet_id")
		scopes.add(scopeControlCabinet, controlCabinetID)
		if err := s.addAncestorScopesForControlCabinet(ctx, scopes, controlCabinetID); err != nil {
			return nil, err
		}
		if err := s.addProjectsForSPSController(ctx, scopes, id); err != nil {
			return nil, err
		}
	case "sps_controller_system_types":
		scopes.add(scopeSPSControllerSystemType, id)
		spsID := uuidField(row, "sps_controller_id")
		scopes.add(scopeSPSController, spsID)
		if err := s.addAncestorScopesForSPSController(ctx, scopes, spsID); err != nil {
			return nil, err
		}
		if err := s.addProjectsForSPSController(ctx, scopes, spsID); err != nil {
			return nil, err
		}
	case "field_devices":
		scopes.add(scopeFieldDevice, id)
		systemTypeID := uuidField(row, "sps_controller_system_type_id")
		scopes.add(scopeSPSControllerSystemType, systemTypeID)
		if err := s.addAncestorScopesForSystemType(ctx, scopes, systemTypeID); err != nil {
			return nil, err
		}
		if err := s.addProjectsForFieldDevice(ctx, scopes, id); err != nil {
			return nil, err
		}
	case "specifications":
		scopes.add(scopeSpecification, id)
		fieldDeviceID := uuidField(row, "field_device_id")
		scopes.add(scopeFieldDevice, fieldDeviceID)
		if err := s.addScopesForFieldDevice(ctx, scopes, fieldDeviceID); err != nil {
			return nil, err
		}
	case "bacnet_objects":
		scopes.add(scopeBacnetObject, id)
		fieldDeviceID := uuidField(row, "field_device_id")
		if fieldDeviceID != uuid.Nil {
			scopes.add(scopeFieldDevice, fieldDeviceID)
			if err := s.addScopesForFieldDevice(ctx, scopes, fieldDeviceID); err != nil {
				return nil, err
			}
		}
		if err := s.addObjectDataScopesForBacnetObject(ctx, scopes, id); err != nil {
			return nil, err
		}
	case "bacnet_object_alarm_values":
		bacnetObjectID := uuidField(row, "bacnet_object_id")
		scopes.add(scopeBacnetObject, bacnetObjectID)
		if err := s.addScopesForBacnetObject(ctx, scopes, bacnetObjectID); err != nil {
			return nil, err
		}
	case "object_data":
		scopes.add(scopeObjectData, id)
		projectID := uuidField(row, "project_id")
		scopes.add(scopeProject, projectID)
	case "project_control_cabinets":
		projectID := uuidField(row, "project_id")
		controlCabinetID := uuidField(row, "control_cabinet_id")
		scopes.add(scopeProject, projectID)
		scopes.add(scopeControlCabinet, controlCabinetID)
		if err := s.addAncestorScopesForControlCabinet(ctx, scopes, controlCabinetID); err != nil {
			return nil, err
		}
	case "project_sps_controllers":
		projectID := uuidField(row, "project_id")
		spsID := uuidField(row, "sps_controller_id")
		scopes.add(scopeProject, projectID)
		scopes.add(scopeSPSController, spsID)
		if err := s.addAncestorScopesForSPSController(ctx, scopes, spsID); err != nil {
			return nil, err
		}
	case "project_field_devices":
		projectID := uuidField(row, "project_id")
		fieldDeviceID := uuidField(row, "field_device_id")
		scopes.add(scopeProject, projectID)
		scopes.add(scopeFieldDevice, fieldDeviceID)
		if err := s.addScopesForFieldDevice(ctx, scopes, fieldDeviceID); err != nil {
			return nil, err
		}
	default:
		scopes.add(table, id)
	}

	return scopes.values(), nil
}

func uuidField(row map[string]any, key string) uuid.UUID {
	if row == nil {
		return uuid.Nil
	}
	raw, ok := row[key]
	if !ok || raw == nil {
		return uuid.Nil
	}
	switch v := raw.(type) {
	case string:
		id, _ := uuid.Parse(v)
		return id
	default:
		id, _ := uuid.Parse(toString(v))
		return id
	}
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func (s *Store) addAncestorScopesForControlCabinet(ctx context.Context, scopes scopeSet, controlCabinetID uuid.UUID) error {
	if controlCabinetID == uuid.Nil {
		return nil
	}
	var row struct {
		BuildingID uuid.UUID `gorm:"column:building_id"`
	}
	if err := s.db.WithContext(ctx).
		Table("control_cabinets").
		Select("building_id").
		Where("id = ?", controlCabinetID).
		Scan(&row).Error; err != nil {
		return err
	}
	scopes.add(scopeBuilding, row.BuildingID)
	return nil
}

func (s *Store) addAncestorScopesForSPSController(ctx context.Context, scopes scopeSet, spsControllerID uuid.UUID) error {
	if spsControllerID == uuid.Nil {
		return nil
	}
	var row struct {
		ControlCabinetID uuid.UUID `gorm:"column:control_cabinet_id"`
		BuildingID       uuid.UUID `gorm:"column:building_id"`
	}
	if err := s.db.WithContext(ctx).
		Table("sps_controllers AS s").
		Select("s.control_cabinet_id, c.building_id").
		Joins("LEFT JOIN control_cabinets AS c ON c.id = s.control_cabinet_id").
		Where("s.id = ?", spsControllerID).
		Scan(&row).Error; err != nil {
		return err
	}
	scopes.add(scopeControlCabinet, row.ControlCabinetID)
	scopes.add(scopeBuilding, row.BuildingID)
	return nil
}

func (s *Store) addAncestorScopesForSystemType(ctx context.Context, scopes scopeSet, systemTypeID uuid.UUID) error {
	if systemTypeID == uuid.Nil {
		return nil
	}
	var row struct {
		SPSControllerID  uuid.UUID `gorm:"column:sps_controller_id"`
		ControlCabinetID uuid.UUID `gorm:"column:control_cabinet_id"`
		BuildingID       uuid.UUID `gorm:"column:building_id"`
	}
	if err := s.db.WithContext(ctx).
		Table("sps_controller_system_types AS st").
		Select("st.sps_controller_id, s.control_cabinet_id, c.building_id").
		Joins("LEFT JOIN sps_controllers AS s ON s.id = st.sps_controller_id").
		Joins("LEFT JOIN control_cabinets AS c ON c.id = s.control_cabinet_id").
		Where("st.id = ?", systemTypeID).
		Scan(&row).Error; err != nil {
		return err
	}
	scopes.add(scopeSPSController, row.SPSControllerID)
	scopes.add(scopeControlCabinet, row.ControlCabinetID)
	scopes.add(scopeBuilding, row.BuildingID)
	return nil
}

func (s *Store) addScopesForFieldDevice(ctx context.Context, scopes scopeSet, fieldDeviceID uuid.UUID) error {
	if fieldDeviceID == uuid.Nil {
		return nil
	}
	var row struct {
		SystemTypeID     uuid.UUID `gorm:"column:sps_controller_system_type_id"`
		SPSControllerID  uuid.UUID `gorm:"column:sps_controller_id"`
		ControlCabinetID uuid.UUID `gorm:"column:control_cabinet_id"`
		BuildingID       uuid.UUID `gorm:"column:building_id"`
	}
	if err := s.db.WithContext(ctx).
		Table("field_devices AS fd").
		Select("fd.sps_controller_system_type_id, st.sps_controller_id, s.control_cabinet_id, c.building_id").
		Joins("LEFT JOIN sps_controller_system_types AS st ON st.id = fd.sps_controller_system_type_id").
		Joins("LEFT JOIN sps_controllers AS s ON s.id = st.sps_controller_id").
		Joins("LEFT JOIN control_cabinets AS c ON c.id = s.control_cabinet_id").
		Where("fd.id = ?", fieldDeviceID).
		Scan(&row).Error; err != nil {
		return err
	}
	scopes.add(scopeSPSControllerSystemType, row.SystemTypeID)
	scopes.add(scopeSPSController, row.SPSControllerID)
	scopes.add(scopeControlCabinet, row.ControlCabinetID)
	scopes.add(scopeBuilding, row.BuildingID)
	return s.addProjectsForFieldDevice(ctx, scopes, fieldDeviceID)
}

func (s *Store) addScopesForBacnetObject(ctx context.Context, scopes scopeSet, bacnetObjectID uuid.UUID) error {
	if bacnetObjectID == uuid.Nil {
		return nil
	}
	var rows []struct {
		FieldDeviceID *uuid.UUID `gorm:"column:field_device_id"`
		ObjectDataID  *uuid.UUID `gorm:"column:object_data_id"`
		ProjectID     *uuid.UUID `gorm:"column:project_id"`
	}
	if err := s.db.WithContext(ctx).
		Raw(`
			SELECT bo.field_device_id, NULL::uuid AS object_data_id, NULL::uuid AS project_id
			FROM bacnet_objects bo
			WHERE bo.id = ?
			UNION
			SELECT NULL::uuid AS field_device_id, odb.object_data_id, od.project_id
			FROM object_data_bacnet_objects odb
			JOIN object_data od ON od.id = odb.object_data_id
			WHERE odb.bacnet_object_id = ?
		`, bacnetObjectID, bacnetObjectID).
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		if row.FieldDeviceID != nil {
			scopes.add(scopeFieldDevice, *row.FieldDeviceID)
			if err := s.addScopesForFieldDevice(ctx, scopes, *row.FieldDeviceID); err != nil {
				return err
			}
		}
		if row.ObjectDataID != nil {
			scopes.add(scopeObjectData, *row.ObjectDataID)
		}
		if row.ProjectID != nil {
			scopes.add(scopeProject, *row.ProjectID)
		}
	}
	return nil
}

func (s *Store) addObjectDataScopesForBacnetObject(ctx context.Context, scopes scopeSet, bacnetObjectID uuid.UUID) error {
	var rows []struct {
		ObjectDataID uuid.UUID  `gorm:"column:object_data_id"`
		ProjectID    *uuid.UUID `gorm:"column:project_id"`
	}
	if err := s.db.WithContext(ctx).
		Table("object_data_bacnet_objects AS odb").
		Select("odb.object_data_id, od.project_id").
		Joins("JOIN object_data od ON od.id = odb.object_data_id").
		Where("odb.bacnet_object_id = ?", bacnetObjectID).
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		scopes.add(scopeObjectData, row.ObjectDataID)
		if row.ProjectID != nil {
			scopes.add(scopeProject, *row.ProjectID)
		}
	}
	return nil
}

func (s *Store) addProjectsForControlCabinet(ctx context.Context, scopes scopeSet, controlCabinetID uuid.UUID) error {
	if controlCabinetID == uuid.Nil {
		return nil
	}
	var ids []uuid.UUID
	if err := s.db.WithContext(ctx).
		Raw(`
			SELECT project_id FROM project_control_cabinets WHERE control_cabinet_id = ?
			UNION
			SELECT psc.project_id
			FROM project_sps_controllers psc
			JOIN sps_controllers s ON s.id = psc.sps_controller_id
			WHERE s.control_cabinet_id = ?
			UNION
			SELECT pfd.project_id
			FROM project_field_devices pfd
			JOIN field_devices fd ON fd.id = pfd.field_device_id
			JOIN sps_controller_system_types st ON st.id = fd.sps_controller_system_type_id
			JOIN sps_controllers s ON s.id = st.sps_controller_id
			WHERE s.control_cabinet_id = ?
		`, controlCabinetID, controlCabinetID, controlCabinetID).
		Scan(&ids).Error; err != nil {
		return err
	}
	for _, id := range ids {
		scopes.add(scopeProject, id)
	}
	return nil
}

func (s *Store) addProjectsForSPSController(ctx context.Context, scopes scopeSet, spsControllerID uuid.UUID) error {
	if spsControllerID == uuid.Nil {
		return nil
	}
	var ids []uuid.UUID
	if err := s.db.WithContext(ctx).
		Raw(`
			SELECT project_id FROM project_sps_controllers WHERE sps_controller_id = ?
			UNION
			SELECT pcc.project_id
			FROM project_control_cabinets pcc
			JOIN sps_controllers s ON s.control_cabinet_id = pcc.control_cabinet_id
			WHERE s.id = ?
			UNION
			SELECT pfd.project_id
			FROM project_field_devices pfd
			JOIN field_devices fd ON fd.id = pfd.field_device_id
			JOIN sps_controller_system_types st ON st.id = fd.sps_controller_system_type_id
			WHERE st.sps_controller_id = ?
		`, spsControllerID, spsControllerID, spsControllerID).
		Scan(&ids).Error; err != nil {
		return err
	}
	for _, id := range ids {
		scopes.add(scopeProject, id)
	}
	return nil
}

func (s *Store) addProjectsForFieldDevice(ctx context.Context, scopes scopeSet, fieldDeviceID uuid.UUID) error {
	if fieldDeviceID == uuid.Nil {
		return nil
	}
	var ids []uuid.UUID
	if err := s.db.WithContext(ctx).
		Raw(`
			SELECT project_id FROM project_field_devices WHERE field_device_id = ?
			UNION
			SELECT psc.project_id
			FROM project_sps_controllers psc
			JOIN sps_controller_system_types st ON st.sps_controller_id = psc.sps_controller_id
			JOIN field_devices fd ON fd.sps_controller_system_type_id = st.id
			WHERE fd.id = ?
			UNION
			SELECT pcc.project_id
			FROM project_control_cabinets pcc
			JOIN sps_controllers s ON s.control_cabinet_id = pcc.control_cabinet_id
			JOIN sps_controller_system_types st ON st.sps_controller_id = s.id
			JOIN field_devices fd ON fd.sps_controller_system_type_id = st.id
			WHERE fd.id = ?
		`, fieldDeviceID, fieldDeviceID, fieldDeviceID).
		Scan(&ids).Error; err != nil {
		return err
	}
	for _, id := range ids {
		scopes.add(scopeProject, id)
	}
	return nil
}
