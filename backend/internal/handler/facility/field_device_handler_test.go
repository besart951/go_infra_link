package facility

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/besart951/go_infra_link/backend/internal/handler/middleware"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestListFieldDevicesReturnsAfterInvalidFilterParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeFieldDeviceHandlerService{}
	handler := NewFieldDeviceHandler(service)

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
	handler := NewFieldDeviceHandler(service)
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

func TestBulkDeleteFieldDevicesRequiresDurableJobsAboveSynchronousLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeFieldDeviceHandlerService{}
	handler := NewFieldDeviceHandler(service)
	ids := make([]uuid.UUID, 501)
	for i := range ids {
		ids[i] = uuid.New()
	}
	body, err := json.Marshal(versionedDeleteRequest(ids))
	if err != nil {
		t.Fatalf("marshal bulk delete request: %v", err)
	}

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodDelete, "/field-devices/bulk-delete", strings.NewReader(string(body)))
	context.Request.Header.Set("Content-Type", "application/json")

	handler.BulkDeleteFieldDevices(context)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.bulkDeleteCalls != 0 {
		t.Fatalf("expected service not to be called, got %d call(s)", service.bulkDeleteCalls)
	}
}

type versionedDeleteHandlerService struct {
	*fakeFieldDeviceHandlerService
	commands []domainFacility.FieldDeviceDeleteCommand
}

func (s *versionedDeleteHandlerService) Delete(_ context.Context, command domainFacility.FieldDeviceDeleteCommand) error {
	s.commands = append(s.commands, command)
	return nil
}

func (s *versionedDeleteHandlerService) BulkDeleteCommands(_ context.Context, commands []domainFacility.FieldDeviceDeleteCommand) *domainFacility.BulkOperationResult {
	s.commands = append(s.commands, commands...)
	return &domainFacility.BulkOperationResult{TotalCount: len(commands), SuccessCount: len(commands)}
}

func TestDeleteFieldDevicePassesRequestedBaseVersion(t *testing.T) {
	service := &versionedDeleteHandlerService{fakeFieldDeviceHandlerService: &fakeFieldDeviceHandlerService{}}
	handler := NewFieldDeviceHandler(service)
	id := uuid.New()
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: id.String()}}
	context.Request = httptest.NewRequest(http.MethodDelete, "/field-devices/"+id.String()+"?base_version=7", nil)

	handler.DeleteFieldDevice(context)

	if context.Writer.Status() != http.StatusNoContent || len(service.commands) != 1 {
		t.Fatalf("status=%d commands=%+v body=%s", context.Writer.Status(), service.commands, recorder.Body.String())
	}
	if service.commands[0].BaseVersion != 7 {
		t.Fatalf("base version was not passed: %+v", service.commands[0])
	}
	if recorder.Header().Get("Deprecation") != "" {
		t.Fatalf("versioned request was marked deprecated: %v", recorder.Header())
	}
}

func TestBulkDeleteRejectsLegacyIDs(t *testing.T) {
	service := &versionedDeleteHandlerService{fakeFieldDeviceHandlerService: &fakeFieldDeviceHandlerService{}}
	handler := NewFieldDeviceHandler(service)
	body, _ := json.Marshal(map[string]any{"ids": []uuid.UUID{uuid.New()}})
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodDelete, "/field-devices/bulk-delete", strings.NewReader(string(body)))
	context.Request.Header.Set("Content-Type", "application/json")

	handler.BulkDeleteFieldDevices(context)

	if recorder.Code != http.StatusBadRequest || recorder.Header().Get("Deprecation") != "" {
		t.Fatalf("status=%d headers=%v body=%s", recorder.Code, recorder.Header(), recorder.Body.String())
	}
}

func TestBulkDeleteFieldDevicesQueuesDurableJobAboveSynchronousLimit(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:bulk-handler?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open job database: %v", err)
	}
	if err := facilityservice.MigrateFacilityJobs(db); err != nil {
		t.Fatalf("migrate jobs: %v", err)
	}
	jobs := facilityservice.NewFacilityJobManagerWithDB(nil, db)
	t.Cleanup(jobs.Close)

	service := &fakeFieldDeviceHandlerService{}
	handler := NewFieldDeviceHandlerWithFacilityJobs(service, nil, jobs)
	request := versionedDeleteRequest(newFieldDeviceIDs(501))
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Set(middleware.ContextUserIDKey, uuid.New())
	ginContext.Request = httptest.NewRequest(http.MethodDelete, "/field-devices/bulk-delete", strings.NewReader(string(body)))
	ginContext.Request.Header.Set("Content-Type", "application/json")
	handler.BulkDeleteFieldDevices(ginContext)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.bulkDeleteCalls != 0 {
		t.Fatalf("expected no synchronous service call, got %d", service.bulkDeleteCalls)
	}
}

func versionedDeleteRequest(ids []uuid.UUID) dto.BulkDeleteFieldDeviceRequest {
	items := make([]dto.BulkDeleteFieldDeviceItem, len(ids))
	for index, id := range ids {
		items[index] = dto.BulkDeleteFieldDeviceItem{ID: id, BaseVersion: 1}
	}
	return dto.BulkDeleteFieldDeviceRequest{Items: items}
}

