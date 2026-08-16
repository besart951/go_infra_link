package facility

import "github.com/google/uuid"

// ChangedFields keeps the externally visible mutation vocabulary next to the
// request contract. Conflict responses, project history, and realtime callers
// must use these methods instead of maintaining their own field lists.

func (r UpdateBuildingRequest) ChangedFields() []string {
	fields := make([]string, 0, 2)
	if r.IWSCode != "" {
		fields = append(fields, "iws_code")
	}
	if r.BuildingGroup != 0 {
		fields = append(fields, "building_group")
	}
	return fields
}

func (r UpdateControlCabinetRequest) ChangedFields() []string {
	fields := make([]string, 0, 2)
	if r.BuildingID != uuid.Nil {
		fields = append(fields, "building_id")
	}
	if r.ControlCabinetNr != nil {
		fields = append(fields, "control_cabinet_nr")
	}
	return fields
}

func (r UpdateSPSControllerRequest) ChangedFields() []string {
	fields := make([]string, 0, 10)
	if r.ControlCabinetID != uuid.Nil {
		fields = append(fields, "control_cabinet_id")
	}
	if r.GADevice != nil {
		fields = append(fields, "ga_device")
	}
	if r.DeviceName != "" {
		fields = append(fields, "device_name")
	}
	if r.DeviceDescription != nil {
		fields = append(fields, "device_description")
	}
	if r.DeviceLocation != nil {
		fields = append(fields, "device_location")
	}
	if r.IPAddress != nil {
		fields = append(fields, "ip_address")
	}
	if r.Subnet != nil {
		fields = append(fields, "subnet")
	}
	if r.Gateway != nil {
		fields = append(fields, "gateway")
	}
	if r.Vlan != nil {
		fields = append(fields, "vlan")
	}
	if r.SystemTypes != nil {
		fields = append(fields, "system_types")
	}
	return fields
}

func (r UpdateSPSControllerSystemTypeRequest) ChangedFields() []string {
	fields := make([]string, 0, 2)
	if r.Number != nil {
		fields = append(fields, "number")
	}
	if r.DocumentName != nil {
		fields = append(fields, "document_name")
	}
	return fields
}

func (r UpdateFieldDeviceRequest) ChangedFields() []string {
	fields := make([]string, 0, 9)
	if r.BMK != nil {
		fields = append(fields, "bmk")
	}
	if r.Description != nil {
		fields = append(fields, "description")
	}
	if r.TextIndividuell != nil {
		fields = append(fields, "text_fix")
	}
	if r.ApparatNr != nil {
		fields = append(fields, "apparat_nr")
	}
	if r.SPSControllerSystemTypeID != uuid.Nil {
		fields = append(fields, "sps_controller_system_type_id")
	}
	if r.SystemPartID != uuid.Nil {
		fields = append(fields, "system_part_id")
	}
	if r.ApparatID != uuid.Nil {
		fields = append(fields, "apparat_id")
	}
	if r.ObjectDataID != nil {
		fields = append(fields, "object_data_id")
	}
	if r.BacnetObjects != nil {
		fields = append(fields, "bacnet_objects")
	}
	return fields
}

func (r BulkUpdateFieldDeviceItem) ChangedFields() []string {
	fields := make([]string, 0, 8)
	if r.BMK.Set {
		fields = append(fields, "bmk")
	}
	if r.Description.Set {
		fields = append(fields, "description")
	}
	if r.TextIndividuell.Set {
		fields = append(fields, "text_fix")
	}
	if r.ApparatNr != nil {
		fields = append(fields, "apparat_nr")
	}
	if r.ApparatID != nil {
		fields = append(fields, "apparat_id")
	}
	if r.SystemPartID != nil {
		fields = append(fields, "system_part_id")
	}
	if r.Specification != nil {
		fields = append(fields, "specification")
	}
	if r.BacnetObjects != nil {
		fields = append(fields, "bacnet_objects")
	}
	return fields
}

func (r BulkUpdateFieldDeviceRequest) ChangedFields() []string {
	fields := make([]string, 0, 8)
	seen := make(map[string]struct{})
	for _, update := range r.Updates {
		for _, field := range update.ChangedFields() {
			if _, exists := seen[field]; exists {
				continue
			}
			seen[field] = struct{}{}
			fields = append(fields, field)
		}
	}
	return fields
}

func FieldDeviceSpecificationChangedFields() []string {
	return []string{"specification"}
}

func FieldDeviceBacnetObjectsChangedFields() []string {
	return []string{"bacnet_objects"}
}

func (r UpdateBacnetObjectRequest) ChangedFields() []string {
	fields := r.BacnetObjectPatchInput.ChangedFields()
	if r.FieldDeviceID != nil {
		fields = append(fields, "field_device_id")
	}
	if r.ObjectDataID != nil {
		fields = append(fields, "object_data_id")
	}
	return fields
}

func (r UpdateBacnetObjectRequest) FieldDeviceConflictFields(id string) []string {
	return []string{"bacnet_objects." + id}
}

