package project

import (
	"errors"
	"fmt"
	"strings"

	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
)

var ErrUnknownProjectEvent = errors.New("unknown project event")

type projectEventSemantics struct {
	aggregate     domainProject.AggregateKind
	action        domainProject.ChangeAction
	changedFields []string
}

var projectEventVocabulary = map[string]projectEventSemantics{
	"project.created":       eventSemantics(domainProject.AggregateProject, domainProject.ChangeCreated),
	"project.updated":       updateSemantics(domainProject.AggregateProject, "name", "description", "status", "start_date", "phase_id"),
	"project.deleted":       eventSemantics(domainProject.AggregateProject, domainProject.ChangeDeleted),
	"project.phase.changed": updateSemantics(domainProject.AggregateProject, "phase_id"),

	"project.user.invited": eventSemantics(domainProject.AggregateProjectUser, domainProject.ChangeInvited),
	"project.user.removed": eventSemantics(domainProject.AggregateProjectUser, domainProject.ChangeRemoved),

	"project.control_cabinet.created": eventSemantics(domainProject.AggregateControlCabinet, domainProject.ChangeCreated),
	"project.control_cabinet.updated": updateSemantics(domainProject.AggregateControlCabinet, "building_id", "control_cabinet_nr"),
	"project.control_cabinet.deleted": eventSemantics(domainProject.AggregateControlCabinet, domainProject.ChangeDeleted),
	"project.control_cabinet.copied":  eventSemantics(domainProject.AggregateControlCabinet, domainProject.ChangeCopied),

	"project.sps_controller.created":            eventSemantics(domainProject.AggregateSPSController, domainProject.ChangeCreated),
	"project.sps_controller.updated":            updateSemantics(domainProject.AggregateSPSController, "control_cabinet_id", "ga_device", "device_name", "device_description", "device_location", "ip_address", "subnet", "gateway", "vlan", "system_types"),
	"project.sps_controller.deleted":            eventSemantics(domainProject.AggregateSPSController, domainProject.ChangeDeleted),
	"project.sps_controller.copied":             eventSemantics(domainProject.AggregateSPSController, domainProject.ChangeCopied),
	"project.sps_controller.ip_address.changed": updateSemantics(domainProject.AggregateSPSController, "ip_address"),

	"project.sps_controller_system_type.created": eventSemantics(domainProject.AggregateSPSControllerSystemType, domainProject.ChangeCreated),
	"project.sps_controller_system_type.updated": updateSemantics(domainProject.AggregateSPSControllerSystemType, "number", "document_name"),
	"project.sps_controller_system_type.deleted": eventSemantics(domainProject.AggregateSPSControllerSystemType, domainProject.ChangeDeleted),
	"project.sps_controller_system_type.copied":  eventSemantics(domainProject.AggregateSPSControllerSystemType, domainProject.ChangeCopied),

	"project.field_device.created":       eventSemantics(domainProject.AggregateFieldDevice, domainProject.ChangeCreated),
	"project.field_device.multi_created": eventSemantics(domainProject.AggregateFieldDevice, domainProject.ChangeCreated),
	"project.field_device.updated":       updateSemantics(domainProject.AggregateFieldDevice, "bmk", "description", "text_fix", "apparat_nr", "sps_controller_system_type_id", "system_part_id", "apparat_id", "specification", "bacnet_objects"),
	"project.field_device.deleted":       eventSemantics(domainProject.AggregateFieldDevice, domainProject.ChangeDeleted),
	"project.field_device.copied":        eventSemantics(domainProject.AggregateFieldDevice, domainProject.ChangeCopied),

	"project.object_data.created": eventSemantics(domainProject.AggregateObjectData, domainProject.ChangeCreated),
	"project.object_data.updated": eventSemantics(domainProject.AggregateObjectData, domainProject.ChangeUpdated),
	"project.object_data.deleted": eventSemantics(domainProject.AggregateObjectData, domainProject.ChangeDeleted),
}

func eventSemantics(aggregate domainProject.AggregateKind, action domainProject.ChangeAction) projectEventSemantics {
	return projectEventSemantics{aggregate: aggregate, action: action, changedFields: []string{}}
}

func updateSemantics(aggregate domainProject.AggregateKind, fields ...string) projectEventSemantics {
	return projectEventSemantics{aggregate: aggregate, action: domainProject.ChangeUpdated, changedFields: fields}
}

func changeSemantics(eventType string) (projectEventSemantics, error) {
	eventKey := strings.TrimSpace(eventType)
	semantics, ok := projectEventVocabulary[eventKey]
	if !ok {
		return projectEventSemantics{}, fmt.Errorf("%w: %q", ErrUnknownProjectEvent, eventKey)
	}
	return semantics, nil
}
