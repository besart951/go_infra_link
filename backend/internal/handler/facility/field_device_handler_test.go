package facility

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appfielddevice "github.com/besart951/go_infra_link/backend/internal/application/facility/fielddevice"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestListFieldDevicesReturnsAfterInvalidFilterParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeFieldDeviceHandlerService{}
	handler := NewFieldDeviceHandler(service, nil, nil, service, nil, nil)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	tracker := &fieldDeviceStatusTrackingWriter{ResponseWriter: context.Writer}
	context.Writer = tracker
	context.Request = httptest.NewRequest(http.MethodGet, "/field-devices?building_id=not-a-uuid", nil)

	handler.ListFieldDevices(context)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.listWithFiltersCalls != 0 {
		t.Fatalf("expected service not to be called after invalid query param, got %d call(s)", service.listWithFiltersCalls)
	}
	if len(tracker.statusWrites) != 1 {
		t.Fatalf("expected exactly one status write, got %v", tracker.statusWrites)
	}
	if tracker.statusWrites[0] != http.StatusBadRequest {
		t.Fatalf("expected only status write to be 400, got %v", tracker.statusWrites)
	}
}

func TestMultiCreateFieldDevicesBindJSONReturnsNestedValidationShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeFieldDeviceHandlerService{}
	handler := NewFieldDeviceHandler(service, nil, nil, service, nil, nil)
	systemTypeID := uuid.NewString()
	systemPartID := uuid.NewString()
	apparatID := uuid.NewString()
	body := `{
		"field_devices": [{
			"apparat_nr": 1,
			"sps_controller_system_type_id": "` + systemTypeID + `",
			"system_part_id": "` + systemPartID + `",
			"apparat_id": "` + apparatID + `",
			"bacnet_objects": [{
				"software_type": "AI",
				"software_number": 1
			}]
		}]
	}`

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/field-devices/multi-create", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	context.Request.Header.Set("X-Request-ID", "req-fd-1")

	handler.MultiCreateFieldDevices(context)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.multiCreateCalls != 0 {
		t.Fatalf("expected service not to be called after bind failure, got %d", service.multiCreateCalls)
	}

	var response dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected error response to decode, got %v", err)
	}
	path := "field_devices[0].bacnet_objects[0].text_fix"
	if response.Code != "validation_error" || response.Error != "validation_error" {
		t.Fatalf("expected validation code and compatibility error, got %+v", response)
	}
	if response.Fields[path] != "is required" {
		t.Fatalf("expected nested compatibility field path %q, got %+v", path, response.Fields)
	}
	if len(response.FieldErrors) != 1 || response.FieldErrors[0].Path != path {
		t.Fatalf("expected nested field_errors path %q, got %+v", path, response.FieldErrors)
	}
	if response.RequestID != "req-fd-1" {
		t.Fatalf("expected request id to be echoed, got %+v", response)
	}
}

