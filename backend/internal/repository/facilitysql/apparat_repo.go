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

type apparatRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewApparatRepository(db *sql.DB, driver string) domainFacility.ApparatRepository {
	return &apparatRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *apparatRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	if len(ids) == 0 {
		return []*domainFacility.Apparat{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, short_name, name, description " +
		"FROM apparats WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
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

	out := make([]*domainFacility.Apparat, 0, len(ids))
	for rows.Next() {
		var a domainFacility.Apparat
		var deletedAt sql.NullTime
		var desc sql.NullString
		if err := rows.Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt, &deletedAt, &a.ShortName, &a.Name, &desc); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			a.DeletedAt = &t
		}
		if desc.Valid {
			v := desc.String
			a.Description = &v
		}
		out = append(out, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *apparatRepo) Create(entity *domainFacility.Apparat) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO apparats (id, created_at, updated_at, deleted_at, short_name, name, description) VALUES (?, ?, ?, NULL, ?, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(q, entity.ID, entity.CreatedAt, entity.UpdatedAt, entity.ShortName, entity.Name, argStringPtr(entity.Description))
	return err
}

func (r *apparatRepo) Update(entity *domainFacility.Apparat) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	q := "UPDATE apparats SET updated_at = ?, short_name = ?, name = ?, description = ? WHERE deleted_at IS NULL AND id = ?"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(q, entity.UpdatedAt, entity.ShortName, entity.Name, argStringPtr(entity.Description), entity.ID)
	return err
}

func (r *apparatRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE apparats SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}
	_, err := r.db.Exec(q, args...)
	return err
}

func (r *apparatRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
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
		like := sqlutil.LikeOperator(r.dialect)
		where += " AND ((short_name " + like + " ?) OR (name " + like + " ?))"
		args = append(args, pattern, pattern)
	}

	countQ := "SELECT COUNT(*) FROM apparats WHERE " + where
	countQ = sqlutil.Rebind(r.dialect, countQ)
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, short_name, name, description FROM apparats WHERE " + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	dataQ = sqlutil.Rebind(r.dialect, dataQ)
	dataArgs := append(append([]any{}, args...), limit, offset)

	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainFacility.Apparat, 0, limit)
	for rows.Next() {
		var a domainFacility.Apparat
		var deletedAt sql.NullTime
		var desc sql.NullString
		if err := rows.Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt, &deletedAt, &a.ShortName, &a.Name, &desc); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			a.DeletedAt = &t
		}
		if desc.Valid {
			v := desc.String
			a.Description = &v
		}
		items = append(items, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.Apparat]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
