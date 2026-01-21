package auth

import (
	"database/sql"
	"time"

	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
	"github.com/google/uuid"
)

type passwordResetRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewPasswordResetTokenRepository(db *sql.DB, driver string) domainAuth.PasswordResetTokenRepository {
	return &passwordResetRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *passwordResetRepo) Create(token *domainAuth.PasswordResetToken) error {
	now := time.Now().UTC()
	if err := token.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO password_reset_tokens (id, created_at, updated_at, deleted_at, user_id, token_hash, token_salt, expires_at, used_at, created_by_admin_id) VALUES (?, ?, ?, NULL, ?, ?, ?, ?, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(q, token.ID, token.CreatedAt, token.UpdatedAt, token.UserID, token.TokenHash, token.TokenSalt, token.ExpiresAt, token.UsedAt, token.CreatedByAdminID)
	return err
}

func (r *passwordResetRepo) GetByTokenHash(tokenHash string) (*domainAuth.PasswordResetToken, error) {
	q := "SELECT id, created_at, updated_at, deleted_at, user_id, token_hash, token_salt, expires_at, used_at, created_by_admin_id FROM password_reset_tokens WHERE deleted_at IS NULL AND token_hash = ? LIMIT 1"
	q = sqlutil.Rebind(r.dialect, q)

	var t domainAuth.PasswordResetToken
	var deletedAt sql.NullTime
	var usedAt sql.NullTime
	var createdBy sql.NullString
	err := r.db.QueryRow(q, tokenHash).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt, &deletedAt, &t.UserID, &t.TokenHash, &t.TokenSalt, &t.ExpiresAt, &usedAt, &createdBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if deletedAt.Valid {
		v := deletedAt.Time
		t.DeletedAt = &v
	}
	if usedAt.Valid {
		v := usedAt.Time
		t.UsedAt = &v
	}
	if createdBy.Valid {
		id, err := uuid.Parse(createdBy.String)
		if err != nil {
			return nil, err
		}
		t.CreatedByAdminID = &id
	}

	return &t, nil
}

func (r *passwordResetRepo) MarkUsedByTokenHash(tokenHash string, usedAt time.Time) (bool, error) {
	now := time.Now().UTC()
	q := "UPDATE password_reset_tokens SET used_at = ?, updated_at = ? WHERE deleted_at IS NULL AND token_hash = ? AND used_at IS NULL AND expires_at > ?"
	q = sqlutil.Rebind(r.dialect, q)

	res, err := r.db.Exec(q, usedAt, now, tokenHash, usedAt)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}
