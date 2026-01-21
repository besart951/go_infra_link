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

type systemTypeRepo struct {
	db *sql.DB
}

func NewSystemTypeRepository(db *sql.DB) domainFacility.SystemTypeRepository {
	return &systemTypeRepo{db: db}
}

func (r *systemTypeRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SystemType, error) {
	if len(ids) == 0 {
		return []*domainFacility.SystemType{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, number_min, number_max, name " +
		"FROM system_types WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(1, len(ids)) + ")"

	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]*domainFacility.SystemType, 0, len(ids))
	for rows.Next() {
		var st domainFacility.SystemType
		var deletedAt sql.NullTime
		if err := rows.Scan(&st.ID, &st.CreatedAt, &st.UpdatedAt, &deletedAt, &st.NumberMin, &st.NumberMax, &st.Name); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			st.DeletedAt = &t
		}
		out = append(out, &st)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *systemTypeRepo) Create(entity *domainFacility.SystemType) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO system_types (id, created_at, updated_at, deleted_at, number_min, number_max, name) VALUES ($1, $2, $3, NULL, $4, $5, $6)"

	_, err := r.db.Exec(q, entity.ID, entity.CreatedAt, entity.UpdatedAt, entity.NumberMin, entity.NumberMax, entity.Name)
	return err
}

func (r *systemTypeRepo) Update(entity *domainFacility.SystemType) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	q := "UPDATE system_types SET updated_at = $1, number_min = $2, number_max = $3, name = $4 WHERE deleted_at IS NULL AND id = $5"

	_, err := r.db.Exec(q, entity.UpdatedAt, entity.NumberMin, entity.NumberMax, entity.Name, entity.ID)
	return err
}

func (r *systemTypeRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE system_types SET deleted_at = $1, updated_at = $2 WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(3, len(ids)) + ")"

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}
	_, err := r.db.Exec(q, args...)
	return err
}

func (r *systemTypeRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemType], error) {
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
	args := make([]any, 0, 4)
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + params.Search + "%"
		where += " AND (name ILIKE $1)"
		args = append(args, pattern)
	}

	countQ := "SELECT COUNT(*) FROM system_types WHERE " + where
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, number_min, number_max, name FROM system_types WHERE " + where + " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	dataArgs := append(append([]any{}, args...), limit, offset)

	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainFacility.SystemType, 0, limit)
	for rows.Next() {
		var st domainFacility.SystemType
		var deletedAt sql.NullTime
		if err := rows.Scan(&st.ID, &st.CreatedAt, &st.UpdatedAt, &deletedAt, &st.NumberMin, &st.NumberMax, &st.Name); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			st.DeletedAt = &t
		}
		items = append(items, st)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.SystemType]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
