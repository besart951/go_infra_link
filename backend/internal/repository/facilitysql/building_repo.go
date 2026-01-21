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

type buildingRepo struct {
	db *sql.DB
}

func NewBuildingRepository(db *sql.DB) domainFacility.BuildingRepository {
	return &buildingRepo{db: db}
}

func (r *buildingRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.Building, error) {
	if len(ids) == 0 {
		return []*domainFacility.Building{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, iws_code, building_group " +
		"FROM buildings WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(1, len(ids)) + ")"

	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]*domainFacility.Building, 0, len(ids))
	for rows.Next() {
		var b domainFacility.Building
		var deletedAt sql.NullTime
		var iwsCode sql.NullString
		if err := rows.Scan(
			&b.ID,
			&b.CreatedAt,
			&b.UpdatedAt,
			&deletedAt,
			&iwsCode,
			&b.BuildingGroup,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			b.DeletedAt = &t
		}
		if iwsCode.Valid {
			b.IWSCode = iwsCode.String
		}
		out = append(out, &b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *buildingRepo) Create(entity *domainFacility.Building) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO buildings (id, created_at, updated_at, deleted_at, iws_code, building_group) " +
		"VALUES ($1, $2, $3, NULL, $4, $5)"

	_, err := r.db.Exec(q, entity.ID, entity.CreatedAt, entity.UpdatedAt, nullIfEmpty(entity.IWSCode), entity.BuildingGroup)
	return err
}

func (r *buildingRepo) Update(entity *domainFacility.Building) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	q := "UPDATE buildings SET updated_at = $1, iws_code = $2, building_group = $3 WHERE deleted_at IS NULL AND id = $4"

	_, err := r.db.Exec(q, entity.UpdatedAt, nullIfEmpty(entity.IWSCode), entity.BuildingGroup, entity.ID)
	return err
}

func (r *buildingRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE buildings SET deleted_at = $1, updated_at = $2 WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(3, len(ids)) + ")"

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := r.db.Exec(q, args...)
	return err
}

func (r *buildingRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Building], error) {
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
		where += " AND (iws_code ILIKE $1)"
		args = append(args, pattern)
	}

	countQ := "SELECT COUNT(*) FROM buildings WHERE " + where
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, iws_code, building_group FROM buildings WHERE " + where + " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	dataArgs := append(append([]any{}, args...), limit, offset)

	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainFacility.Building, 0, limit)
	for rows.Next() {
		var b domainFacility.Building
		var deletedAt sql.NullTime
		var iwsCode sql.NullString
		if err := rows.Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt, &deletedAt, &iwsCode, &b.BuildingGroup); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			b.DeletedAt = &t
		}
		if iwsCode.Valid {
			b.IWSCode = iwsCode.String
		}
		items = append(items, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.Building]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
