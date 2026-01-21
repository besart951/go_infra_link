package team

import (
	"database/sql"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
	"github.com/google/uuid"
)

type memberRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewTeamMemberRepository(db *sql.DB, driver string) domainTeam.TeamMemberRepository {
	return &memberRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *memberRepo) GetUserRole(teamID, userID uuid.UUID) (*domainTeam.MemberRole, error) {
	q := "SELECT role FROM team_members WHERE deleted_at IS NULL AND team_id = ? AND user_id = ? LIMIT 1"
	q = sqlutil.Rebind(r.dialect, q)

	var role string
	err := r.db.QueryRow(q, teamID, userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	rv := domainTeam.MemberRole(role)
	return &rv, nil
}

func (r *memberRepo) Upsert(member *domainTeam.TeamMember) error {
	now := time.Now().UTC()
	if err := member.Base.InitForCreate(now); err != nil {
		return err
	}
	if member.JoinedAt.IsZero() {
		member.JoinedAt = now
	}

	q := "UPDATE team_members SET updated_at = ?, deleted_at = NULL, role = ? WHERE team_id = ? AND user_id = ?"
	q = sqlutil.Rebind(r.dialect, q)
	res, err := r.db.Exec(q, now, member.Role, member.TeamID, member.UserID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}

	q = "INSERT INTO team_members (id, created_at, updated_at, deleted_at, team_id, user_id, role, joined_at) VALUES (?, ?, ?, NULL, ?, ?, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)
	_, err = r.db.Exec(q, member.ID, member.CreatedAt, member.UpdatedAt, member.TeamID, member.UserID, member.Role, member.JoinedAt)
	return err
}

func (r *memberRepo) Delete(teamID, userID uuid.UUID) error {
	now := time.Now().UTC()
	q := "UPDATE team_members SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND team_id = ? AND user_id = ?"
	q = sqlutil.Rebind(r.dialect, q)
	_, err := r.db.Exec(q, now, now, teamID, userID)
	return err
}

func (r *memberRepo) ListByTeam(teamID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainTeam.TeamMember], error) {
	page := params.Page
	limit := params.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	countQ := "SELECT COUNT(*) FROM team_members WHERE deleted_at IS NULL AND team_id = ?"
	countQ = sqlutil.Rebind(r.dialect, countQ)
	var total int64
	if err := r.db.QueryRow(countQ, teamID).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, team_id, user_id, role, joined_at FROM team_members WHERE deleted_at IS NULL AND team_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	dataQ = sqlutil.Rebind(r.dialect, dataQ)
	rows, err := r.db.Query(dataQ, teamID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainTeam.TeamMember, 0, limit)
	for rows.Next() {
		var m domainTeam.TeamMember
		var deletedAt sql.NullTime
		var role string
		if err := rows.Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt, &deletedAt, &m.TeamID, &m.UserID, &role, &m.JoinedAt); err != nil {
			return nil, err
		}
		m.Role = domainTeam.MemberRole(role)
		if deletedAt.Valid {
			v := deletedAt.Time
			m.DeletedAt = &v
		}
		items = append(items, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainTeam.TeamMember]{Items: items, Total: total, Page: page, TotalPages: domain.CalculateTotalPages(total, limit)}, nil
}