func (r BacnetObjectPatchInput) ChangedFields() []string {
	fields := make([]string, 0, 14)
	if r.TextFix != nil {
		fields = append(fields, "text_fix")
	}
	if r.Description != nil {
		fields = append(fields, "description")
	}
	if r.GMSVisible != nil {
		fields = append(fields, "gms_visible")
	}
	if r.Optional != nil {
		fields = append(fields, "optional")
	}
	if r.TextIndividual != nil {
		fields = append(fields, "text_individual")
	}
	if r.SoftwareType != nil {
		fields = append(fields, "software_type")
	}
	if r.SoftwareNumber != nil {
		fields = append(fields, "software_number")
	}
	if r.HardwareType != nil {
		fields = append(fields, "hardware_type")
	}
	if r.HardwareQuantity != nil {
		fields = append(fields, "hardware_quantity")
	}
	if r.SoftwareReferenceID != nil {
		fields = append(fields, "software_reference_id")
	}
	if r.StateTextID != nil {
		fields = append(fields, "state_text_id")
	}
	if r.NotificationClassID != nil {
		fields = append(fields, "notification_class_id")
	}
	if r.AlarmDefinitionID != nil {
		fields = append(fields, "alarm_definition_id")
	}
	if r.AlarmTypeID != nil {
		fields = append(fields, "alarm_type_id")
	}
	return fields
}

func (r UpdateObjectDataRequest) ChangedFields() []string {
	fields := make([]string, 0, 6)
	if r.Description != nil {
		fields = append(fields, "description")
	}
	if r.Version != nil {
		fields = append(fields, "version")
	}
	if r.IsActive != nil {
		fields = append(fields, "is_active")
	}
	if r.ProjectID != nil {
		fields = append(fields, "project_id")
	}
	if r.ApparatIDs != nil {
		fields = append(fields, "apparat_ids")
	}
	if r.BacnetObjects != nil {
		fields = append(fields, "bacnet_objects")
	}
	return fields
}

func (r UpdateSystemTypeRequest) ChangedFields() []string {
	fields := make([]string, 0, 3)
	if r.NumberMin != 0 {
		fields = append(fields, "number_min")
	}
	if r.NumberMax != 0 {
		fields = append(fields, "number_max")
	}
	if r.Name != "" {
		fields = append(fields, "name")
	}
	return fields
}

func (r UpdateSystemPartRequest) ChangedFields() []string {
	fields := make([]string, 0, 3)
	if r.ShortName != "" {
		fields = append(fields, "short_name")
	}
	if r.Name != "" {
		fields = append(fields, "name")
	}
	if r.Description != nil {
		fields = append(fields, "description")
	}
	return fields
}

func (r UpdateApparatRequest) ChangedFields() []string {
	fields := make([]string, 0, 4)
	if r.ShortName != "" {
		fields = append(fields, "short_name")
	}
	if r.Name != "" {
		fields = append(fields, "name")
	}
	if r.Description != nil {
		fields = append(fields, "description")
	}
	if r.SystemPartIDs != nil {
		fields = append(fields, "system_part_ids")
	}
	return fields
}

func (r UpdateStateTextRequest) ChangedFields() []string {
	fields := make([]string, 0, 17)
	if r.RefNumber != nil {
		fields = append(fields, "ref_number")
	}
	if r.StateText1 != nil {
		fields = append(fields, "state_text1")
	}
	if r.StateText2 != nil {
		fields = append(fields, "state_text2")
	}
	if r.StateText3 != nil {
		fields = append(fields, "state_text3")
	}
	if r.StateText4 != nil {
		fields = append(fields, "state_text4")
	}
	if r.StateText5 != nil {
		fields = append(fields, "state_text5")
	}
	if r.StateText6 != nil {
		fields = append(fields, "state_text6")
	}
	if r.StateText7 != nil {
		fields = append(fields, "state_text7")
	}
	if r.StateText8 != nil {
		fields = append(fields, "state_text8")
	}
	if r.StateText9 != nil {
		fields = append(fields, "state_text9")
	}
	if r.StateText10 != nil {
		fields = append(fields, "state_text10")
	}
	if r.StateText11 != nil {
		fields = append(fields, "state_text11")
	}
	if r.StateText12 != nil {
		fields = append(fields, "state_text12")
	}
	if r.StateText13 != nil {
		fields = append(fields, "state_text13")
	}
	if r.StateText14 != nil {
		fields = append(fields, "state_text14")
	}
	if r.StateText15 != nil {
		fields = append(fields, "state_text15")
	}
	if r.StateText16 != nil {
		fields = append(fields, "state_text16")
	}
	return fields
}

func (r UpdateNotificationClassRequest) ChangedFields() []string {
	fields := make([]string, 0, 11)
	if r.EventCategory != nil {
		fields = append(fields, "event_category")
	}
	if r.Nc != nil {
		fields = append(fields, "nc")
	}
	if r.ObjectDescription != nil {
		fields = append(fields, "object_description")
	}
	if r.InternalDescription != nil {
		fields = append(fields, "internal_description")
	}
	if r.Meaning != nil {
		fields = append(fields, "meaning")
	}
	if r.AckRequiredNotNormal != nil {
		fields = append(fields, "ack_required_not_normal")
	}
	if r.AckRequiredError != nil {
		fields = append(fields, "ack_required_error")
	}
	if r.AckRequiredNormal != nil {
		fields = append(fields, "ack_required_normal")
	}
	if r.NormNotNormal != nil {
		fields = append(fields, "norm_not_normal")
	}
	if r.NormError != nil {
		fields = append(fields, "norm_error")
	}
	if r.NormNormal != nil {
		fields = append(fields, "norm_normal")
	}
	return fields
}

