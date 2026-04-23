package projectsql

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestProjectControlCabinetRepo_ListAndDeleteByControlCabinetIDs(t *testing.T) {
	ctx := context.Background()
	projectOneID := uuid.New()
	projectTwoID := uuid.New()
	targetOneID := uuid.New()
	targetTwoID := uuid.New()
	keepID := uuid.New()

	repo := NewProjectControlCabinetRepository(newProjectLinkRepoTestDB(t))
	seedProjectControlCabinetLinks(t, ctx, repo,
		&domainProject.ProjectControlCabinet{ProjectID: projectOneID, ControlCabinetID: targetOneID},
		&domainProject.ProjectControlCabinet{ProjectID: projectTwoID, ControlCabinetID: targetTwoID},
		&domainProject.ProjectControlCabinet{ProjectID: projectOneID, ControlCabinetID: keepID},
	)

	items, err := repo.GetByControlCabinetIDs(ctx, []uuid.UUID{targetOneID, targetTwoID})
	if err != nil {
		t.Fatalf("expected list by control cabinet ids to succeed, got %v", err)
	}
	if len(items) != 2 || !sameUUIDSet(projectControlCabinetIDs(items), []uuid.UUID{targetOneID, targetTwoID}) {
		t.Fatalf("expected targeted control cabinet links, got %+v", items)
	}

	if err := repo.DeleteByControlCabinetIDs(ctx, []uuid.UUID{targetOneID, targetTwoID}); err != nil {
		t.Fatalf("expected delete by control cabinet ids to succeed, got %v", err)
	}

	remaining, err := repo.GetPaginatedList(ctx, domain.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("expected remaining link lookup to succeed, got %v", err)
	}
	if len(remaining.Items) != 1 || remaining.Items[0].ControlCabinetID != keepID {
		t.Fatalf("expected only unrelated control cabinet links to remain, got %+v", remaining.Items)
	}

	items, err = repo.GetByControlCabinetIDs(ctx, []uuid.UUID{targetOneID, targetTwoID})
	if err != nil {
		t.Fatalf("expected post-delete lookup to succeed, got %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected targeted control cabinet links to be removed, got %+v", items)
	}
}

func TestProjectSPSControllerRepo_ListAndDeleteBySPSControllerIDs(t *testing.T) {
	ctx := context.Background()
	projectOneID := uuid.New()
	projectTwoID := uuid.New()
	targetOneID := uuid.New()
	targetTwoID := uuid.New()
	keepID := uuid.New()

	repo := NewProjectSPSControllerRepository(newProjectLinkRepoTestDB(t))
	seedProjectSPSControllerLinks(t, ctx, repo,
		&domainProject.ProjectSPSController{ProjectID: projectOneID, SPSControllerID: targetOneID},
		&domainProject.ProjectSPSController{ProjectID: projectTwoID, SPSControllerID: targetTwoID},
		&domainProject.ProjectSPSController{ProjectID: projectOneID, SPSControllerID: keepID},
	)

	items, err := repo.GetBySPSControllerIDs(ctx, []uuid.UUID{targetOneID, targetTwoID})
	if err != nil {
		t.Fatalf("expected list by sps controller ids to succeed, got %v", err)
	}
	if len(items) != 2 || !sameUUIDSet(projectSPSControllerIDs(items), []uuid.UUID{targetOneID, targetTwoID}) {
		t.Fatalf("expected targeted sps controller links, got %+v", items)
	}

	if err := repo.DeleteBySPSControllerIDs(ctx, []uuid.UUID{targetOneID, targetTwoID}); err != nil {
		t.Fatalf("expected delete by sps controller ids to succeed, got %v", err)
	}

	remaining, err := repo.GetPaginatedList(ctx, domain.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("expected remaining link lookup to succeed, got %v", err)
	}
	if len(remaining.Items) != 1 || remaining.Items[0].SPSControllerID != keepID {
		t.Fatalf("expected only unrelated sps controller links to remain, got %+v", remaining.Items)
	}

	items, err = repo.GetBySPSControllerIDs(ctx, []uuid.UUID{targetOneID, targetTwoID})
	if err != nil {
		t.Fatalf("expected post-delete lookup to succeed, got %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected targeted sps controller links to be removed, got %+v", items)
	}
}

