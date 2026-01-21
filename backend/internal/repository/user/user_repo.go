package user

import (
	"database/sql"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
	"github.com/google/uuid"
)

type userRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewUserRepository(db *sql.DB, driver string) domainUser.UserRepository {
	return &userRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *userRepo) GetByIds(ids []uuid.UUID) ([]*domainUser.User, error) {
	if len(ids) == 0 {
		return []*domainUser.User{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, first_name, last_name, email, password, is_active, role, disabled_at, locked_until, failed_login_attempts, last_login_at, created_by_id " +
		"FROM users WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
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

	out := make([]*domainUser.User, 0, len(ids))
	for rows.Next() {
		var u domainUser.User
		var deletedAt sql.NullTime
		var disabledAt sql.NullTime
		var lockedUntil sql.NullTime
		var lastLoginAt sql.NullTime
		var role sql.NullString
		var createdByID sql.NullString

		err := rows.Scan(
			&u.ID,
			&u.CreatedAt,
			&u.UpdatedAt,
			&deletedAt,
			&u.FirstName,
			&u.LastName,
			&u.Email,
			&u.Password,
			&u.IsActive,
			&role,
			&disabledAt,
			&lockedUntil,
			&u.FailedLoginAttempts,
			&lastLoginAt,
			&createdByID,
		)
		if err != nil {
			return nil, err
		}

		if deletedAt.Valid {
			t := deletedAt.Time
			u.DeletedAt = &t
		}
		if role.Valid {
			u.Role = domainUser.Role(role.String)
		}
		if disabledAt.Valid {
			t := disabledAt.Time
			u.DisabledAt = &t
		}
		if lockedUntil.Valid {
			t := lockedUntil.Time
			u.LockedUntil = &t
		}
		if lastLoginAt.Valid {
			t := lastLoginAt.Time
			u.LastLoginAt = &t
		}
		if createdByID.Valid {
			id, err := uuid.Parse(createdByID.String)
			if err != nil {
				return nil, err
			}
			u.CreatedByID = &id
		}

		out = append(out, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *userRepo) GetByEmail(email string) (*domainUser.User, error) {
	q := "SELECT id, created_at, updated_at, deleted_at, first_name, last_name, email, password, is_active, role, disabled_at, locked_until, failed_login_attempts, last_login_at, created_by_id " +
		"FROM users WHERE deleted_at IS NULL AND email = ? LIMIT 1"
	q = sqlutil.Rebind(r.dialect, q)

	var u domainUser.User
	var deletedAt sql.NullTime
	var disabledAt sql.NullTime
	var lockedUntil sql.NullTime
	var lastLoginAt sql.NullTime
	var role sql.NullString
	var createdByID sql.NullString

	err := r.db.QueryRow(q, email).Scan(
		&u.ID,
		&u.CreatedAt,
		&u.UpdatedAt,
		&deletedAt,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.IsActive,
		&role,
		&disabledAt,
		&lockedUntil,
		&u.FailedLoginAttempts,
		&lastLoginAt,
		&createdByID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if deletedAt.Valid {
		t := deletedAt.Time
		u.DeletedAt = &t
	}
	if role.Valid {
		u.Role = domainUser.Role(role.String)
	}
	if disabledAt.Valid {
		t := disabledAt.Time
		u.DisabledAt = &t
	}
	if lockedUntil.Valid {
		t := lockedUntil.Time
		u.LockedUntil = &t
	}
	if lastLoginAt.Valid {
		t := lastLoginAt.Time
		u.LastLoginAt = &t
	}
	if createdByID.Valid {
		id, err := uuid.Parse(createdByID.String)
		if err != nil {
			return nil, err
		}
		u.CreatedByID = &id
	}

	return &u, nil
}

func (r *userRepo) Create(entity *domainUser.User) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO users (id, created_at, updated_at, deleted_at, first_name, last_name, email, password, is_active, role, disabled_at, locked_until, failed_login_attempts, last_login_at, created_by_id) " +
		"VALUES (?, ?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		entity.ID,
		entity.CreatedAt,
		entity.UpdatedAt,
		entity.FirstName,
		entity.LastName,
		entity.Email,
		entity.Password,
		entity.IsActive,
		entity.Role,
		entity.DisabledAt,
		entity.LockedUntil,
		entity.FailedLoginAttempts,
		entity.LastLoginAt,
		entity.CreatedByID,
	)
	return err
}

func (r *userRepo) Update(entity *domainUser.User) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	setPassword := strings.TrimSpace(entity.Password) != ""

	var q string
	var args []any
	if setPassword {
		q = "UPDATE users SET updated_at = ?, first_name = ?, last_name = ?, email = ?, password = ?, is_active = ?, role = ?, disabled_at = ?, locked_until = ?, failed_login_attempts = ?, last_login_at = ?, created_by_id = ? WHERE deleted_at IS NULL AND id = ?"
		args = []any{entity.UpdatedAt, entity.FirstName, entity.LastName, entity.Email, entity.Password, entity.IsActive, entity.Role, entity.DisabledAt, entity.LockedUntil, entity.FailedLoginAttempts, entity.LastLoginAt, entity.CreatedByID, entity.ID}
	} else {
		q = "UPDATE users SET updated_at = ?, first_name = ?, last_name = ?, email = ?, is_active = ?, role = ?, disabled_at = ?, locked_until = ?, failed_login_attempts = ?, last_login_at = ?, created_by_id = ? WHERE deleted_at IS NULL AND id = ?"
		args = []any{entity.UpdatedAt, entity.FirstName, entity.LastName, entity.Email, entity.IsActive, entity.Role, entity.DisabledAt, entity.LockedUntil, entity.FailedLoginAttempts, entity.LastLoginAt, entity.CreatedByID, entity.ID}
	}
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(q, args...)
	return err
}

func (r *userRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE users SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := r.db.Exec(q, args...)
	return err
}

func (r *userRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainUser.User], error) {
	page := params.Page
	limit := params.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	searchFields := []string{"first_name", "last_name", "email"}
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

	countQ := "SELECT COUNT(*) FROM users WHERE " + where
	countQ = sqlutil.Rebind(r.dialect, countQ)

	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, first_name, last_name, email, password, is_active, role, disabled_at, locked_until, failed_login_attempts, last_login_at, created_by_id FROM users WHERE " + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	dataQ = sqlutil.Rebind(r.dialect, dataQ)

	dataArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainUser.User, 0, limit)
	for rows.Next() {
		var u domainUser.User
		var deletedAt sql.NullTime
		var disabledAt sql.NullTime
		var lockedUntil sql.NullTime
		var lastLoginAt sql.NullTime
		var role sql.NullString
		var createdByID sql.NullString
		if err := rows.Scan(
			&u.ID,
			&u.CreatedAt,
			&u.UpdatedAt,
			&deletedAt,
			&u.FirstName,
			&u.LastName,
			&u.Email,
			&u.Password,
			&u.IsActive,
			&role,
			&disabledAt,
			&lockedUntil,
			&u.FailedLoginAttempts,
			&lastLoginAt,
			&createdByID,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			u.DeletedAt = &t
		}
		if role.Valid {
			u.Role = domainUser.Role(role.String)
		}
		if disabledAt.Valid {
			t := disabledAt.Time
			u.DisabledAt = &t
		}
		if lockedUntil.Valid {
			t := lockedUntil.Time
			u.LockedUntil = &t
		}
		if lastLoginAt.Valid {
			t := lastLoginAt.Time
			u.LastLoginAt = &t
		}
		if createdByID.Valid {
			id, err := uuid.Parse(createdByID.String)
			if err != nil {
				return nil, err
			}
			u.CreatedByID = &id
		}
		items = append(items, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainUser.User]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

var _ domainUser.UserEmailRepository = (*userRepo)(nil)
