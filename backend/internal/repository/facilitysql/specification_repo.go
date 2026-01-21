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

type specificationRepo struct {
	db      *sql.DB
	dialect sqlutil.Dialect
}

func NewSpecificationRepository(db *sql.DB, driver string) domainFacility.SpecificationRepository {
	return &specificationRepo{db: db, dialect: sqlutil.DialectFromDriver(driver)}
}

func (r *specificationRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.Specification, error) {
	if len(ids) == 0 {
		return []*domainFacility.Specification{}, nil
	}

	q := "SELECT id, created_at, updated_at, deleted_at, specification_supplier, specification_brand, specification_type, additional_info_motor_valve, additional_info_size, additional_information_installation_location, electrical_connection_ph, electrical_connection_acdc, electrical_connection_amperage, electrical_connection_power, electrical_connection_rotation " +
		"FROM specifications WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
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

	out := make([]*domainFacility.Specification, 0, len(ids))
	for rows.Next() {
		var s domainFacility.Specification
		var deletedAt sql.NullTime
		var supplier sql.NullString
		var brand sql.NullString
		var typ sql.NullString
		var motorValve sql.NullString
		var size sql.NullInt64
		var installLoc sql.NullString
		var ph sql.NullInt64
		var acdc sql.NullString
		var amperage sql.NullFloat64
		var power sql.NullFloat64
		var rotation sql.NullInt64

		if err := rows.Scan(
			&s.ID,
			&s.CreatedAt,
			&s.UpdatedAt,
			&deletedAt,
			&supplier,
			&brand,
			&typ,
			&motorValve,
			&size,
			&installLoc,
			&ph,
			&acdc,
			&amperage,
			&power,
			&rotation,
		); err != nil {
			return nil, err
		}

		if deletedAt.Valid {
			t := deletedAt.Time
			s.DeletedAt = &t
		}
		if supplier.Valid {
			v := supplier.String
			s.SpecificationSupplier = &v
		}
		if brand.Valid {
			v := brand.String
			s.SpecificationBrand = &v
		}
		if typ.Valid {
			v := typ.String
			s.SpecificationType = &v
		}
		if motorValve.Valid {
			v := motorValve.String
			s.AdditionalInfoMotorValve = &v
		}
		if size.Valid {
			v := int(size.Int64)
			s.AdditionalInfoSize = &v
		}
		if installLoc.Valid {
			v := installLoc.String
			s.AdditionalInformationInstallationLocation = &v
		}
		if ph.Valid {
			v := int(ph.Int64)
			s.ElectricalConnectionPH = &v
		}
		if acdc.Valid {
			v := acdc.String
			s.ElectricalConnectionACDC = &v
		}
		if amperage.Valid {
			v := amperage.Float64
			s.ElectricalConnectionAmperage = &v
		}
		if power.Valid {
			v := power.Float64
			s.ElectricalConnectionPower = &v
		}
		if rotation.Valid {
			v := int(rotation.Int64)
			s.ElectricalConnectionRotation = &v
		}

		out = append(out, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *specificationRepo) Create(entity *domainFacility.Specification) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	q := "INSERT INTO specifications (id, created_at, updated_at, deleted_at, specification_supplier, specification_brand, specification_type, additional_info_motor_valve, additional_info_size, additional_information_installation_location, electrical_connection_ph, electrical_connection_acdc, electrical_connection_amperage, electrical_connection_power, electrical_connection_rotation) " +
		"VALUES (?, ?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		entity.ID,
		entity.CreatedAt,
		entity.UpdatedAt,
		argStringPtr(entity.SpecificationSupplier),
		argStringPtr(entity.SpecificationBrand),
		argStringPtr(entity.SpecificationType),
		argStringPtr(entity.AdditionalInfoMotorValve),
		argIntPtr(entity.AdditionalInfoSize),
		argStringPtr(entity.AdditionalInformationInstallationLocation),
		argIntPtr(entity.ElectricalConnectionPH),
		argStringPtr(entity.ElectricalConnectionACDC),
		argFloatPtr(entity.ElectricalConnectionAmperage),
		argFloatPtr(entity.ElectricalConnectionPower),
		argIntPtr(entity.ElectricalConnectionRotation),
	)
	return err
}

