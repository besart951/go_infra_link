package facilitysql

import (
	"database/sql"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/sqlutil"
	"github.com/google/uuid"
)

type spsControllerRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewSPSControllerRepository(db *sql.DB, driver string) domainFacility.SPSControllerRepository {
	return &spsControllerRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *spsControllerRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.SPSController, error) {
	if len(ids) == 0 {
		return []*domainFacility.SPSController{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, control_cabinet_id, project_id, ga_device, device_name, device_description, device_location, ip_address, subnet, gateway, vlan " +
		"FROM sps_controllers WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
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

	out := make([]*domainFacility.SPSController, 0, len(ids))
	for rows.Next() {
		var s domainFacility.SPSController
		var deletedAt sql.NullTime
		var projectID sql.NullString
		var ga sql.NullString
		var desc sql.NullString
		var loc sql.NullString
		var ip sql.NullString
		var subnet sql.NullString
		var gateway sql.NullString
		var vlan sql.NullString
		if err := rows.Scan(
			&s.ID,
			&s.CreatedAt,
			&s.UpdatedAt,
			&deletedAt,
			&s.ControlCabinetID,
			&projectID,
			&ga,
			&s.DeviceName,
			&desc,
			&loc,
			&ip,
			&subnet,
			&gateway,
			&vlan,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			s.DeletedAt = &t
		}
		if projectID.Valid {
			id, err := uuid.Parse(projectID.String)
			if err != nil {
				return nil, err
			}
			s.ProjectID = &id
		}
		if ga.Valid {
			v := ga.String
			s.GADevice = &v
		}
		if desc.Valid {
			v := desc.String
			s.DeviceDescription = &v
		}
		if loc.Valid {
			v := loc.String
			s.DeviceLocation = &v
		}
		if ip.Valid {
			v := ip.String
			s.IPAddress = &v
		}
		if subnet.Valid {
			v := subnet.String
			s.Subnet = &v
		}
		if gateway.Valid {
			v := gateway.String
			s.Gateway = &v
		}
		if vlan.Valid {
			v := vlan.String
			s.Vlan = &v
		}
		out = append(out, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *spsControllerRepo) Create(entity *domainFacility.SPSController) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO sps_controllers (id, created_at, updated_at, deleted_at, control_cabinet_id, project_id, ga_device, device_name, device_description, device_location, ip_address, subnet, gateway, vlan) " +
		"VALUES (?, ?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		entity.ID,
		entity.CreatedAt,
		entity.UpdatedAt,
		entity.ControlCabinetID,
		argUUIDPtr(entity.ProjectID),
		argStringPtr(entity.GADevice),
		entity.DeviceName,
		argStringPtr(entity.DeviceDescription),
		argStringPtr(entity.DeviceLocation),
		argStringPtr(entity.IPAddress),
		argStringPtr(entity.Subnet),
		argStringPtr(entity.Gateway),
		argStringPtr(entity.Vlan),
	)
	return err
}

func (r *spsControllerRepo) Update(entity *domainFacility.SPSController) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	q := "UPDATE sps_controllers SET updated_at = ?, control_cabinet_id = ?, project_id = ?, ga_device = ?, device_name = ?, device_description = ?, device_location = ?, ip_address = ?, subnet = ?, gateway = ?, vlan = ? " +
		"WHERE deleted_at IS NULL AND id = ?"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		entity.UpdatedAt,
		entity.ControlCabinetID,
		argUUIDPtr(entity.ProjectID),
		argStringPtr(entity.GADevice),
		entity.DeviceName,
		argStringPtr(entity.DeviceDescription),
		argStringPtr(entity.DeviceLocation),
		argStringPtr(entity.IPAddress),
		argStringPtr(entity.Subnet),
		argStringPtr(entity.Gateway),
		argStringPtr(entity.Vlan),
		entity.ID,
	)
	return err
}

func (r *spsControllerRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE sps_controllers SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}
	_, err := r.db.Exec(q, args...)
	return err
}

func (r *spsControllerRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSController], error) {
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
	args := make([]any, 0, 8)
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + params.Search + "%"
		like := sqlutil.LikeOperator(r.dialect)
		where += " AND ((device_name " + like + " ?) OR (ip_address " + like + " ?))"
		args = append(args, pattern, pattern)
	}

	countQ := "SELECT COUNT(*) FROM sps_controllers WHERE " + where
	countQ = sqlutil.Rebind(r.dialect, countQ)
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, control_cabinet_id, project_id, ga_device, device_name, device_description, device_location, ip_address, subnet, gateway, vlan " +
		"FROM sps_controllers WHERE " + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	dataQ = sqlutil.Rebind(r.dialect, dataQ)
	dataArgs := append(append([]any{}, args...), limit, offset)

	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainFacility.SPSController, 0, limit)
	for rows.Next() {
		var s domainFacility.SPSController
		var deletedAt sql.NullTime
		var projectID sql.NullString
		var ga sql.NullString
		var desc sql.NullString
		var loc sql.NullString
		var ip sql.NullString
		var subnet sql.NullString
		var gateway sql.NullString
		var vlan sql.NullString
		if err := rows.Scan(
			&s.ID,
			&s.CreatedAt,
			&s.UpdatedAt,
			&deletedAt,
			&s.ControlCabinetID,
			&projectID,
			&ga,
			&s.DeviceName,
			&desc,
			&loc,
			&ip,
			&subnet,
			&gateway,
			&vlan,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			s.DeletedAt = &t
		}
		if projectID.Valid {
			id, err := uuid.Parse(projectID.String)
			if err != nil {
				return nil, err
			}
			s.ProjectID = &id
		}
		if ga.Valid {
			v := ga.String
			s.GADevice = &v
		}
		if desc.Valid {
			v := desc.String
			s.DeviceDescription = &v
		}
		if loc.Valid {
			v := loc.String
			s.DeviceLocation = &v
		}
		if ip.Valid {
			v := ip.String
			s.IPAddress = &v
		}
		if subnet.Valid {
			v := subnet.String
			s.Subnet = &v
		}
		if gateway.Valid {
			v := gateway.String
			s.Gateway = &v
		}
		if vlan.Valid {
			v := vlan.String
			s.Vlan = &v
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.SPSController]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
