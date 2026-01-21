package team

import (
	"database/sql"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
	"github.com/google/uuid"
)

type teamRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewTeamRepository(db *sql.DB, driver string) domainTeam.TeamRepository {
	return &teamRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *teamRepo) GetByIds(ids []uuid.UUID) ([]*domainTeam.Team, error) {
	if len(ids) == 0 {
		return []*domainTeam.Team{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, name, description FROM teams WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
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

	out := make([]*domainTeam.Team, 0, len(ids))
	for rows.Next() {
		var t domainTeam.Team
		var deletedAt sql.NullTime
		var desc sql.NullString
		if err := rows.Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt, &deletedAt, &t.Name, &desc); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			v := deletedAt.Time
			t.DeletedAt = &v
		}
		if desc.Valid {
			v := desc.String
			t.Description = &v
		}
		out = append(out, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *teamRepo) Create(entity *domainTeam.Team) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO teams (id, created_at, updated_at, deleted_at, name, description) VALUES (?, ?, ?, NULL, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(q, entity.ID, entity.CreatedAt, entity.UpdatedAt, entity.Name, entity.Description)
	return err
}

func (r *teamRepo) Update(entity *domainTeam.Team) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	q := "UPDATE teams SET updated_at = ?, name = ?, description = ? WHERE deleted_at IS NULL AND id = ?"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(q, entity.UpdatedAt, entity.Name, entity.Description, entity.ID)
	return err
}

func (r *teamRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE teams SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := r.db.Exec(q, args...)
	return err
}

func (r *teamRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainTeam.Team], error) {
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
	args := make([]any, 0, 2)
	if strings.TrimSpace(params.Search) != "" {
		like := sqlutil.LikeOperator(r.dialect)
		pattern := "%" + params.Search + "%"
		where += " AND (name " + like + " ? OR description " + like + " ?)"
		args = append(args, pattern, pattern)
	}

	countQ := "SELECT COUNT(*) FROM teams WHERE " + where
	countQ = sqlutil.Rebind(r.dialect, countQ)
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, name, description FROM teams WHERE " + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	dataQ = sqlutil.Rebind(r.dialect, dataQ)
	dataArgs := append(append([]any{}, args...), limit, offset)

	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainTeam.Team, 0, limit)
	for rows.Next() {
		var t domainTeam.Team
		var deletedAt sql.NullTime
		var desc sql.NullString
		if err := rows.Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt, &deletedAt, &t.Name, &desc); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			v := deletedAt.Time
			t.DeletedAt = &v
		}
		if desc.Valid {
			v := desc.String
			t.Description = &v
		}
		items = append(items, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainTeam.Team]{Items: items, Total: total, Page: page, TotalPages: domain.CalculateTotalPages(total, limit)}, nil
}