func TestMultiCreateFieldDevicesUsesTypedApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)

	firstID := uuid.New()
	secondID := uuid.New()
	firstSystemTypeID := uuid.New()
	secondSystemTypeID := uuid.New()
	firstSystemPartID := uuid.New()
	secondSystemPartID := uuid.New()
	firstApparatID := uuid.New()
	secondApparatID := uuid.New()
	objectDataID := uuid.New()
	service := &fakeFieldDeviceHandlerService{}
	multiCreator := &fakeFieldDeviceMultiCreator{
		result: &domainFacility.FieldDeviceMultiCreateResult{
			Results: []domainFacility.FieldDeviceCreateResult{
				{Index: 0, Success: true, FieldDevice: &domainFacility.FieldDevice{
					Base:                      domain.Base{ID: firstID},
					ApparatNr:                 1,
					SPSControllerSystemTypeID: firstSystemTypeID,
					SystemPartID:              firstSystemPartID,
					ApparatID:                 firstApparatID,
				}},
				{Index: 1, Success: true, FieldDevice: &domainFacility.FieldDevice{
					Base:                      domain.Base{ID: secondID},
					ApparatNr:                 2,
					SPSControllerSystemTypeID: secondSystemTypeID,
					SystemPartID:              secondSystemPartID,
					ApparatID:                 secondApparatID,
				}},
			},
			TotalRequests: 2,
			SuccessCount:  2,
		},
	}
	handler := NewFieldDeviceHandler(service, nil, nil, service, multiCreator, nil)
	body := `{"field_devices":[` +
		`{"apparat_nr":1,"sps_controller_system_type_id":"` + firstSystemTypeID.String() +
		`","system_part_id":"` + firstSystemPartID.String() +
		`","apparat_id":"` + firstApparatID.String() +
		`","object_data_id":"` + objectDataID.String() + `"},` +
		`{"apparat_nr":2,"sps_controller_system_type_id":"` + secondSystemTypeID.String() +
		`","system_part_id":"` + secondSystemPartID.String() +
		`","apparat_id":"` + secondApparatID.String() +
		`","bacnet_objects":[{"text_fix":"AI1","software_type":"ai","software_number":1}]}` +
		`]}`
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(
		http.MethodPost,
		"/field-devices/multi-create",
		strings.NewReader(body),
	)
	ginContext.Request.Header.Set("Content-Type", "application/json")

	handler.MultiCreateFieldDevices(ginContext)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if multiCreator.calls != 1 || len(multiCreator.command.Items) != 2 {
		t.Fatalf("multi-create command: calls=%d command=%+v", multiCreator.calls, multiCreator.command)
	}
	first := multiCreator.command.Items[0]
	second := multiCreator.command.Items[1]
	if first.FieldDevice == nil || first.FieldDevice.SPSControllerSystemTypeID != firstSystemTypeID ||
		first.ObjectDataID == nil || *first.ObjectDataID != objectDataID || len(first.BacnetObjects) != 0 {
		t.Fatalf("ObjectData create item: %+v", first)
	}
	if second.FieldDevice == nil || second.FieldDevice.SPSControllerSystemTypeID != secondSystemTypeID ||
		second.ObjectDataID != nil || len(second.BacnetObjects) != 1 ||
		second.BacnetObjects[0].TextFix != "AI1" ||
		second.BacnetObjects[0].SoftwareType != domainFacility.BacnetSoftwareTypeAI {
		t.Fatalf("explicit BACnet create item: %+v", second)
	}
	if service.multiCreateCalls != 0 {
		t.Fatalf("legacy MultiCreate called directly %d times", service.multiCreateCalls)
	}
	var response dto.MultiCreateFieldDeviceResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.TotalRequests != 2 || response.SuccessCount != 2 ||
		len(response.Results) != 2 || response.Results[0].FieldDevice == nil ||
		response.Results[0].FieldDevice.ID != firstID || response.Results[1].FieldDevice == nil ||
		response.Results[1].FieldDevice.ID != secondID {
		t.Fatalf("legacy response changed: %+v", response)
	}
	if strings.Contains(recorder.Body.String(), "operation_id") ||
		strings.Contains(recorder.Body.String(), "batch_id") {
		t.Fatalf("internal mutation metadata leaked: %s", recorder.Body.String())
	}
}

func TestListFieldDevicesAcceptsMultiValueFilterParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	firstBuildingID := uuid.New()
	secondBuildingID := uuid.New()
	service := &fakeFieldDeviceHandlerService{}
	handler := NewFieldDeviceHandler(service, nil, nil, service, nil, nil)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodGet,
		"/field-devices?building_id="+firstBuildingID.String()+"|"+secondBuildingID.String(),
		nil,
	)

	handler.ListFieldDevices(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.listWithFiltersCalls != 1 {
		t.Fatalf("expected service to be called once, got %d call(s)", service.listWithFiltersCalls)
	}
	if len(service.lastFilters.BuildingIDs) != 2 {
		t.Fatalf("expected two building ids, got %+v", service.lastFilters.BuildingIDs)
	}
	if service.lastFilters.BuildingIDs[0] != firstBuildingID || service.lastFilters.BuildingIDs[1] != secondBuildingID {
		t.Fatalf("unexpected building ids: %+v", service.lastFilters.BuildingIDs)
	}
}

func TestToFieldDeviceSpecificationPatchPreservesExplicitNull(t *testing.T) {
	var req dto.UpdateFieldDeviceSpecificationRequest
	if err := json.Unmarshal([]byte(`{"specification_supplier":null,"specification_brand":"Replacement"}`), &req); err != nil {
		t.Fatalf("expected specification patch request to decode, got %v", err)
	}

	patch := toFieldDeviceSpecificationPatch(req)
	if patch == nil {
		t.Fatal("expected patch to be created")
	}
	if !patch.HasSpecificationSupplier || patch.SpecificationSupplier != nil {
		t.Fatalf("expected explicit null supplier to be preserved, got %+v", patch)
	}
	if !patch.HasSpecificationBrand || patch.SpecificationBrand == nil || *patch.SpecificationBrand != "Replacement" {
		t.Fatalf("expected replacement brand to be preserved, got %+v", patch)
	}
}

