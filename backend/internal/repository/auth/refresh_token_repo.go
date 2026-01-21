package auth

import (
	"database/sql"
	"time"

	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
)

type refreshTokenRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewRefreshTokenRepository(db *sql.DB, driver string) domainAuth.RefreshTokenRepository {
	return &refreshTokenRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *refreshTokenRepo) Create(token *domainAuth.RefreshToken) error {
	now := time.Now().UTC()
	if err := token.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO refresh_tokens (id, created_at, updated_at, deleted_at, user_id, token_hash, expires_at, revoked_at, created_by_ip, user_agent) " +
		"VALUES (?, ?, ?, NULL, ?, ?, ?, ?, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		token.ID,
		token.CreatedAt,
		token.UpdatedAt,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.RevokedAt,
		token.CreatedByIP,
		token.UserAgent,
	)
	return err
}

func (r *refreshTokenRepo) GetByTokenHash(tokenHash string) (*domainAuth.RefreshToken, error) {
	q := "SELECT id, created_at, updated_at, deleted_at, user_id, token_hash, expires_at, revoked_at, created_by_ip, user_agent " +
		"FROM refresh_tokens WHERE deleted_at IS NULL AND token_hash = ? LIMIT 1"
	q = sqlutil.Rebind(r.dialect, q)

	var token domainAuth.RefreshToken
	var deletedAt sql.NullTime
	var revokedAt sql.NullTime
	var createdByIP sql.NullString
	var userAgent sql.NullString

	err := r.db.QueryRow(q, tokenHash).Scan(
		&token.ID,
		&token.CreatedAt,
		&token.UpdatedAt,
		&deletedAt,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&revokedAt,
		&createdByIP,
		&userAgent,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if deletedAt.Valid {
		t := deletedAt.Time
		token.DeletedAt = &t
	}
	if revokedAt.Valid {
		t := revokedAt.Time
		token.RevokedAt = &t
	}
	if createdByIP.Valid {
		v := createdByIP.String
		token.CreatedByIP = &v
	}
	if userAgent.Valid {
		v := userAgent.String
		token.UserAgent = &v
	}

	return &token, nil
}

func (r *refreshTokenRepo) RevokeByTokenHash(tokenHash string, revokedAt time.Time) error {
	q := "UPDATE refresh_tokens SET revoked_at = ?, updated_at = ? WHERE deleted_at IS NULL AND token_hash = ?"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(q, revokedAt, time.Now().UTC(), tokenHash)
	return err
}

func (r *refreshTokenRepo) DeleteExpired(before time.Time) error {
	q := "UPDATE refresh_tokens SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND expires_at <= ?"
	q = sqlutil.Rebind(r.dialect, q)

	now := time.Now().UTC()
	_, err := r.db.Exec(q, now, now, before)
	return err
}
