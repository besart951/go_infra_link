package facilitysql

import (
	"database/sql"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
	"github.com/google/uuid"
)

type projectFieldDeviceRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewProjectFieldDeviceRepository(db *sql.DB, driver string) domainFacility.ProjectFieldDeviceStore {
	return &projectFieldDeviceRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *projectFieldDeviceRepo) Link(projectID uuid.UUID, fieldDeviceID uuid.UUID) error {
	q := "INSERT INTO project_field_devices (project_id, field_device_id) VALUES (?, ?) ON CONFLICT DO NOTHING"
	q = sqlutil.Rebind(r.dialect, q)
	_, err := r.db.Exec(q, projectID, fieldDeviceID)
	return err
}

func (r *projectFieldDeviceRepo) Unlink(projectID uuid.UUID, fieldDeviceID uuid.UUID) error {
	q := "DELETE FROM project_field_devices WHERE project_id = ? AND field_device_id = ?"
	q = sqlutil.Rebind(r.dialect, q)
	_, err := r.db.Exec(q, projectID, fieldDeviceID)
	return err
}

func (r *projectFieldDeviceRepo) GetProjectIDsByFieldDevice(fieldDeviceID uuid.UUID) ([]uuid.UUID, error) {
	q := "SELECT project_id FROM project_field_devices WHERE field_device_id = ?"
	q = sqlutil.Rebind(r.dialect, q)
	
	rows, err := r.db.Query(q, fieldDeviceID)
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

func (r *projectFieldDeviceRepo) GetFieldDeviceIDsByProject(projectID uuid.UUID) ([]uuid.UUID, error) {
	q := "SELECT field_device_id FROM project_field_devices WHERE project_id = ?"
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
