package project

import (
	"database/sql"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"

	"github.com/google/uuid"
)

type projectRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewProjectRepository(db *sql.DB, driver string) domainProject.ProjectRepository {
	return &projectRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *projectRepo) GetByIds(ids []uuid.UUID) ([]*domainProject.Project, error) {
	if len(ids) == 0 {
		return []*domainProject.Project{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, name, description, status, start_date, phase_id, creator_id " +
		"FROM projects WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
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

	out := make([]*domainProject.Project, 0, len(ids))
	for rows.Next() {
		var p domainProject.Project
		var deletedAt sql.NullTime
		var desc sql.NullString
		var startDate sql.NullTime

		if err := rows.Scan(
			&p.ID,
			&p.CreatedAt,
			&p.UpdatedAt,
			&deletedAt,
			&p.Name,
			&desc,
			&p.Status,
			&startDate,
			&p.PhaseID,
			&p.CreatorID,
		); err != nil {
			return nil, err
		}

		if deletedAt.Valid {
			t := deletedAt.Time
			p.DeletedAt = &t
		}
		if desc.Valid {
			p.Description = desc.String
		}
		if startDate.Valid {
			t := startDate.Time
			p.StartDate = &t
		}

		out = append(out, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *projectRepo) Create(entity *domainProject.Project) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO projects (id, created_at, updated_at, deleted_at, name, description, status, start_date, phase_id, creator_id) " +
		"VALUES (?, ?, ?, NULL, ?, ?, ?, ?, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		entity.ID,
		entity.CreatedAt,
		entity.UpdatedAt,
		entity.Name,
		nullIfEmpty(entity.Description),
		entity.Status,
		entity.StartDate,
		entity.PhaseID,
		entity.CreatorID,
	)
	return err
}

func (r *projectRepo) Update(entity *domainProject.Project) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	q := "UPDATE projects SET updated_at = ?, name = ?, description = ?, status = ?, start_date = ?, phase_id = ?, creator_id = ? " +
		"WHERE deleted_at IS NULL AND id = ?"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		entity.UpdatedAt,
		entity.Name,
		nullIfEmpty(entity.Description),
		entity.Status,
		entity.StartDate,
		entity.PhaseID,
		entity.CreatorID,
		entity.ID,
	)
	return err
}

func (r *projectRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE projects SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := r.db.Exec(q, args...)
	return err
}

func (r *projectRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainProject.Project], error) {
	page := params.Page
	limit := params.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	searchFields := []string{"name", "description"}
	where := "deleted_at IS NULL"
	args := make([]any, 0, 8)
	if strings.TrimSpace(params.Search) != "" {
		like := sqlutil.LikeOperator(r.dialect)
		pattern := "%" + params.Search + "%"
		parts := make([]string, 0, len(searchFields))
		for _, f := range searchFields {
			parts = append(parts, f+" "+like+" ?")
			args = append(args, pattern)
		}
		where += " AND (" + strings.Join(parts, " OR ") + ")"
	}

	countQ := "SELECT COUNT(*) FROM projects WHERE " + where
	countQ = sqlutil.Rebind(r.dialect, countQ)
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, name, description, status, start_date, phase_id, creator_id FROM projects WHERE " + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	dataQ = sqlutil.Rebind(r.dialect, dataQ)

	dataArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainProject.Project, 0, limit)
	for rows.Next() {
		var p domainProject.Project
		var deletedAt sql.NullTime
		var desc sql.NullString
		var startDate sql.NullTime
		if err := rows.Scan(
			&p.ID,
			&p.CreatedAt,
			&p.UpdatedAt,
			&deletedAt,
			&p.Name,
			&desc,
			&p.Status,
			&startDate,
			&p.PhaseID,
			&p.CreatorID,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			p.DeletedAt = &t
		}
		if desc.Valid {
			p.Description = desc.String
		}
		if startDate.Valid {
			t := startDate.Time
			p.StartDate = &t
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainProject.Project]{
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
