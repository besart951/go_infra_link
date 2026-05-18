package project

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestProjectRepo_MembershipAndUserScopedListing(t *testing.T) {
	ctx := context.Background()
	db := newProjectRepoTestDB(t)
	repo := NewProjectRepository(db)

	creator := seedProjectRepoUser(t, db, "creator@example.com")
	member := seedProjectRepoUser(t, db, "member@example.com")
	executionPhase := seedProjectRepoPhase(t, db, "Ausführung")
	planningPhase := seedProjectRepoPhase(t, db, "Planung")

	ongoing := &domainProject.Project{
		Name:        "Ongoing Migration",
		Description: "mapped via ProjectRecord",
		Status:      domainProject.StatusOngoing,
		PhaseID:     executionPhase.ID,
		CreatorID:   creator.ID,
	}
	if err := repo.Create(ctx, ongoing); err != nil {
		t.Fatalf("expected project create to succeed, got %v", err)
	}
	if ongoing.ID == uuid.Nil {
		t.Fatal("expected created project id to be assigned")
	}

	planned := &domainProject.Project{
		Name:      "Planned Migration",
		Status:    domainProject.StatusPlanned,
		PhaseID:   planningPhase.ID,
		CreatorID: creator.ID,
	}
	if err := repo.Create(ctx, planned); err != nil {
		t.Fatalf("expected second project create to succeed, got %v", err)
	}

	if err := repo.AddUser(ctx, ongoing.ID, creator.ID); err != nil {
		t.Fatalf("expected creator membership to be added, got %v", err)
	}
	if err := repo.AddUser(ctx, ongoing.ID, member.ID); err != nil {
		t.Fatalf("expected member membership to be added, got %v", err)
	}
	if err := repo.AddUser(ctx, planned.ID, member.ID); err != nil {
		t.Fatalf("expected planned project membership to be added, got %v", err)
	}

	items, err := repo.GetByIds(ctx, []uuid.UUID{ongoing.ID})
	if err != nil {
		t.Fatalf("expected project lookup to succeed, got %v", err)
	}
	if len(items) != 1 || items[0].Name != ongoing.Name || items[0].CreatorID != creator.ID || items[0].Status != domainProject.StatusOngoing {
		t.Fatalf("expected mapped project fields, got %+v", items)
	}
	if items[0].Phase == nil || items[0].Phase.Name != executionPhase.Name {
		t.Fatalf("expected phase to be attached, got %+v", items[0].Phase)
	}

	hasUser, err := repo.HasUser(ctx, ongoing.ID, member.ID)
	if err != nil {
		t.Fatalf("expected membership lookup to succeed, got %v", err)
	}
	if !hasUser {
		t.Fatal("expected added member to be present on the project")
	}

	users, err := repo.ListUsers(ctx, ongoing.ID)
	if err != nil {
		t.Fatalf("expected project users to load, got %v", err)
	}
	if len(users) != 2 || !sameUserSet(users, []uuid.UUID{creator.ID, member.ID}) {
		t.Fatalf("expected creator and member to be listed, got %+v", users)
	}

	status := domainProject.StatusOngoing
	list, err := repo.GetPaginatedListForUserWithStatus(ctx, domain.PaginationParams{Page: 1, Limit: 10}, member.ID, &status, nil)
	if err != nil {
		t.Fatalf("expected user-scoped listing to succeed, got %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].ID != ongoing.ID {
		t.Fatalf("expected only the ongoing project for the member, got %+v", list.Items)
	}
	if list.Items[0].Phase == nil || list.Items[0].Phase.Name != executionPhase.Name {
		t.Fatalf("expected listed project phase to be attached, got %+v", list.Items[0].Phase)
	}

	phaseFiltered, err := repo.GetPaginatedListWithStatus(ctx, domain.PaginationParams{Page: 1, Limit: 10}, nil, &planningPhase.ID)
	if err != nil {
		t.Fatalf("expected phase-filtered listing to succeed, got %v", err)
	}
	if len(phaseFiltered.Items) != 1 || phaseFiltered.Items[0].ID != planned.ID {
		t.Fatalf("expected only the planning project, got %+v", phaseFiltered.Items)
	}

	phaseSearch, err := repo.GetPaginatedListForUserWithStatus(ctx, domain.PaginationParams{Page: 1, Limit: 10, Search: "Ausführung"}, member.ID, nil, nil)
	if err != nil {
		t.Fatalf("expected phase-name search to succeed, got %v", err)
	}
	if len(phaseSearch.Items) != 1 || phaseSearch.Items[0].ID != ongoing.ID {
		t.Fatalf("expected phase-name search to find ongoing project, got %+v", phaseSearch.Items)
	}

	if err := repo.RemoveUser(ctx, ongoing.ID, member.ID); err != nil {
		t.Fatalf("expected member removal to succeed, got %v", err)
	}

	hasUser, err = repo.HasUser(ctx, ongoing.ID, member.ID)
	if err != nil {
		t.Fatalf("expected membership lookup after removal to succeed, got %v", err)
	}
	if hasUser {
		t.Fatal("expected member to be removed from the project")
	}
}

func newProjectRepoTestDB(t *testing.T) *gorm.DB {
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

	if err := db.AutoMigrate(&domainProject.Phase{}, &ProjectRecord{}, &ProjectUserRecord{}, &domainUser.User{}); err != nil {
		t.Fatalf("expected project repo tables to migrate, got %v", err)
	}

	return db
}

func seedProjectRepoUser(t *testing.T, db *gorm.DB, email string) *domainUser.User {
	t.Helper()

	user := &domainUser.User{
		FirstName: "Repo",
		LastName:  "Tester",
		Email:     domainUser.EmailPtr(email),
		Password:  "password",
	}
	if err := user.Base.InitForCreate(time.Now().UTC()); err != nil {
		t.Fatalf("expected user seed id init to succeed, got %v", err)
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("expected user seed to succeed, got %v", err)
	}
	return user
}

func seedProjectRepoPhase(t *testing.T, db *gorm.DB, name string) *domainProject.Phase {
	t.Helper()

	phase := &domainProject.Phase{Name: name}
	if err := phase.Base.InitForCreate(time.Now().UTC()); err != nil {
		t.Fatalf("expected phase seed id init to succeed, got %v", err)
	}
	if err := db.Create(phase).Error; err != nil {
		t.Fatalf("expected phase seed to succeed, got %v", err)
	}
	return phase
}

func sameUserSet(users []domainUser.User, want []uuid.UUID) bool {
	if len(users) != len(want) {
		return false
	}

	counts := make(map[uuid.UUID]int, len(users))
	for _, user := range users {
		counts[user.ID]++
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
