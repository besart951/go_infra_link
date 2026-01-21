package facilitysql

import (
	"database/sql"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
	"github.com/google/uuid"
)

type objectDataRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewObjectDataRepository(db *sql.DB, driver string) domainFacility.ObjectDataStore {
	return &objectDataRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *objectDataRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.ObjectData, error) {
	if len(ids) == 0 {
		return []*domainFacility.ObjectData{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, description, version, is_active, project_id FROM object_data WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]*domainFacility.ObjectData, 0, len(ids))
	for rows.Next() {
		var o domainFacility.ObjectData
		var deletedAt sql.NullTime
		var projectID sql.NullString
		if err := rows.Scan(
			&o.ID,
			&o.CreatedAt,
			&o.UpdatedAt,
			&deletedAt,
			&o.Description,
			&o.Version,
			&o.IsActive,
			&projectID,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			o.DeletedAt = &t
		}
		if projectID.Valid {
			id, err := uuid.Parse(projectID.String)
			if err != nil {
				return nil, err
			}
			o.ProjectID = &id
		}
		out = append(out, &o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *objectDataRepo) Create(entity *domainFacility.ObjectData) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO object_data (id, created_at, updated_at, deleted_at, description, version, is_active, project_id) VALUES (?, ?, ?, NULL, ?, ?, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(q, entity.ID, entity.CreatedAt, entity.UpdatedAt, entity.Description, entity.Version, entity.IsActive, argUUIDPtr(entity.ProjectID))
	return err
}

func (r *objectDataRepo) Update(entity *domainFacility.ObjectData) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	q := "UPDATE object_data SET updated_at = ?, description = ?, version = ?, is_active = ?, project_id = ? WHERE deleted_at IS NULL AND id = ?"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(q, entity.UpdatedAt, entity.Description, entity.Version, entity.IsActive, argUUIDPtr(entity.ProjectID), entity.ID)
	return err
}

func (r *objectDataRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE object_data SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := r.db.Exec(q, args...)
	return err
}

func (r *objectDataRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page := params.Page
	limit := params.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	where := "deleted_at IS NULL"
	args := make([]any, 0, 6)
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + params.Search + "%"
		where += " AND (description " + sqlutil.LikeOperator(r.dialect) + " ?)"
		args = append(args, pattern)
	}

	countQ := "SELECT COUNT(*) FROM object_data WHERE " + where
	countQ = sqlutil.Rebind(r.dialect, countQ)
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, description, version, is_active, project_id FROM object_data WHERE " + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	dataQ = sqlutil.Rebind(r.dialect, dataQ)
	dataArgs := append(append([]any{}, args...), limit, offset)

	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainFacility.ObjectData, 0, limit)
	for rows.Next() {
		var o domainFacility.ObjectData
		var deletedAt sql.NullTime
		var projectID sql.NullString
		if err := rows.Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt, &deletedAt, &o.Description, &o.Version, &o.IsActive, &projectID); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			o.DeletedAt = &t
		}
		if projectID.Valid {
			id, err := uuid.Parse(projectID.String)
			if err != nil {
				return nil, err
			}
			o.ProjectID = &id
		}
		items = append(items, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.ObjectData]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *objectDataRepo) GetBacnetObjectIDs(objectDataID uuid.UUID) ([]uuid.UUID, error) {
	q := "SELECT bacnet_object_id FROM object_data_bacnet_objects WHERE object_data_id = ?"
	q = sqlutil.Rebind(r.dialect, q)

	rows, err := r.db.Query(q, objectDataID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	ids := make([]uuid.UUID, 0, 16)
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