func (r *specificationRepo) Update(entity *domainFacility.Specification) error {
	now := time.Now().UTC()
	entity.Base.TouchForUpdate(now)

	q := "UPDATE specifications SET updated_at = ?, specification_supplier = ?, specification_brand = ?, specification_type = ?, additional_info_motor_valve = ?, additional_info_size = ?, additional_information_installation_location = ?, electrical_connection_ph = ?, electrical_connection_acdc = ?, electrical_connection_amperage = ?, electrical_connection_power = ?, electrical_connection_rotation = ? " +
		"WHERE deleted_at IS NULL AND id = ?"
	q = sqlutil.Rebind(r.dialect, q)

	_, err := r.db.Exec(
		q,
		entity.UpdatedAt,
		argStringPtr(entity.SpecificationSupplier),
		argStringPtr(entity.SpecificationBrand),
		argStringPtr(entity.SpecificationType),
		argStringPtr(entity.AdditionalInfoMotorValve),
		argIntPtr(entity.AdditionalInfoSize),
		argStringPtr(entity.AdditionalInformationInstallationLocation),
		argIntPtr(entity.ElectricalConnectionPH),
		argStringPtr(entity.ElectricalConnectionACDC),
		argFloatPtr(entity.ElectricalConnectionAmperage),
		argFloatPtr(entity.ElectricalConnectionPower),
		argIntPtr(entity.ElectricalConnectionRotation),
		entity.ID,
	)
	return err
}

func (r *specificationRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	q := "UPDATE specifications SET deleted_at = ?, updated_at = ? WHERE deleted_at IS NULL AND id IN (" + sqlutil.Placeholders(len(ids)) + ")"
	q = sqlutil.Rebind(r.dialect, q)

	args := make([]any, 0, 2+len(ids))
	args = append(args, now, now)
	for _, id := range ids {
		args = append(args, id)
	}
	_, err := r.db.Exec(q, args...)
	return err
}

func (r *specificationRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Specification], error) {
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
		where += " AND ((specification_supplier " + like + " ?) OR (specification_brand " + like + " ?) OR (specification_type " + like + " ?))"
		args = append(args, pattern, pattern, pattern)
	}

	countQ := "SELECT COUNT(*) FROM specifications WHERE " + where
	countQ = sqlutil.Rebind(r.dialect, countQ)
	var total int64
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	dataQ := "SELECT id, created_at, updated_at, deleted_at, specification_supplier, specification_brand, specification_type, additional_info_motor_valve, additional_info_size, additional_information_installation_location, electrical_connection_ph, electrical_connection_acdc, electrical_connection_amperage, electrical_connection_power, electrical_connection_rotation " +
		"FROM specifications WHERE " + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	dataQ = sqlutil.Rebind(r.dialect, dataQ)
	dataArgs := append(append([]any{}, args...), limit, offset)

	rows, err := r.db.Query(dataQ, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]domainFacility.Specification, 0, limit)
	for rows.Next() {
		var s domainFacility.Specification
		var deletedAt sql.NullTime
		var supplier sql.NullString
		var brand sql.NullString
		var typ sql.NullString
		var motorValve sql.NullString
		var size sql.NullInt64
		var installLoc sql.NullString
		var ph sql.NullInt64
		var acdc sql.NullString
		var amperage sql.NullFloat64
		var power sql.NullFloat64
		var rotation sql.NullInt64
		if err := rows.Scan(
			&s.ID,
			&s.CreatedAt,
			&s.UpdatedAt,
			&deletedAt,
			&supplier,
			&brand,
			&typ,
			&motorValve,
			&size,
			&installLoc,
			&ph,
			&acdc,
			&amperage,
			&power,
			&rotation,
		); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			t := deletedAt.Time
			s.DeletedAt = &t
		}
		if supplier.Valid {
			v := supplier.String
			s.SpecificationSupplier = &v
		}
		if brand.Valid {
			v := brand.String
			s.SpecificationBrand = &v
		}
		if typ.Valid {
			v := typ.String
			s.SpecificationType = &v
		}
		if motorValve.Valid {
			v := motorValve.String
			s.AdditionalInfoMotorValve = &v
		}
		if size.Valid {
			v := int(size.Int64)
			s.AdditionalInfoSize = &v
		}
		if installLoc.Valid {
			v := installLoc.String
			s.AdditionalInformationInstallationLocation = &v
		}
		if ph.Valid {
			v := int(ph.Int64)
			s.ElectricalConnectionPH = &v
		}
		if acdc.Valid {
			v := acdc.String
			s.ElectricalConnectionACDC = &v
		}
		if amperage.Valid {
			v := amperage.Float64
			s.ElectricalConnectionAmperage = &v
		}
		if power.Valid {
			v := power.Float64
			s.ElectricalConnectionPower = &v
		}
		if rotation.Valid {
			v := int(rotation.Int64)
			s.ElectricalConnectionRotation = &v
		}

		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.Specification]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