func (r UpdateAlarmDefinitionRequest) ChangedFields() []string {
	fields := make([]string, 0, 3)
	if r.Name != nil {
		fields = append(fields, "name")
	}
	if r.AlarmNote != nil {
		fields = append(fields, "alarm_note")
	}
	if r.AlarmTypeID != nil {
		fields = append(fields, "alarm_type_id")
	}
	return fields
}

func (r UpdateUnitRequest) ChangedFields() []string {
	fields := make([]string, 0, 3)
	if r.Code != nil {
		fields = append(fields, "code")
	}
	if r.Symbol != nil {
		fields = append(fields, "symbol")
	}
	if r.Name != nil {
		fields = append(fields, "name")
	}
	return fields
}

func (r UpdateAlarmFieldRequest) ChangedFields() []string {
	fields := make([]string, 0, 4)
	if r.Key != nil {
		fields = append(fields, "key")
	}
	if r.Label != nil {
		fields = append(fields, "label")
	}
	if r.DataType != nil {
		fields = append(fields, "data_type")
	}
	if r.DefaultUnitCode != nil {
		fields = append(fields, "default_unit_code")
	}
	return fields
}

func (r UpdateAlarmTypeFieldRequest) ChangedFields() []string {
	fields := make([]string, 0, 7)
	if r.DisplayOrder != nil {
		fields = append(fields, "display_order")
	}
	if r.IsRequired != nil {
		fields = append(fields, "is_required")
	}
	if r.IsUserEditable != nil {
		fields = append(fields, "is_user_editable")
	}
	if r.DefaultValueJSON != nil {
		fields = append(fields, "default_value_json")
	}
	if r.ValidationJSON != nil {
		fields = append(fields, "validation_json")
	}
	if r.DefaultUnitID != nil {
		fields = append(fields, "default_unit_id")
	}
	if r.UIGroup != nil {
		fields = append(fields, "ui_group")
	}
	return fields
}

func (r UpdateAlarmTypeRequest) ChangedFields() []string {
	if r.Name == nil {
		return []string{}
	}
	return []string{"name"}
}

func (r UpdateSpecificationRequest) ChangedFields() []string {
	fields := make([]string, 0, 11)
	if r.SpecificationSupplier != nil {
		fields = append(fields, "specification_supplier")
	}
	if r.SpecificationBrand != nil {
		fields = append(fields, "specification_brand")
	}
	if r.SpecificationType != nil {
		fields = append(fields, "specification_type")
	}
	if r.AdditionalInfoMotorValve != nil {
		fields = append(fields, "additional_info_motor_valve")
	}
	if r.AdditionalInfoSize != nil {
		fields = append(fields, "additional_info_size")
	}
	if r.AdditionalInformationInstallationLocation != nil {
		fields = append(fields, "additional_information_installation_location")
	}
	if r.ElectricalConnectionPH != nil {
		fields = append(fields, "electrical_connection_ph")
	}
	if r.ElectricalConnectionACDC != nil {
		fields = append(fields, "electrical_connection_acdc")
	}
	if r.ElectricalConnectionAmperage != nil {
		fields = append(fields, "electrical_connection_amperage")
	}
	if r.ElectricalConnectionPower != nil {
		fields = append(fields, "electrical_connection_power")
	}
	if r.ElectricalConnectionRotation != nil {
		fields = append(fields, "electrical_connection_rotation")
	}
	return fields
}

func (r SpecificationInput) ChangedFields() []string {
	fields := make([]string, 0, 11)
	if r.SpecificationSupplier.Set {
		fields = append(fields, "specification_supplier")
	}
	if r.SpecificationBrand.Set {
		fields = append(fields, "specification_brand")
	}
	if r.SpecificationType.Set {
		fields = append(fields, "specification_type")
	}
	if r.AdditionalInfoMotorValve.Set {
		fields = append(fields, "additional_info_motor_valve")
	}
	if r.AdditionalInfoSize.Set {
		fields = append(fields, "additional_info_size")
	}
	if r.AdditionalInformationInstallationLocation.Set {
		fields = append(fields, "additional_information_installation_location")
	}
	if r.ElectricalConnectionPH.Set {
		fields = append(fields, "electrical_connection_ph")
	}
	if r.ElectricalConnectionACDC.Set {
		fields = append(fields, "electrical_connection_acdc")
	}
	if r.ElectricalConnectionAmperage.Set {
		fields = append(fields, "electrical_connection_amperage")
	}
	if r.ElectricalConnectionPower.Set {
		fields = append(fields, "electrical_connection_power")
	}
	if r.ElectricalConnectionRotation.Set {
		fields = append(fields, "electrical_connection_rotation")
	}
	return fields
}
