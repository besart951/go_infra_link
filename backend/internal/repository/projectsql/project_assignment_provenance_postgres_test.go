package projectsql

import (
	"context"
	"os"
	"testing"

	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestProjectAssignmentSourcesRetainExplicitAndInheritedClaims(
	t *testing.T,
) {
	dsn := os.Getenv("FACILITY_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("FACILITY_TEST_DATABASE_URL is not set")
	}
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	tx := database.Begin()
	if tx.Error != nil {
		t.Fatalf("begin test transaction: %v", tx.Error)
	}
	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})

	projectID := uuid.New()
	controllerID := uuid.New()
	fieldDeviceID := uuid.New()
	spsLinkID := uuid.New()
	fieldLinkID := uuid.New()
	if err := tx.Exec(`SET LOCAL session_replication_role = 'replica'`).Error; err != nil {
		t.Fatalf("disable fixture constraints: %v", err)
	}
	if err := tx.Exec(`
		INSERT INTO project_sps_controllers (
			id, project_id, sps_controller_id
		) VALUES (?, ?, ?)
	`, spsLinkID, projectID, controllerID).Error; err != nil {
		t.Fatalf("seed SPS project link: %v", err)
	}
	if err := tx.Exec(`
		INSERT INTO project_field_devices (
			id, project_id, field_device_id
		) VALUES (?, ?, ?)
	`, fieldLinkID, projectID, fieldDeviceID).Error; err != nil {
		t.Fatalf("seed FieldDevice project link: %v", err)
	}
	if err := tx.Exec(`SET LOCAL session_replication_role = 'origin'`).Error; err != nil {
		t.Fatalf("restore fixture constraints: %v", err)
	}

	ctx := context.Background()
	spsRepo := &projectSPSControllerRepo{db: tx}
	fieldRepo := &projectFieldDeviceRepo{db: tx}
	cabinetSourceID := uuid.New()
	if err := spsRepo.AddAssignmentSource(
		ctx,
		projectID,
		[]uuid.UUID{controllerID},
		domainProject.ExplicitAssignmentSource(),
	); err != nil {
		t.Fatalf("add explicit SPS source: %v", err)
	}
	if err := spsRepo.AddAssignmentSource(
		ctx,
		projectID,
		[]uuid.UUID{controllerID},
		domainProject.AssignmentSource{
			Kind:           domainProject.AssignmentSourceControlCabinet,
			SourceEntityID: cabinetSourceID,
		},
	); err != nil {
		t.Fatalf("add inherited SPS source: %v", err)
	}
	if err := fieldRepo.AddAssignmentSource(
		ctx,
		projectID,
		[]uuid.UUID{fieldDeviceID},
		domainProject.ExplicitAssignmentSource(),
	); err != nil {
		t.Fatalf("add explicit FieldDevice source: %v", err)
	}
	if err := fieldRepo.AddAssignmentSource(
		ctx,
		projectID,
		[]uuid.UUID{fieldDeviceID},
		domainProject.AssignmentSource{
			Kind:           domainProject.AssignmentSourceSPSController,
			SourceEntityID: controllerID,
		},
	); err != nil {
		t.Fatalf("add inherited FieldDevice source: %v", err)
	}

	hasOther, err := spsRepo.HasAssignmentSourceOtherThan(
		ctx,
		spsLinkID,
		domainProject.AssignmentSource{
			Kind:           domainProject.AssignmentSourceExplicit,
			SourceEntityID: controllerID,
		},
	)
	if err != nil {
		t.Fatalf("inspect SPS sources: %v", err)
	}
	if !hasOther {
		t.Fatal("expected inherited SPS source in addition to explicit source")
	}

	assertProjectAssignmentSourceCount(
		t,
		tx,
		"project_sps_controller_assignment_sources",
		"project_sps_controller_id",
		spsLinkID,
		2,
	)
	assertProjectAssignmentSourceCount(
		t,
		tx,
		"project_field_device_assignment_sources",
		"project_field_device_id",
		fieldLinkID,
		2,
	)

	explicitProjects, err := spsRepo.ListProjectIDsByAssignmentSource(
		ctx,
		domainProject.AssignmentSource{
			Kind:           domainProject.AssignmentSourceExplicit,
			SourceEntityID: controllerID,
		},
	)
	if err != nil {
		t.Fatalf("list explicit SPS projects: %v", err)
	}
	if len(explicitProjects) != 1 || explicitProjects[0] != projectID {
		t.Fatalf("explicit SPS projects: got %v, want [%s]", explicitProjects, projectID)
	}

	processed, unclaimed, err := spsRepo.RemoveAssignmentSourceBatch(
		ctx,
		projectID,
		domainProject.AssignmentSource{
			Kind:           domainProject.AssignmentSourceExplicit,
			SourceEntityID: controllerID,
		},
		uuid.Nil,
		100,
	)
	if err != nil {
		t.Fatalf("remove explicit SPS source: %v", err)
	}
	if len(processed) != 1 || processed[0] != spsLinkID || len(unclaimed) != 0 {
		t.Fatalf(
			"explicit source removal: processed=%v unclaimed=%v",
			processed,
			unclaimed,
		)
	}

	processed, unclaimed, err = spsRepo.RemoveAssignmentSourceBatch(
		ctx,
		projectID,
		domainProject.AssignmentSource{
			Kind:           domainProject.AssignmentSourceControlCabinet,
			SourceEntityID: cabinetSourceID,
		},
		uuid.Nil,
		100,
	)
	if err != nil {
		t.Fatalf("remove inherited SPS source: %v", err)
	}
	if len(processed) != 1 || processed[0] != spsLinkID ||
		len(unclaimed) != 1 || unclaimed[0] != spsLinkID {
		t.Fatalf(
			"last source removal: processed=%v unclaimed=%v",
			processed,
			unclaimed,
		)
	}
}

func assertProjectAssignmentSourceCount(
	t *testing.T,
	db *gorm.DB,
	table string,
	column string,
	linkID uuid.UUID,
	want int64,
) {
	t.Helper()
	var count int64
	if err := db.Table(table).Where(column+" = ?", linkID).Count(&count).Error; err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if count != want {
		t.Fatalf("%s source count: got %d, want %d", table, count, want)
	}
}