func TestListAvailableApparatNumbersSuppressesCanceledRequestError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeFieldDeviceHandlerService{listAvailableErr: context.Canceled}
	handler := NewFieldDeviceHandler(service, nil, nil, service, nil, nil)

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(
		http.MethodGet,
		"/field-devices/available-apparat-nr?sps_controller_system_type_id="+uuid.NewString()+"&apparat_id="+uuid.NewString()+"&system_part_id="+uuid.NewString(),
		nil,
	)

	handler.ListAvailableApparatNumbers(ginContext)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected no response to be written, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", recorder.Body.String())
	}
	if !ginContext.IsAborted() {
		t.Fatal("expected request to be aborted")
	}
	if service.listAvailableCalls != 1 {
		t.Fatalf("expected service to be called once, got %d", service.listAvailableCalls)
	}
}

func TestListFieldDevicesSuppressesCanceledRequestError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeFieldDeviceHandlerService{listWithFiltersErr: context.Canceled}
	handler := NewFieldDeviceHandler(service, nil, nil, service, nil, nil)

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodGet, "/field-devices?page=1&limit=20", nil)

	handler.ListFieldDevices(ginContext)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected no response to be written, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", recorder.Body.String())
	}
	if !ginContext.IsAborted() {
		t.Fatal("expected request to be aborted")
	}
	if service.listWithFiltersCalls != 1 {
		t.Fatalf("expected service to be called once, got %d", service.listWithFiltersCalls)
	}
}

func TestBulkUpdateFieldDevicesUsesApplicationBulkUpdater(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeFieldDeviceHandlerService{}
	bulkUpdater := &fakeFieldDeviceBulkUpdater{
		result: &domainFacility.BulkOperationResult{
			TotalCount:   1,
			SuccessCount: 1,
			Results: []domainFacility.BulkOperationResultItem{
				{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Success: true},
			},
		},
	}
	handler := NewFieldDeviceHandler(service, nil, nil, bulkUpdater, nil, nil)
	body := `{"updates":[{"id":"00000000-0000-0000-0000-000000000001","bmk":"B-1"}]}`

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(
		http.MethodPatch,
		"/field-devices/bulk-update",
		strings.NewReader(body),
	)
	ginContext.Request.Header.Set("Content-Type", "application/json")

	handler.BulkUpdateFieldDevices(ginContext)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, internalKey := range []string{"operation_id", "batch_id", "phases"} {
		if strings.Contains(recorder.Body.String(), `"`+internalKey+`"`) {
			t.Fatalf("legacy bulk response leaked internal key %q: %s", internalKey, recorder.Body.String())
		}
	}
	if bulkUpdater.calls != 1 {
		t.Fatalf("expected application bulk updater once, got %d", bulkUpdater.calls)
	}
	if service.bulkUpdateCalls != 0 {
		t.Fatalf("legacy service BulkUpdate must not be called directly, got %d", service.bulkUpdateCalls)
	}
}

