package auth

import (
	"database/sql"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
	"github.com/google/uuid"
)

type loginAttemptRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewLoginAttemptRepository(db *sql.DB, driver string) domainAuth.LoginAttemptRepository {
	return &loginAttemptRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *loginAttemptRepo) Create(attempt *domainAuth.LoginAttempt) error {
	now := time.Now().UTC()
	if attempt.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		attempt.ID = id
	}
	if attempt.CreatedAt.IsZero() {
		attempt.CreatedAt = now
	}

	q := "INSERT INTO login_attempts (id, created_at, user_id, email, ip, user_agent, success, failure_reason) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(q, attempt.ID, attempt.CreatedAt, attempt.UserID, attempt.Email, attempt.IP, attempt.UserAgent, attempt.Success, attempt.FailureReason)
	return err
}

func (r *loginAttemptRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainAuth.LoginAttempt], error) {
	page := params.Page
	limit := params.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	where := "1=1"
	args := make([]any, 0, 4)
	if strings.TrimSpace(params.Search) != "" {
		like := sqlutil.LikeOperator(r.dialect)
		pattern := "%" + params.Search + "%"
		where += " AND (email " + like + " ? OR ip " + like + " ? OR user_agent " + like + " ? OR failure_reason " + like + " ?)"
		args = append(args, pattern, pattern, pattern, pattern)
	}

	countQ := "SELECT COUNT(*) FROM login_attempts WHERE " + where
	countQ = sqlutil.Rebind(r.dialect, countQ)
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, user_id, email, ip, user_agent, success, failure_reason FROM login_attempts WHERE " + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	dataQ = sqlutil.Rebind(r.dialect, dataQ)

	dataArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainAuth.LoginAttempt, 0, limit)
	for rows.Next() {
		var a domainAuth.LoginAttempt
		var userID sql.NullString
		var email sql.NullString
		var ip sql.NullString
		var ua sql.NullString
		var failureReason sql.NullString
		if err := rows.Scan(&a.ID, &a.CreatedAt, &userID, &email, &ip, &ua, &a.Success, &failureReason); err != nil {
			return nil, err
		}
		if userID.Valid {
			id, err := uuid.Parse(userID.String)
			if err != nil {
				return nil, err
			}
			a.UserID = &id
		}
		if email.Valid {
			v := email.String
			a.Email = &v
		}
		if ip.Valid {
			v := ip.String
			a.IP = &v
		}
		if ua.Valid {
			v := ua.String
			a.UserAgent = &v
		}
		if failureReason.Valid {
			v := failureReason.String
			a.FailureReason = &v
		}
		items = append(items, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainAuth.LoginAttempt]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