func newFieldDeviceIDs(count int) []uuid.UUID {
	ids := make([]uuid.UUID, count)
	for index := range ids {
		ids[index] = uuid.New()
	}
	return ids
}

func TestListFieldDevicesAcceptsMultiValueFilterParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	firstBuildingID := uuid.New()
	secondBuildingID := uuid.New()
	service := &fakeFieldDeviceHandlerService{}
	handler := NewFieldDeviceHandler(service)

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
	if service.listCursorCalls != 1 {
		t.Fatalf("expected cursor service to be called once, got %d call(s)", service.listCursorCalls)
	}
	if len(service.lastFilters.BuildingIDs) != 2 {
		t.Fatalf("expected two building ids, got %+v", service.lastFilters.BuildingIDs)
	}
	if service.lastFilters.BuildingIDs[0] != firstBuildingID || service.lastFilters.BuildingIDs[1] != secondBuildingID {
		t.Fatalf("unexpected building ids: %+v", service.lastFilters.BuildingIDs)
	}
}

func TestListFieldDevicesUsesLegacyPaginationOnlyWhenPageIsPresent(t *testing.T) {
	service := &fakeFieldDeviceHandlerService{}
	handler := NewFieldDeviceHandler(service)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodGet, "/field-devices?page=1&limit=20", nil)

	handler.ListFieldDevices(ginContext)

	if recorder.Code != http.StatusOK || service.listWithFiltersCalls != 1 || service.listCursorCalls != 0 {
		t.Fatalf("legacy response=%d legacy=%d cursor=%d", recorder.Code, service.listWithFiltersCalls, service.listCursorCalls)
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
	handler := NewFieldDeviceHandler(service)

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
	handler := NewFieldDeviceHandler(service)

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

type fieldDeviceStatusTrackingWriter struct {
	gin.ResponseWriter
	statusWrites []int
}

func (w *fieldDeviceStatusTrackingWriter) WriteHeader(code int) {
	w.statusWrites = append(w.statusWrites, code)
	w.ResponseWriter.WriteHeader(code)
}

type fakeFieldDeviceHandlerService struct {
	listWithFiltersCalls int
	listCursorCalls      int
	listAvailableCalls   int
	multiCreateCalls     int
	bulkDeleteCalls      int
	listWithFiltersErr   error
	listAvailableErr     error
	lastFilters          domainFacility.FieldDeviceFilterParams
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
	return nil, domain.ErrNotFound
}

func (s *fakeFieldDeviceHandlerService) CopyByID(context.Context, uuid.UUID) (*domainFacility.FieldDevice, error) {
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

func (s *fakeFieldDeviceHandlerService) ListCursor(_ context.Context, query domainFacility.FieldDeviceCursorQuery) (*domainFacility.FieldDeviceCursorPage, error) {
	s.listCursorCalls++
	s.lastFilters = query.Filters
	if s.listWithFiltersErr != nil {
		return nil, s.listWithFiltersErr
	}
	return &domainFacility.FieldDeviceCursorPage{Items: []domainFacility.FieldDevice{}}, nil
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
	return nil
}

func (s *fakeFieldDeviceHandlerService) DeleteByID(context.Context, uuid.UUID) error {
	return nil
}

func (s *fakeFieldDeviceHandlerService) Delete(context.Context, domainFacility.FieldDeviceDeleteCommand) error {
	return nil
}

func (s *fakeFieldDeviceHandlerService) ListBacnetObjects(context.Context, uuid.UUID) ([]domainFacility.BacnetObject, error) {
	return nil, nil
}

func (s *fakeFieldDeviceHandlerService) CreateSpecification(context.Context, uuid.UUID, *domainFacility.Specification) error {
	return nil
}

func (s *fakeFieldDeviceHandlerService) GetSpecification(context.Context, uuid.UUID) (*domainFacility.Specification, error) {
	return nil, domain.ErrNotFound
}

func (s *fakeFieldDeviceHandlerService) UpdateSpecificationPatch(context.Context, uuid.UUID, *domainFacility.SpecificationPatch) (*domainFacility.Specification, error) {
	return nil, nil
}

func (s *fakeFieldDeviceHandlerService) DeleteSpecification(context.Context, uuid.UUID) error {
	return nil
}

func (s *fakeFieldDeviceHandlerService) DeleteSpecificationAtVersion(context.Context, uuid.UUID, uint64) error {
	return nil
}

func (s *fakeFieldDeviceHandlerService) BulkUpdate(context.Context, []domainFacility.BulkFieldDeviceUpdate) *domainFacility.BulkOperationResult {
	return &domainFacility.BulkOperationResult{}
}

func (s *fakeFieldDeviceHandlerService) BulkDelete(context.Context, []uuid.UUID) *domainFacility.BulkOperationResult {
	s.bulkDeleteCalls++
	return &domainFacility.BulkOperationResult{}
}

func (s *fakeFieldDeviceHandlerService) BulkDeleteCommands(context.Context, []domainFacility.FieldDeviceDeleteCommand) *domainFacility.BulkOperationResult {
	s.bulkDeleteCalls++
	return &domainFacility.BulkOperationResult{}
}
