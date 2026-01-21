package facilitysql

import (
	"database/sql"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
	"github.com/google/uuid"
)

type objectDataBacnetObjectRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewObjectDataBacnetObjectRepository(db *sql.DB, driver string) domainFacility.ObjectDataBacnetObjectStore {
	return &objectDataBacnetObjectRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *objectDataBacnetObjectRepo) Add(objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error {
	q := "INSERT INTO object_data_bacnet_objects (object_data_id, bacnet_object_id) VALUES (?, ?)"
	q = sqlutil.Rebind(r.dialect, q)
	_, err := r.db.Exec(q, objectDataID, bacnetObjectID)
	return err
}

func (r *objectDataBacnetObjectRepo) Delete(objectDataID uuid.UUID, bacnetObjectID uuid.UUID) error {
	q := "DELETE FROM object_data_bacnet_objects WHERE object_data_id = ? AND bacnet_object_id = ?"
	q = sqlutil.Rebind(r.dialect, q)
	_, err := r.db.Exec(q, objectDataID, bacnetObjectID)
	return err
}
