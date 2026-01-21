package facilitysql

import (
	"database/sql"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
	"github.com/google/uuid"
)

type projectControlCabinetRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewProjectControlCabinetRepository(db *sql.DB, driver string) domainFacility.ProjectControlCabinetStore {
	return &projectControlCabinetRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *projectControlCabinetRepo) Link(projectID uuid.UUID, controlCabinetID uuid.UUID) error {
	q := "INSERT INTO project_control_cabinets (project_id, control_cabinet_id) VALUES (?, ?) ON CONFLICT DO NOTHING"
	q = sqlutil.Rebind(r.dialect, q)
	_, err := r.db.Exec(q, projectID, controlCabinetID)
	return err
}

func (r *projectControlCabinetRepo) Unlink(projectID uuid.UUID, controlCabinetID uuid.UUID) error {
	q := "DELETE FROM project_control_cabinets WHERE project_id = ? AND control_cabinet_id = ?"
	q = sqlutil.Rebind(r.dialect, q)
	_, err := r.db.Exec(q, projectID, controlCabinetID)
	return err
}

func (r *projectControlCabinetRepo) GetProjectIDsByControlCabinet(controlCabinetID uuid.UUID) ([]uuid.UUID, error) {
	q := "SELECT project_id FROM project_control_cabinets WHERE control_cabinet_id = ?"
	q = sqlutil.Rebind(r.dialect, q)
	
	rows, err := r.db.Query(q, controlCabinetID)
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

func (r *projectControlCabinetRepo) GetControlCabinetIDsByProject(projectID uuid.UUID) ([]uuid.UUID, error) {
	q := "SELECT control_cabinet_id FROM project_control_cabinets WHERE project_id = ?"
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
