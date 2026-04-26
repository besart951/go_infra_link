package facility

import (
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func TestValidateAlarmDefinition_RequiresAlarmTypeAndName(t *testing.T) {
	if err := validateAlarmDefinition(nil); err != domain.ErrInvalidArgument {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}

	item := &domainFacility.AlarmDefinition{}
	err := validateAlarmDefinition(item)
	ve, ok := err.(*domain.ValidationError)
	if !ok {
		t.Fatalf("expected validation error, got %T (%v)", err, err)
	}
	if _, exists := ve.Fields["alarmdefinition.name"]; !exists {
		t.Fatalf("expected name validation error, got %+v", ve.Fields)
	}
	if _, exists := ve.Fields["alarmdefinition.alarm_type_id"]; !exists {
		t.Fatalf("expected alarm_type_id validation error, got %+v", ve.Fields)
	}

	alarmTypeID := uuid.New()
	item = &domainFacility.AlarmDefinition{Name: "Limit Alarm", AlarmTypeID: &alarmTypeID}
	if err := validateAlarmDefinition(item); err != nil {
		t.Fatalf("expected valid alarm definition, got %v", err)
	}
}
