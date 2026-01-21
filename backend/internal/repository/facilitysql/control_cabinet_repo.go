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

type controlCabinetRepo struct {
	db *sql.DB
}

func NewControlCabinetRepository(db *sql.DB) domainFacility.ControlCabinetRepository {
	return &controlCabinetRepo{db: db}
}

func (r *controlCabinetRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.ControlCabinet, error) {
	if len(ids) == 0 {
		return []*domainFacility.ControlCabinet{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, building_id, project_id, control_cabinet_nr " +
		"FROM control_cabinets WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(1, len(ids)) + ")"

	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]*domainFacility.ControlCabinet, 0, len(ids))
	for rows.Next() {
		var c domainFacility.ControlCabinet
		var deletedAt sql.NullTime
		var projectID sql.NullString
		var nr sql.NullString
		if err := rows.Scan(
			&c.ID,
			&c.CreatedAt,
			&c.UpdatedAt,
			&deletedAt,
			&c.BuildingID,
			&projectID,
			&nr,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			c.DeletedAt = &t
		}
		if projectID.Valid {
			id, err := uuid.Parse(projectID.String)
			if err != nil {
				return nil, err
			}
			c.ProjectID = &id
		}
		if nr.Valid {
			v := nr.String
			c.ControlCabinetNr = &v
		}
		out = append(out, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *controlCabinetRepo) Create(entity *domainFacility.ControlCabinet) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO control_cabinets (id, created_at, updated_at, deleted_at, building_id, project_id, control_cabinet_nr) " +
		"VALUES ($1, $2, $3, NULL, $4, $5, $6)"

	_, err := r.db.Exec(q, entity.ID, entity.CreatedAt, entity.UpdatedAt, entity.BuildingID, argUUIDPtr(entity.ProjectID), argStringPtr(entity.ControlCabinetNr))
	return err
}

func (r *controlCabinetRepo) Update(entity *domainFacility.ControlCabinet) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	q := "UPDATE control_cabinets SET updated_at = $1, building_id = $2, project_id = $3, control_cabinet_nr = $4 WHERE deleted_at IS NULL AND id = $5"

	_, err := r.db.Exec(q, entity.UpdatedAt, entity.BuildingID, argUUIDPtr(entity.ProjectID), argStringPtr(entity.ControlCabinetNr), entity.ID)
	return err
}

func (r *controlCabinetRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE control_cabinets SET deleted_at = $1, updated_at = $2 WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(3, len(ids)) + ")"

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := r.db.Exec(q, args...)
	return err
}

func (r *controlCabinetRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
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
		where += " AND (control_cabinet_nr ILIKE $1)"
		args = append(args, pattern)
	}

	countQ := "SELECT COUNT(*) FROM control_cabinets WHERE " + where
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, building_id, project_id, control_cabinet_nr FROM control_cabinets WHERE " + where + " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	dataArgs := append(append([]any{}, args...), limit, offset)

	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainFacility.ControlCabinet, 0, limit)
	for rows.Next() {
		var c domainFacility.ControlCabinet
		var deletedAt sql.NullTime
		var projectID sql.NullString
		var nr sql.NullString
		if err := rows.Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt, &deletedAt, &c.BuildingID, &projectID, &nr); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			c.DeletedAt = &t
		}
		if projectID.Valid {
			id, err := uuid.Parse(projectID.String)
			if err != nil {
				return nil, err
			}
			c.ProjectID = &id
		}
		if nr.Valid {
			v := nr.String
			c.ControlCabinetNr = &v
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.ControlCabinet]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
