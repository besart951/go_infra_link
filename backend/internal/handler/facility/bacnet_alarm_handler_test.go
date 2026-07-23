package facility

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appbacnetobject "github.com/besart951/go_infra_link/backend/internal/application/facility/bacnetobject"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type bacnetAlarmReadServiceStub struct {
	getValuesCalls int
	putValuesCalls int
}

func (*bacnetAlarmReadServiceStub) GetSchema(
	context.Context,
	uuid.UUID,
) (*domainFacility.AlarmType, error) {
	return nil, nil
}

func (s *bacnetAlarmReadServiceStub) GetValues(
	context.Context,
	uuid.UUID,
) ([]domainFacility.BacnetObjectAlarmValue, error) {
	s.getValuesCalls++
	return nil, nil
}

func (s *bacnetAlarmReadServiceStub) PutValues(
	context.Context,
	uuid.UUID,
	[]domainFacility.BacnetObjectAlarmValue,
) error {
	s.putValuesCalls++
	return nil
}

type bacnetAlarmValueReplacerStub struct {
	command appbacnetobject.ReplaceAlarmValuesCommand
	values  []domainFacility.BacnetObjectAlarmValue
	err     error
	calls   int
}

func (s *bacnetAlarmValueReplacerStub) ReplaceAlarmValues(
	_ context.Context,
	command appbacnetobject.ReplaceAlarmValuesCommand,
) ([]domainFacility.BacnetObjectAlarmValue, error) {
	s.calls++
	s.command = command
	return s.values, s.err
}

func TestPutAlarmValuesUsesTypedApplicationCommandAndCommittedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	bacnetObjectID := uuid.New()
	alarmTypeFieldID := uuid.New()
	unitID := uuid.New()
	valueID := uuid.New()
	createdAt := time.Date(2026, time.July, 20, 23, 30, 0, 0, time.UTC)
	readService := &bacnetAlarmReadServiceStub{}
	replacer := &bacnetAlarmValueReplacerStub{values: []domainFacility.BacnetObjectAlarmValue{{
		Base: domain.Base{
			ID:        valueID,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
		BacnetObjectID:   bacnetObjectID,
		AlarmTypeFieldID: alarmTypeFieldID,
		ValueInteger:     int64Pointer(7),
		UnitID:           &unitID,
		Source:           domainFacility.AlarmValueSourceUser,
	}}}
	handler := NewBacnetAlarmHandler(readService, replacer)
	router := gin.New()
	router.PUT("/facility/bacnet-objects/:id/alarm-values", handler.PutAlarmValues)
	body := []byte(`{"values":[{"alarm_type_field_id":"` + alarmTypeFieldID.String() +
		`","value_integer":7,"unit_id":"` + unitID.String() + `"}]}`)
	request := httptest.NewRequest(
		http.MethodPut,
		"/facility/bacnet-objects/"+bacnetObjectID.String()+"/alarm-values",
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status: got %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if replacer.calls != 1 || readService.putValuesCalls != 0 || readService.getValuesCalls != 0 {
		t.Fatalf("routing: replacer=%d legacyPut=%d legacyGet=%d",
			replacer.calls,
			readService.putValuesCalls,
			readService.getValuesCalls,
		)
	}
	if replacer.command.BacnetObjectID != bacnetObjectID ||
		len(replacer.command.Values) != 1 ||
		replacer.command.Values[0].AlarmTypeFieldID != alarmTypeFieldID ||
		replacer.command.Values[0].ValueInteger == nil ||
		*replacer.command.Values[0].ValueInteger != 7 ||
		replacer.command.Values[0].UnitID == nil ||
		*replacer.command.Values[0].UnitID != unitID ||
		replacer.command.Values[0].Source != "" {
		t.Fatalf("typed command: %+v", replacer.command)
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"id":"`+valueID.String()+`"`)) ||
		!bytes.Contains(recorder.Body.Bytes(), []byte(`"source":"user"`)) {
		t.Fatalf("response did not use committed application result: %s", recorder.Body.String())
	}
}

func TestPutAlarmValuesPreservesFetchFailureMappingForTransactionalReload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	bacnetObjectID := uuid.New()
	alarmTypeFieldID := uuid.New()
	reloadErr := errors.New("reload failed")
	handler := NewBacnetAlarmHandler(
		&bacnetAlarmReadServiceStub{},
		&bacnetAlarmValueReplacerStub{err: &appbacnetobject.AlarmValuesReloadError{Err: reloadErr}},
	)
	router := gin.New()
	router.PUT("/facility/bacnet-objects/:id/alarm-values", handler.PutAlarmValues)
	body := []byte(`{"values":[{"alarm_type_field_id":"` + alarmTypeFieldID.String() + `"}]}`)
	request := httptest.NewRequest(
		http.MethodPut,
		"/facility/bacnet-objects/"+bacnetObjectID.String()+"/alarm-values",
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError ||
		!bytes.Contains(recorder.Body.Bytes(), []byte(`"code":"fetch_failed"`)) {
		t.Fatalf("reload mapping: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func int64Pointer(value int64) *int64 {
	return &value
}
