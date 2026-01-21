package facilitysql

import (
	"database/sql"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
	"github.com/google/uuid"
)

type projectSPSControllerRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewProjectSPSControllerRepository(db *sql.DB, driver string) domainFacility.ProjectSPSControllerStore {
	return &projectSPSControllerRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *projectSPSControllerRepo) Link(projectID uuid.UUID, spsControllerID uuid.UUID) error {
	q := "INSERT INTO project_sps_controllers (project_id, sps_controller_id) VALUES (?, ?) ON CONFLICT DO NOTHING"
	q = sqlutil.Rebind(r.dialect, q)
	_, err := r.db.Exec(q, projectID, spsControllerID)
	return err
}

func (r *projectSPSControllerRepo) Unlink(projectID uuid.UUID, spsControllerID uuid.UUID) error {
	q := "DELETE FROM project_sps_controllers WHERE project_id = ? AND sps_controller_id = ?"
	q = sqlutil.Rebind(r.dialect, q)
	_, err := r.db.Exec(q, projectID, spsControllerID)
	return err
}

func (r *projectSPSControllerRepo) GetProjectIDsBySPSController(spsControllerID uuid.UUID) ([]uuid.UUID, error) {
	q := "SELECT project_id FROM project_sps_controllers WHERE sps_controller_id = ?"
	q = sqlutil.Rebind(r.dialect, q)
	
	rows, err := r.db.Query(q, spsControllerID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return ids, nil
}

func (r *projectSPSControllerRepo) GetSPSControllerIDsByProject(projectID uuid.UUID) ([]uuid.UUID, error) {
	q := "SELECT sps_controller_id FROM project_sps_controllers WHERE project_id = ?"
	q = sqlutil.Rebind(r.dialect, q)
	
	rows, err := r.db.Query(q, projectID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return ids, nil
}