func TestBulkDeleteFieldDevicesUsesTypedApplicationCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)

	firstID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	secondID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	service := &fakeFieldDeviceHandlerService{}
	bulkDeleter := &fakeFieldDeviceBulkDeleter{
		result: &domainFacility.BulkOperationResult{
			Results: []domainFacility.BulkOperationResultItem{
				{ID: firstID, Success: true},
				{ID: secondID, Success: false, Error: "delete failed"},
			},
			TotalCount:   2,
			SuccessCount: 1,
			FailureCount: 1,
		},
	}
	handler := NewFieldDeviceHandler(service, nil, nil, service, nil, bulkDeleter)
	body := `{"ids":["` + firstID.String() + `","` + secondID.String() + `"]}`

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(
		http.MethodDelete,
		"/field-devices/bulk-delete",
		strings.NewReader(body),
	)
	ginContext.Request.Header.Set("Content-Type", "application/json")

	handler.BulkDeleteFieldDevices(ginContext)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if bulkDeleter.calls != 1 {
		t.Fatalf("expected application bulk deleter once, got %d", bulkDeleter.calls)
	}
	if len(bulkDeleter.command.FieldDeviceIDs) != 2 ||
		bulkDeleter.command.FieldDeviceIDs[0] != firstID ||
		bulkDeleter.command.FieldDeviceIDs[1] != secondID {
		t.Fatalf("unexpected typed bulk-delete command: %+v", bulkDeleter.command)
	}
	if service.bulkDeleteCalls != 0 {
		t.Fatalf("legacy service BulkDelete must not be called directly, got %d", service.bulkDeleteCalls)
	}

	var response dto.BulkDeleteFieldDeviceResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.TotalCount != 2 || response.SuccessCount != 1 || response.FailureCount != 1 ||
		len(response.Results) != 2 || response.Results[0].ID != firstID ||
		!response.Results[0].Success || response.Results[1].ID != secondID ||
		response.Results[1].Success || response.Results[1].Error != "delete failed" {
		t.Fatalf("legacy response changed: %+v", response)
	}
	for _, internalKey := range []string{"operation_id", "batch_id", "changes"} {
		if strings.Contains(recorder.Body.String(), `"`+internalKey+`"`) {
			t.Fatalf("legacy bulk response leaked internal key %q: %s", internalKey, recorder.Body.String())
		}
	}
}

func TestUpdateFieldDeviceUsesApplicationUpdaterAndPreservesEmptyReplacement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fieldDeviceID := uuid.New()
	systemPartID := uuid.New()
	service := &fakeFieldDeviceHandlerService{}
	updater := &fakeFieldDeviceUpdater{
		result: &domainFacility.FieldDevice{
			Base:         domain.Base{ID: fieldDeviceID},
			SystemPartID: systemPartID,
		},
	}
	handler := NewFieldDeviceHandler(service, updater, nil, service, nil, nil)
	body := `{
		"bmk": "B-1",
		"system_part_id": "` + systemPartID.String() + `",
		"bacnet_objects": []
	}`

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(
		http.MethodPut,
		"/field-devices/"+fieldDeviceID.String(),
		strings.NewReader(body),
	)
	ginContext.Params = gin.Params{{Key: "id", Value: fieldDeviceID.String()}}
	ginContext.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateFieldDevice(ginContext)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if updater.calls != 1 {
		t.Fatalf("expected application updater once, got %d", updater.calls)
	}
	if service.getByIDCalls != 0 || service.updateWithBacnetCalls != 0 {
		t.Fatalf(
			"legacy update path must not be called, got get=%d update=%d",
			service.getByIDCalls,
			service.updateWithBacnetCalls,
		)
	}
	if updater.command.FieldDeviceID != fieldDeviceID {
		t.Fatalf("expected field device %s, got %s", fieldDeviceID, updater.command.FieldDeviceID)
	}
	if updater.command.BMK == nil || *updater.command.BMK != "B-1" {
		t.Fatalf("expected BMK update, got %v", updater.command.BMK)
	}
	if updater.command.SystemPartID == nil || *updater.command.SystemPartID != systemPartID {
		t.Fatalf("expected system part %s, got %v", systemPartID, updater.command.SystemPartID)
	}
	if updater.command.BacnetObjects == nil || len(*updater.command.BacnetObjects) != 0 {
		t.Fatalf("expected explicit empty BACnet replacement, got %v", updater.command.BacnetObjects)
	}
	if updater.command.ObjectDataID != nil {
		t.Fatalf("expected no ObjectData replacement, got %s", *updater.command.ObjectDataID)
	}
}

