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

type fieldDeviceRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewFieldDeviceRepository(db *sql.DB, driver string) domainFacility.FieldDeviceRepository {
	return &fieldDeviceRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *fieldDeviceRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.FieldDevice, error) {
	if len(ids) == 0 {
		return []*domainFacility.FieldDevice{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, bmk, description, apparat_nr, sps_controller_system_type_id, system_part_id, specification_id, project_id, apparat_id " +
		"FROM field_devices WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
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

	out := make([]*domainFacility.FieldDevice, 0, len(ids))
	for rows.Next() {
		var fd domainFacility.FieldDevice
		var deletedAt sql.NullTime
		var bmk sql.NullString
		var desc sql.NullString
		var apparatNr sql.NullInt64
		var systemPartID sql.NullString
		var specificationID sql.NullString
		var projectID sql.NullString

		if err := rows.Scan(
			&fd.ID,
			&fd.CreatedAt,
			&fd.UpdatedAt,
			&deletedAt,
			&bmk,
			&desc,
			&apparatNr,
			&fd.SPSControllerSystemTypeID,
			&systemPartID,
			&specificationID,
			&projectID,
			&fd.ApparatID,
		); err != nil {
			return nil, err
		}

		if deletedAt.Valid {
			t := deletedAt.Time
			fd.DeletedAt = &t
		}
		if bmk.Valid {
			v := bmk.String
			fd.BMK = &v
		}
		if desc.Valid {
			v := desc.String
			fd.Description = &v
		}
		if apparatNr.Valid {
			v := int(apparatNr.Int64)
			fd.ApparatNr = &v
		}
		if systemPartID.Valid {
			id, err := uuid.Parse(systemPartID.String)
			if err != nil {
				return nil, err
			}
			fd.SystemPartID = &id
		}
		if specificationID.Valid {
			id, err := uuid.Parse(specificationID.String)
			if err != nil {
				return nil, err
			}
			fd.SpecificationID = &id
		}
		if projectID.Valid {
			id, err := uuid.Parse(projectID.String)
			if err != nil {
				return nil, err
			}
			fd.ProjectID = &id
		}

		out = append(out, &fd)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *fieldDeviceRepo) Create(entity *domainFacility.FieldDevice) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO field_devices (id, created_at, updated_at, deleted_at, bmk, description, apparat_nr, sps_controller_system_type_id, system_part_id, specification_id, project_id, apparat_id) " +
		"VALUES (?, ?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		entity.ID,
		entity.CreatedAt,
		entity.UpdatedAt,
		argStringPtr(entity.BMK),
		argStringPtr(entity.Description),
		argIntPtr(entity.ApparatNr),
		entity.SPSControllerSystemTypeID,
		argUUIDPtr(entity.SystemPartID),
		argUUIDPtr(entity.SpecificationID),
		argUUIDPtr(entity.ProjectID),
		entity.ApparatID,
	)
	return err
}

func (r *fieldDeviceRepo) Update(entity *domainFacility.FieldDevice) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	q := "UPDATE field_devices SET updated_at = ?, bmk = ?, description = ?, apparat_nr = ?, sps_controller_system_type_id = ?, system_part_id = ?, specification_id = ?, project_id = ?, apparat_id = ? " +
		"WHERE deleted_at IS NULL AND id = ?"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		entity.UpdatedAt,
		argStringPtr(entity.BMK),
		argStringPtr(entity.Description),
		argIntPtr(entity.ApparatNr),
		entity.SPSControllerSystemTypeID,
		argUUIDPtr(entity.SystemPartID),
		argUUIDPtr(entity.SpecificationID),
		argUUIDPtr(entity.ProjectID),
		entity.ApparatID,
		entity.ID,
	)
	return err
}

func (r *fieldDeviceRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE field_devices SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := r.db.Exec(q, args...)
	return err
}

func (r *fieldDeviceRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
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
	args := make([]any, 0, 8)
	if strings.TrimSpace(params.Search) != "" {
		like := sqlutil.LikeOperator(r.dialect)
		pattern := "%" + params.Search + "%"
		where += " AND ((bmk " + like + " ?) OR (description " + like + " ?))"
		args = append(args, pattern, pattern)
	}

	countQ := "SELECT COUNT(*) FROM field_devices WHERE " + where
	countQ = sqlutil.Rebind(r.dialect, countQ)
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, bmk, description, apparat_nr, sps_controller_system_type_id, system_part_id, specification_id, project_id, apparat_id " +
		"FROM field_devices WHERE " + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	dataQ = sqlutil.Rebind(r.dialect, dataQ)

	dataArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainFacility.FieldDevice, 0, limit)
	for rows.Next() {
		var fd domainFacility.FieldDevice
		var deletedAt sql.NullTime
		var bmk sql.NullString
		var desc sql.NullString
		var apparatNr sql.NullInt64
		var systemPartID sql.NullString
		var specificationID sql.NullString
		var projectID sql.NullString
		if err := rows.Scan(
			&fd.ID,
			&fd.CreatedAt,
			&fd.UpdatedAt,
			&deletedAt,
			&bmk,
			&desc,
			&apparatNr,
			&fd.SPSControllerSystemTypeID,
			&systemPartID,
			&specificationID,
			&projectID,
			&fd.ApparatID,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			fd.DeletedAt = &t
		}
		if bmk.Valid {
			v := bmk.String
			fd.BMK = &v
		}
		if desc.Valid {
			v := desc.String
			fd.Description = &v
		}
		if apparatNr.Valid {
			v := int(apparatNr.Int64)
			fd.ApparatNr = &v
		}
		if systemPartID.Valid {
			id, err := uuid.Parse(systemPartID.String)
			if err != nil {
				return nil, err
			}
			fd.SystemPartID = &id
		}
		if specificationID.Valid {
			id, err := uuid.Parse(specificationID.String)
			if err != nil {
				return nil, err
			}
			fd.SpecificationID = &id
		}
		if projectID.Valid {
			id, err := uuid.Parse(projectID.String)
			if err != nil {
				return nil, err
			}
			fd.ProjectID = &id
		}
		items = append(items, fd)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.FieldDevice]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
