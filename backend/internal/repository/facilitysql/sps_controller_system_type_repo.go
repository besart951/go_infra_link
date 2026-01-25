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

type spsControllerSystemTypeRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewSPSControllerSystemTypeRepository(db *sql.DB, driver string) domainFacility.SPSControllerSystemTypeStore {
	return &spsControllerSystemTypeRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *spsControllerSystemTypeRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SPSControllerSystemType, error) {
	if len(ids) == 0 {
		return []*domainFacility.SPSControllerSystemType{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, number, document_name, sps_controller_id, system_type_id " +
		"FROM sps_controller_system_types WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
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

	out := make([]*domainFacility.SPSControllerSystemType, 0, len(ids))
	for rows.Next() {
		var s domainFacility.SPSControllerSystemType
		var deletedAt sql.NullTime
		var number sql.NullInt64
		var documentName sql.NullString
		if err := rows.Scan(
			&s.ID,
			&s.CreatedAt,
			&s.UpdatedAt,
			&deletedAt,
			&number,
			&documentName,
			&s.SPSControllerID,
			&s.SystemTypeID,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			s.DeletedAt = &t
		}
		if number.Valid {
			v := int(number.Int64)
			s.Number = &v
		}
		if documentName.Valid {
			v := documentName.String
			s.DocumentName = &v
		}
		out = append(out, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *spsControllerSystemTypeRepo) Create(entity *domainFacility.SPSControllerSystemType) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO sps_controller_system_types (id, created_at, updated_at, deleted_at, number, document_name, sps_controller_id, system_type_id) " +
		"VALUES (?, ?, ?, NULL, ?, ?, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		entity.ID,
		entity.CreatedAt,
		entity.UpdatedAt,
		argIntPtr(entity.Number),
		argStringPtr(entity.DocumentName),
		entity.SPSControllerID,
		entity.SystemTypeID,
	)
	return err
}

func (r *spsControllerSystemTypeRepo) Update(entity *domainFacility.SPSControllerSystemType) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	q := "UPDATE sps_controller_system_types SET updated_at = ?, number = ?, document_name = ?, sps_controller_id = ?, system_type_id = ? WHERE deleted_at IS NULL AND id = ?"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		entity.UpdatedAt,
		argIntPtr(entity.Number),
		argStringPtr(entity.DocumentName),
		entity.SPSControllerID,
		entity.SystemTypeID,
		entity.ID,
	)
	return err
}

func (r *spsControllerSystemTypeRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE sps_controller_system_types SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := r.db.Exec(q, args...)
	return err
}

func (r *spsControllerSystemTypeRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	page := params.Page
	limit := params.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	where := "s.deleted_at IS NULL"
	args := make([]any, 0, 4)
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + params.Search + "%"
		where += " AND (s.document_name " + sqlutil.LikeOperator(r.dialect) + " ? OR sc.device_name " + sqlutil.LikeOperator(r.dialect) + " ? OR st.name " + sqlutil.LikeOperator(r.dialect) + " ?)"
		args = append(args, pattern, pattern, pattern)
	}

	countQ := `SELECT COUNT(*) FROM sps_controller_system_types s
		LEFT JOIN sps_controllers sc ON s.sps_controller_id = sc.id
		LEFT JOIN system_types st ON s.system_type_id = st.id
		WHERE ` + where
	countQ = sqlutil.Rebind(r.dialect, countQ)
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := `SELECT s.id, s.created_at, s.updated_at, s.deleted_at, s.number, s.document_name, s.sps_controller_id, s.system_type_id,
		sc.device_name, st.name
		FROM sps_controller_system_types s
		LEFT JOIN sps_controllers sc ON s.sps_controller_id = sc.id
		LEFT JOIN system_types st ON s.system_type_id = st.id
		WHERE ` + where + ` ORDER BY s.created_at DESC LIMIT ? OFFSET ?`
	
	dataQ = sqlutil.Rebind(r.dialect, dataQ)
	dataArgs := append(append([]any{}, args...), limit, offset)

	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainFacility.SPSControllerSystemType, 0, limit)
	for rows.Next() {
		var s domainFacility.SPSControllerSystemType
		var deletedAt sql.NullTime
		var number sql.NullInt64
		var documentName sql.NullString
		var spsName sql.NullString
		var sysName sql.NullString

		if err := rows.Scan(
			&s.ID,
			&s.CreatedAt,
			&s.UpdatedAt,
			&deletedAt,
			&number,
			&documentName,
			&s.SPSControllerID,
			&s.SystemTypeID,
			&spsName,
			&sysName,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			s.DeletedAt = &t
		}
		if number.Valid {
			v := int(number.Int64)
			s.Number = &v
		}
		if documentName.Valid {
			v := documentName.String
			s.DocumentName = &v
		}
		if spsName.Valid {
			s.SPSController.DeviceName = spsName.String
		}
		if sysName.Valid {
			s.SystemType.Name = sysName.String
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.SPSControllerSystemType]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *spsControllerSystemTypeRepo) SoftDeleteBySPSControllerIDs(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE sps_controller_system_types SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND sps_controller_id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := r.db.Exec(q, args...)
	return err
}

func argIntPtr(i *int) any {
	if i == nil {
		return nil
	}
	return *i
}

func argStringPtr(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}
