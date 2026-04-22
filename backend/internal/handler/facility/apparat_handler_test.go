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

func TestApparatHandler_ListApparats_CharacterizesCombinedFiltersAndSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	objectDataID := uuid.New()
	systemPartID := uuid.New()
	objectOnlyID := uuid.New()
	intersectionID := uuid.New()
	systemPartOnlyID := uuid.New()
	desc := "Primary pump"

	apparatSvc := &fakeApparatHandlerService{
		items: map[uuid.UUID]*domainFacility.Apparat{
			objectOnlyID:     {Base: domain.Base{ID: objectOnlyID}, ShortName: "OBJ", Name: "Object only"},
			intersectionID:   {Base: domain.Base{ID: intersectionID}, ShortName: "PMP", Name: "Pump", Description: &desc},
			systemPartOnlyID: {Base: domain.Base{ID: systemPartOnlyID}, ShortName: "SYS", Name: "System part only"},
		},
	}
	systemPartSvc := &fakeSystemPartHandlerService{
		apparatIDsBySystemPart: map[uuid.UUID][]uuid.UUID{
			systemPartID: {intersectionID, systemPartOnlyID},
		},
	}
	objectDataSvc := &fakeObjectDataHandlerService{
		apparatIDsByObjectData: map[uuid.UUID][]uuid.UUID{
			objectDataID: {objectOnlyID, intersectionID},
		},
	}

	handler := NewApparatHandler(apparatSvc, systemPartSvc, objectDataSvc)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodGet,
		"/apparats?object_data_id="+objectDataID.String()+"&system_part_id="+systemPartID.String()+"&search=primary&page=1&limit=10",
		nil,
	)

	handler.ListApparats(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response dto.ApparatListResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected response json to decode, got %v", err)
	}
	if response.Total != 1 || response.Page != 1 || response.TotalPages != 1 {
		t.Fatalf("expected one filtered result on page 1, got %+v", response)
	}
	if len(response.Items) != 1 || response.Items[0].ID != intersectionID {
		t.Fatalf("expected only intersected and search-matching apparat %s, got %+v", intersectionID, response.Items)
	}
	if !sameUUIDSequence(apparatSvc.lastGetByIDs, []uuid.UUID{intersectionID}) {
		t.Fatalf("expected handler to fetch only intersected apparat id, got %v", apparatSvc.lastGetByIDs)
	}
}

func TestApparatHandler_ListApparats_CharacterizesEmptyFilteredResult(t *testing.T) {
	gin.SetMode(gin.TestMode)

	objectDataID := uuid.New()
	systemPartID := uuid.New()
	handler := NewApparatHandler(
		&fakeApparatHandlerService{items: map[uuid.UUID]*domainFacility.Apparat{}},
		&fakeSystemPartHandlerService{apparatIDsBySystemPart: map[uuid.UUID][]uuid.UUID{systemPartID: {uuid.New()}}},
		&fakeObjectDataHandlerService{apparatIDsByObjectData: map[uuid.UUID][]uuid.UUID{objectDataID: {uuid.New()}}},
	)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodGet,
		"/apparats?object_data_id="+objectDataID.String()+"&system_part_id="+systemPartID.String(),
		nil,
	)

	handler.ListApparats(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response dto.ApparatListResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected response json to decode, got %v", err)
	}
	if len(response.Items) != 0 || response.Total != 0 || response.Page != 1 || response.TotalPages != 0 {
		t.Fatalf("expected current empty-filter response shape, got %+v", response)
	}
}

type fakeApparatHandlerService struct {
	items        map[uuid.UUID]*domainFacility.Apparat
	lastGetByIDs []uuid.UUID
}

func (s *fakeApparatHandlerService) Create(context.Context, *domainFacility.Apparat) error {
	return nil
}

func (s *fakeApparatHandlerService) GetByID(_ context.Context, id uuid.UUID) (*domainFacility.Apparat, error) {
	items, err := s.GetByIDs(context.Background(), []uuid.UUID{id})
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return items[0], nil
}

func (s *fakeApparatHandlerService) GetByIDs(_ context.Context, ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	s.lastGetByIDs = append([]uuid.UUID(nil), ids...)
	out := make([]*domainFacility.Apparat, 0, len(ids))
	for _, id := range ids {
		if item, ok := s.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (s *fakeApparatHandlerService) List(context.Context, int, int, string) (*domain.PaginatedList[domainFacility.Apparat], error) {
	items := make([]domainFacility.Apparat, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainFacility.Apparat]{Items: items, Total: int64(len(items)), Page: 1, TotalPages: 1}, nil
}

func (s *fakeApparatHandlerService) Update(context.Context, *domainFacility.Apparat) error {
	return nil
}

func (s *fakeApparatHandlerService) DeleteByID(context.Context, uuid.UUID) error {
	return nil
}

func (s *fakeApparatHandlerService) GetSystemPartIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

type fakeSystemPartHandlerService struct {
	apparatIDsBySystemPart map[uuid.UUID][]uuid.UUID
}

func (s *fakeSystemPartHandlerService) Create(context.Context, *domainFacility.SystemPart) error {
	return nil
}

func (s *fakeSystemPartHandlerService) GetByID(context.Context, uuid.UUID) (*domainFacility.SystemPart, error) {
	return nil, nil
}

func (s *fakeSystemPartHandlerService) GetByIDs(context.Context, []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	return nil, nil
}

func (s *fakeSystemPartHandlerService) GetApparatIDs(_ context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	return append([]uuid.UUID(nil), s.apparatIDsBySystemPart[id]...), nil
}

func (s *fakeSystemPartHandlerService) List(context.Context, int, int, string) (*domain.PaginatedList[domainFacility.SystemPart], error) {
	return nil, nil
}

func (s *fakeSystemPartHandlerService) Update(context.Context, *domainFacility.SystemPart) error {
	return nil
}

func (s *fakeSystemPartHandlerService) DeleteByID(context.Context, uuid.UUID) error {
	return nil
}

type fakeObjectDataHandlerService struct {
	apparatIDsByObjectData map[uuid.UUID][]uuid.UUID
}

func (s *fakeObjectDataHandlerService) Create(context.Context, *domainFacility.ObjectData) error {
	return nil
}

func (s *fakeObjectDataHandlerService) GetByID(context.Context, uuid.UUID) (*domainFacility.ObjectData, error) {
	return nil, nil
}

func (s *fakeObjectDataHandlerService) List(context.Context, int, int, string) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return nil, nil
}

func (s *fakeObjectDataHandlerService) ListByApparatID(context.Context, int, int, string, uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return nil, nil
}

func (s *fakeObjectDataHandlerService) ListBySystemPartID(context.Context, int, int, string, uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return nil, nil
}

func (s *fakeObjectDataHandlerService) ListByApparatAndSystemPartID(context.Context, int, int, string, uuid.UUID, uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	return nil, nil
}

func (s *fakeObjectDataHandlerService) Update(context.Context, *domainFacility.ObjectData) error {
	return nil
}

func (s *fakeObjectDataHandlerService) DeleteByID(context.Context, uuid.UUID) error {
	return nil
}

func (s *fakeObjectDataHandlerService) GetBacnetObjectIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (s *fakeObjectDataHandlerService) GetApparatIDs(_ context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	return append([]uuid.UUID(nil), s.apparatIDsByObjectData[id]...), nil
}

func (s *fakeObjectDataHandlerService) ExistsByDescription(context.Context, *uuid.UUID, string, *uuid.UUID) (bool, error) {
	return false, nil
}

func sameUUIDSequence(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