func TestProjectFieldDeviceRepo_ListAndDeleteByFieldDeviceIDs(t *testing.T) {
	ctx := context.Background()
	projectOneID := uuid.New()
	projectTwoID := uuid.New()
	targetOneID := uuid.New()
	targetTwoID := uuid.New()
	keepID := uuid.New()

	repo := NewProjectFieldDeviceRepository(newProjectLinkRepoTestDB(t))
	seedProjectFieldDeviceLinks(t, ctx, repo,
		&domainProject.ProjectFieldDevice{ProjectID: projectOneID, FieldDeviceID: targetOneID},
		&domainProject.ProjectFieldDevice{ProjectID: projectTwoID, FieldDeviceID: targetTwoID},
		&domainProject.ProjectFieldDevice{ProjectID: projectOneID, FieldDeviceID: keepID},
	)

	items, err := repo.GetByFieldDeviceIDs(ctx, []uuid.UUID{targetOneID, targetTwoID})
	if err != nil {
		t.Fatalf("expected list by field device ids to succeed, got %v", err)
	}
	if len(items) != 2 || !sameUUIDSet(projectFieldDeviceIDs(items), []uuid.UUID{targetOneID, targetTwoID}) {
		t.Fatalf("expected targeted field device links, got %+v", items)
	}

	if err := repo.DeleteByFieldDeviceIDs(ctx, []uuid.UUID{targetOneID, targetTwoID}); err != nil {
		t.Fatalf("expected delete by field device ids to succeed, got %v", err)
	}

	remaining, err := repo.GetPaginatedList(ctx, domain.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("expected remaining link lookup to succeed, got %v", err)
	}
	if len(remaining.Items) != 1 || remaining.Items[0].FieldDeviceID != keepID {
		t.Fatalf("expected only unrelated field device links to remain, got %+v", remaining.Items)
	}

	items, err = repo.GetByFieldDeviceIDs(ctx, []uuid.UUID{targetOneID, targetTwoID})
	if err != nil {
		t.Fatalf("expected post-delete lookup to succeed, got %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected targeted field device links to be removed, got %+v", items)
	}
}

func newProjectLinkRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.NewReplacer("/", "_", " ", "_", "#", "_").Replace(t.Name()))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("expected sqlite db to open, got %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("expected sql db handle, got %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(
		&domainProject.ProjectControlCabinet{},
		&domainProject.ProjectSPSController{},
		&domainProject.ProjectFieldDevice{},
	); err != nil {
		t.Fatalf("expected project link tables to migrate, got %v", err)
	}

	return db
}

func seedProjectControlCabinetLinks(t *testing.T, ctx context.Context, repo domainProject.ProjectControlCabinetRepository, items ...*domainProject.ProjectControlCabinet) {
	t.Helper()
	for _, item := range items {
		if err := repo.Create(ctx, item); err != nil {
			t.Fatalf("expected project control cabinet seed to succeed, got %v", err)
		}
	}
}

func seedProjectSPSControllerLinks(t *testing.T, ctx context.Context, repo domainProject.ProjectSPSControllerRepository, items ...*domainProject.ProjectSPSController) {
	t.Helper()
	for _, item := range items {
		if err := repo.Create(ctx, item); err != nil {
			t.Fatalf("expected project sps controller seed to succeed, got %v", err)
		}
	}
}

func seedProjectFieldDeviceLinks(t *testing.T, ctx context.Context, repo domainProject.ProjectFieldDeviceRepository, items ...*domainProject.ProjectFieldDevice) {
	t.Helper()
	for _, item := range items {
		if err := repo.Create(ctx, item); err != nil {
			t.Fatalf("expected project field device seed to succeed, got %v", err)
		}
	}
}

func projectControlCabinetIDs(items []*domainProject.ProjectControlCabinet) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		out = append(out, item.ControlCabinetID)
	}
	return out
}

func projectSPSControllerIDs(items []*domainProject.ProjectSPSController) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		out = append(out, item.SPSControllerID)
	}
	return out
}

func projectFieldDeviceIDs(items []*domainProject.ProjectFieldDevice) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		out = append(out, item.FieldDeviceID)
	}
	return out
}

func sameUUIDSet(got, want []uuid.UUID) bool {
	if len(got) != len(want) {
		return false
	}

	counts := make(map[uuid.UUID]int, len(got))
	for _, id := range got {
		counts[id]++
	}
	for _, id := range want {
		counts[id]--
		if counts[id] < 0 {
			return false
		}
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}
