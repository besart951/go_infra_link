package facilitysql

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
	"github.com/google/uuid"
)

type systemPartRepo struct {
	db *sql.DB
}

func NewSystemPartRepository(db *sql.DB) domainFacility.SystemPartRepository {
	return &systemPartRepo{db: db}
}

func (r *systemPartRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	if len(ids) == 0 {
		return []*domainFacility.SystemPart{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, short_name, name, description " +
		"FROM system_parts WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(1, len(ids)) + ")"

	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]*domainFacility.SystemPart, 0, len(ids))
	for rows.Next() {
		var sp domainFacility.SystemPart
		var deletedAt sql.NullTime
		var desc sql.NullString
		if err := rows.Scan(&sp.ID, &sp.CreatedAt, &sp.UpdatedAt, &deletedAt, &sp.ShortName, &sp.Name, &desc); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			sp.DeletedAt = &t
		}
		if desc.Valid {
			v := desc.String
			sp.Description = &v
		}
		out = append(out, &sp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *systemPartRepo) Create(entity *domainFacility.SystemPart) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO system_parts (id, created_at, updated_at, deleted_at, short_name, name, description) VALUES ($1, $2, $3, NULL, $4, $5, $6)"

	_, err := r.db.Exec(q, entity.ID, entity.CreatedAt, entity.UpdatedAt, entity.ShortName, entity.Name, argStringPtr(entity.Description))
	return err
}

func (r *systemPartRepo) Update(entity *domainFacility.SystemPart) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	q := "UPDATE system_parts SET updated_at = $1, short_name = $2, name = $3, description = $4 WHERE deleted_at IS NULL AND id = $5"

	_, err := r.db.Exec(q, entity.UpdatedAt, entity.ShortName, entity.Name, argStringPtr(entity.Description), entity.ID)
	return err
}

func (r *systemPartRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE system_parts SET deleted_at = $1, updated_at = $2 WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(3, len(ids)) + ")"

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}
	_, err := r.db.Exec(q, args...)
	return err
}

func (r *systemPartRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemPart], error) {
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
		where += " AND ((short_name ILIKE $1) OR (name ILIKE $2))"
		args = append(args, pattern, pattern)
	}

	countQ := "SELECT COUNT(*) FROM system_parts WHERE " + where
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, short_name, name, description FROM system_parts WHERE " + where + " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	dataArgs := append(append([]any{}, args...), limit, offset)

	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainFacility.SystemPart, 0, limit)
	for rows.Next() {
		var sp domainFacility.SystemPart
		var deletedAt sql.NullTime
		var desc sql.NullString
		if err := rows.Scan(&sp.ID, &sp.CreatedAt, &sp.UpdatedAt, &deletedAt, &sp.ShortName, &sp.Name, &desc); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			sp.DeletedAt = &t
		}
		if desc.Valid {
			v := desc.String
			sp.Description = &v
		}
		items = append(items, sp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.SystemPart]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
