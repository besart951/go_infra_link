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
	intersectionID := uuid.New()
	desc := "Primary pump"

	apparatSvc := &fakeApparatHandlerService{
		listWithFiltersResult: &domain.PaginatedList[domainFacility.Apparat]{
			Items: []domainFacility.Apparat{
				{
					Base:        domain.Base{ID: intersectionID},
					ShortName:   "PMP",
					Name:        "Pump",
					Description: &desc,
				},
			},
			Total:      1,
			Page:       1,
			TotalPages: 1,
		},
	}

	handler := NewApparatHandler(apparatSvc)
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
	if apparatSvc.lastListParams.Search != "primary" || apparatSvc.lastListParams.Page != 1 || apparatSvc.lastListParams.Limit != 10 {
		t.Fatalf("expected pagination params to be forwarded, got %+v", apparatSvc.lastListParams)
	}
	if apparatSvc.lastListFilters.ObjectDataID == nil || *apparatSvc.lastListFilters.ObjectDataID != objectDataID {
		t.Fatalf("expected object_data_id %s, got %+v", objectDataID, apparatSvc.lastListFilters)
	}
	if apparatSvc.lastListFilters.SystemPartID == nil || *apparatSvc.lastListFilters.SystemPartID != systemPartID {
		t.Fatalf("expected system_part_id %s, got %+v", systemPartID, apparatSvc.lastListFilters)
	}
}

func TestApparatHandler_ListApparats_CharacterizesEmptyFilteredResult(t *testing.T) {
	gin.SetMode(gin.TestMode)

	objectDataID := uuid.New()
	systemPartID := uuid.New()
	handler := NewApparatHandler(&fakeApparatHandlerService{
		listWithFiltersResult: &domain.PaginatedList[domainFacility.Apparat]{
			Items:      []domainFacility.Apparat{},
			Total:      0,
			Page:       1,
			TotalPages: 0,
		},
	})
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
	items                 map[uuid.UUID]*domainFacility.Apparat
	listWithFiltersResult *domain.PaginatedList[domainFacility.Apparat]
	listWithFiltersErr    error
	lastListParams        domain.PaginationParams
	lastListFilters       domainFacility.ApparatFilterParams
}

func (s *fakeApparatHandlerService) Create(context.Context, *domainFacility.Apparat) error {
	return nil
}

func (s *fakeApparatHandlerService) CreateWithSystemPartIDs(context.Context, *domainFacility.Apparat, []uuid.UUID) error {
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
	out := make([]*domainFacility.Apparat, 0, len(ids))
	for _, id := range ids {
		if item, ok := s.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (s *fakeApparatHandlerService) List(_ context.Context, _ int, _ int, _ string) (*domain.PaginatedList[domainFacility.Apparat], error) {
	if s.listWithFiltersResult != nil {
		return s.listWithFiltersResult, s.listWithFiltersErr
	}
	items := make([]domainFacility.Apparat, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainFacility.Apparat]{Items: items, Total: int64(len(items)), Page: 1, TotalPages: 1}, nil
}

func (s *fakeApparatHandlerService) ListWithFilters(_ context.Context, params domain.PaginationParams, filters domainFacility.ApparatFilterParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	s.lastListParams = params
	s.lastListFilters = filters
	if s.listWithFiltersResult != nil || s.listWithFiltersErr != nil {
		return s.listWithFiltersResult, s.listWithFiltersErr
	}
	return &domain.PaginatedList[domainFacility.Apparat]{Items: []domainFacility.Apparat{}, Total: 0, Page: params.Page, TotalPages: 0}, nil
}

func (s *fakeApparatHandlerService) Update(context.Context, *domainFacility.Apparat) error {
	return nil
}

func (s *fakeApparatHandlerService) UpdateWithSystemPartIDs(context.Context, *domainFacility.Apparat, *[]uuid.UUID) error {
	return nil
}

func (s *fakeApparatHandlerService) DeleteByID(context.Context, uuid.UUID) error {
	return nil
}

func (s *fakeApparatHandlerService) GetSystemPartIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