func TestUpdateFieldDeviceMapsInitialApplicationLoadNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fieldDeviceID := uuid.New()
	systemPartID := uuid.New()
	updater := &fakeFieldDeviceUpdater{
		err: &appfielddevice.LoadError{Err: domain.ErrNotFound},
	}
	handler := NewFieldDeviceHandler(
		&fakeFieldDeviceHandlerService{},
		updater,
		nil,
		&fakeFieldDeviceBulkUpdater{},
		nil,
		nil,
	)
	body := `{"system_part_id":"` + systemPartID.String() + `"}`

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(
		http.MethodPut,
		"/field-devices/"+fieldDeviceID.String(),
		strings.NewReader(body),
	)
	ginContext.Params = gin.Params{{Key: "id", Value: fieldDeviceID.String()}}
	ginContext.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateFieldDevice(ginContext)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestDeleteFieldDeviceRoutesTypedCommandWithoutLegacyDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fieldDeviceID := uuid.New()
	service := &fakeFieldDeviceHandlerService{}
	deleter := &fakeFieldDeviceDeleter{}
	handler := NewFieldDeviceHandler(
		service,
		nil,
		deleter,
		&fakeFieldDeviceBulkUpdater{},
		nil,
		nil,
	)

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(
		http.MethodDelete,
		"/field-devices/"+fieldDeviceID.String(),
		nil,
	)
	ginContext.Params = gin.Params{{Key: "id", Value: fieldDeviceID.String()}}

	handler.DeleteFieldDevice(ginContext)

	if ginContext.Writer.Status() != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d body=%s", ginContext.Writer.Status(), recorder.Body.String())
	}
	if deleter.calls != 1 || deleter.command.FieldDeviceID != fieldDeviceID {
		t.Fatalf("unexpected typed delete: %+v", deleter)
	}
	if service.deleteCalls != 0 {
		t.Fatalf("legacy DeleteByID called %d times", service.deleteCalls)
	}
}

func TestFieldDeviceUpdateMapperDistinguishesOmittedAndEmptyBacnetObjects(t *testing.T) {
	fieldDeviceID := uuid.New()
	systemPartID := uuid.New()

	var omitted dto.UpdateFieldDeviceRequest
	if err := json.Unmarshal(
		[]byte(`{"system_part_id":"`+systemPartID.String()+`"}`),
		&omitted,
	); err != nil {
		t.Fatalf("decode omitted request: %v", err)
	}
	if command := toFieldDeviceUpdateCommand(fieldDeviceID, omitted); command.BacnetObjects != nil {
		t.Fatalf("omitted bacnet_objects must preserve children, got %v", command.BacnetObjects)
	}

	var explicitEmpty dto.UpdateFieldDeviceRequest
	if err := json.Unmarshal(
		[]byte(`{"system_part_id":"`+systemPartID.String()+`","bacnet_objects":[]}`),
		&explicitEmpty,
	); err != nil {
		t.Fatalf("decode explicit empty request: %v", err)
	}
	command := toFieldDeviceUpdateCommand(fieldDeviceID, explicitEmpty)
	if command.BacnetObjects == nil || len(*command.BacnetObjects) != 0 {
		t.Fatalf("empty bacnet_objects must replace children with none, got %v", command.BacnetObjects)
	}
}

type fieldDeviceStatusTrackingWriter struct {
	gin.ResponseWriter
	statusWrites []int
}

func (w *fieldDeviceStatusTrackingWriter) WriteHeader(code int) {
	w.statusWrites = append(w.statusWrites, code)
	w.ResponseWriter.WriteHeader(code)
}

type fakeFieldDeviceHandlerService struct {
	listWithFiltersCalls  int
	listAvailableCalls    int
	multiCreateCalls      int
	bulkUpdateCalls       int
	bulkDeleteCalls       int
	getByIDCalls          int
	updateWithBacnetCalls int
	deleteCalls           int
	listWithFiltersErr    error
	listAvailableErr      error
	lastFilters           domainFacility.FieldDeviceFilterParams
}

func (s *fakeFieldDeviceHandlerService) Create(context.Context, *domainFacility.FieldDevice) error {
	return nil
}

func (s *fakeFieldDeviceHandlerService) CreateWithBacnetObjects(context.Context, *domainFacility.FieldDevice, *uuid.UUID, []domainFacility.BacnetObject) error {
	return nil
}

func (s *fakeFieldDeviceHandlerService) MultiCreate(context.Context, []domainFacility.FieldDeviceCreateItem) *domainFacility.FieldDeviceMultiCreateResult {
	s.multiCreateCalls++
	return &domainFacility.FieldDeviceMultiCreateResult{}
}

func (s *fakeFieldDeviceHandlerService) GetByID(context.Context, uuid.UUID) (*domainFacility.FieldDevice, error) {
	s.getByIDCalls++
	return nil, domain.ErrNotFound
}

func (s *fakeFieldDeviceHandlerService) List(context.Context, int, int, string) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	return &domain.PaginatedList[domainFacility.FieldDevice]{}, nil
}

