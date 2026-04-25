package facility

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	dto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	listAvailableCalls   int
}

func (s *fakeFieldDeviceHandlerService) Create(context.Context, *domainFacility.FieldDevice) error {
	return nil
}

func (s *fakeFieldDeviceHandlerService) CreateWithBacnetObjects(context.Context, *domainFacility.FieldDevice, *uuid.UUID, []domainFacility.BacnetObject) error {
	return nil
}

func (s *fakeFieldDeviceHandlerService) MultiCreate(context.Context, []domainFacility.FieldDeviceCreateItem) *domainFacility.FieldDeviceMultiCreateResult {
	return &domainFacility.FieldDeviceMultiCreateResult{}
}

func (s *fakeFieldDeviceHandlerService) GetByID(context.Context, uuid.UUID) (*domainFacility.FieldDevice, error) {
	return nil, domain.ErrNotFound
}

func (s *fakeFieldDeviceHandlerService) List(context.Context, int, int, string) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	return &domain.PaginatedList[domainFacility.FieldDevice]{}, nil
}

func (s *fakeFieldDeviceHandlerService) ListWithFilters(context.Context, domain.PaginationParams, domainFacility.FieldDeviceFilterParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	s.listWithFiltersCalls++
	return &domain.PaginatedList[domainFacility.FieldDevice]{
		Items:      []domainFacility.FieldDevice{},
		Page:       1,
		TotalPages: 1,
	}, nil
}

func (s *fakeFieldDeviceHandlerService) ListAvailableApparatNumbers(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]int, error) {
	s.listAvailableCalls++
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
	return &domainFacility.BulkOperationResult{}
}

func (s *fakeFieldDeviceHandlerService) BulkDelete(context.Context, []uuid.UUID) *domainFacility.BulkOperationResult {
	return &domainFacility.BulkOperationResult{}
}