func (s *fakeFieldDeviceHandlerService) ListWithFilters(_ context.Context, _ domain.PaginationParams, filters domainFacility.FieldDeviceFilterParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	s.listWithFiltersCalls++
	s.lastFilters = filters
	if s.listWithFiltersErr != nil {
		return nil, s.listWithFiltersErr
	}
	return &domain.PaginatedList[domainFacility.FieldDevice]{
		Items:      []domainFacility.FieldDevice{},
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (s *fakeFieldDeviceHandlerService) ListAvailableApparatNumbers(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]int, error) {
	s.listAvailableCalls++
	if s.listAvailableErr != nil {
		return nil, s.listAvailableErr
	}
	return nil, nil
}

func (s *fakeFieldDeviceHandlerService) GetFieldDeviceOptions(context.Context) (*domainFacility.FieldDeviceOptions, error) {
	return &domainFacility.FieldDeviceOptions{}, nil
}

func (s *fakeFieldDeviceHandlerService) Update(context.Context, *domainFacility.FieldDevice) error {
	return nil
}

func (s *fakeFieldDeviceHandlerService) UpdateWithBacnetObjects(context.Context, *domainFacility.FieldDevice, *uuid.UUID, *[]domainFacility.BacnetObject) error {
	s.updateWithBacnetCalls++
	return nil
}

func (s *fakeFieldDeviceHandlerService) DeleteByID(context.Context, uuid.UUID) error {
	s.deleteCalls++
	return nil
}

func (s *fakeFieldDeviceHandlerService) ListBacnetObjects(context.Context, uuid.UUID) ([]domainFacility.BacnetObject, error) {
	return nil, nil
}

func (s *fakeFieldDeviceHandlerService) CreateSpecification(context.Context, uuid.UUID, *domainFacility.Specification) error {
	return nil
}

func (s *fakeFieldDeviceHandlerService) UpdateSpecificationPatch(context.Context, uuid.UUID, *domainFacility.SpecificationPatch) (*domainFacility.Specification, error) {
	return nil, nil
}

func (s *fakeFieldDeviceHandlerService) BulkUpdate(context.Context, []domainFacility.BulkFieldDeviceUpdate) *domainFacility.BulkOperationResult {
	s.bulkUpdateCalls++
	return &domainFacility.BulkOperationResult{}
}

func (s *fakeFieldDeviceHandlerService) BulkDelete(context.Context, []uuid.UUID) *domainFacility.BulkOperationResult {
	s.bulkDeleteCalls++
	return &domainFacility.BulkOperationResult{}
}

type fakeFieldDeviceBulkUpdater struct {
	result *domainFacility.BulkOperationResult
	calls  int
}

type fakeFieldDeviceMultiCreator struct {
	command appfielddevice.MultiCreateCommand
	result  *domainFacility.FieldDeviceMultiCreateResult
	calls   int
}

type fakeFieldDeviceBulkDeleter struct {
	command appfielddevice.BulkDeleteCommand
	result  *domainFacility.BulkOperationResult
	calls   int
}

func (s *fakeFieldDeviceBulkDeleter) BulkDelete(
	_ context.Context,
	command appfielddevice.BulkDeleteCommand,
) *domainFacility.BulkOperationResult {
	s.calls++
	s.command = command
	return s.result
}

func (s *fakeFieldDeviceMultiCreator) MultiCreate(
	_ context.Context,
	command appfielddevice.MultiCreateCommand,
) *domainFacility.FieldDeviceMultiCreateResult {
	s.calls++
	s.command = command
	return s.result
}

type fakeFieldDeviceUpdater struct {
	command appfielddevice.UpdateCommand
	result  *domainFacility.FieldDevice
	err     error
	calls   int
}

type fakeFieldDeviceDeleter struct {
	command appfielddevice.DeleteCommand
	err     error
	calls   int
}

func (s *fakeFieldDeviceDeleter) Delete(
	_ context.Context,
	command appfielddevice.DeleteCommand,
) error {
	s.calls++
	s.command = command
	return s.err
}

func (s *fakeFieldDeviceUpdater) Update(
	_ context.Context,
	command appfielddevice.UpdateCommand,
) (*domainFacility.FieldDevice, error) {
	s.calls++
	s.command = command
	return s.result, s.err
}

func (s *fakeFieldDeviceBulkUpdater) BulkUpdate(
	context.Context,
	[]domainFacility.BulkFieldDeviceUpdate,
) *domainFacility.BulkOperationResult {
	s.calls++
	return s.result
}
