// Package wire provides dependency injection wiring for the application.
// It separates the construction of dependencies from business logic.
package wire

import (
	"database/sql"

	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	authrepo "github.com/besart951/go_infra_link/backend/internal/repository/auth"
	facilityrepo "github.com/besart951/go_infra_link/backend/internal/repository/facilitysql"
	projectrepo "github.com/besart951/go_infra_link/backend/internal/repository/project"
	teamrepo "github.com/besart951/go_infra_link/backend/internal/repository/team"
	userrepo "github.com/besart951/go_infra_link/backend/internal/repository/user"
)

// Repositories holds all repository instances.
type Repositories struct {
	Project      domainProject.ProjectRepository
	User         domainUser.UserRepository
	UserEmail    domainUser.UserEmailRepository
	RefreshToken domainAuth.RefreshTokenRepository
	LoginAttempt domainAuth.LoginAttemptRepository
	PasswordReset domainAuth.PasswordResetTokenRepository
	Team         domainTeam.TeamRepository
	TeamMember   domainTeam.TeamMemberRepository

	FacilityBuildings                domainFacility.BuildingRepository
	FacilitySystemTypes              domainFacility.SystemTypeRepository
	FacilitySystemParts              domainFacility.SystemPartRepository
	FacilitySpecifications           domainFacility.SpecificationStore
	FacilityApparats                 domainFacility.ApparatRepository
	FacilityControlCabinet           domainFacility.ControlCabinetRepository
	FacilityFieldDevices             domainFacility.FieldDeviceStore
	FacilitySPSControllers           domainFacility.SPSControllerRepository
	FacilitySPSControllerSystemTypes domainFacility.SPSControllerSystemTypeStore
	FacilityBacnetObjects            domainFacility.BacnetObjectStore
	FacilityObjectData               domainFacility.ObjectDataStore
	FacilityObjectDataBacnetObjects  domainFacility.ObjectDataBacnetObjectStore
}

// NewRepositories creates all repository instances from the database connection.
func NewRepositories(db *sql.DB, driver string) (*Repositories, error) {
	userRepo := userrepo.NewUserRepository(db, driver)
	userEmailRepo, ok := userRepo.(domainUser.UserEmailRepository)
	if !ok {
		return nil, ErrUserRepoMissingEmailLookup
	}

	return &Repositories{
		Project:      projectrepo.NewProjectRepository(db, driver),
		User:         userRepo,
		UserEmail:    userEmailRepo,
		RefreshToken: authrepo.NewRefreshTokenRepository(db, driver),
		LoginAttempt: authrepo.NewLoginAttemptRepository(db, driver),
		PasswordReset: authrepo.NewPasswordResetTokenRepository(db, driver),
		Team:         teamrepo.NewTeamRepository(db, driver),
		TeamMember:   teamrepo.NewTeamMemberRepository(db, driver),

		FacilityBuildings:                facilityrepo.NewBuildingRepository(db, driver),
		FacilitySystemTypes:              facilityrepo.NewSystemTypeRepository(db, driver),
		FacilitySystemParts:              facilityrepo.NewSystemPartRepository(db, driver),
		FacilitySpecifications:           facilityrepo.NewSpecificationRepository(db, driver),
		FacilityApparats:                 facilityrepo.NewApparatRepository(db, driver),
		FacilityControlCabinet:           facilityrepo.NewControlCabinetRepository(db, driver),
		FacilityFieldDevices:             facilityrepo.NewFieldDeviceRepository(db, driver),
		FacilitySPSControllers:           facilityrepo.NewSPSControllerRepository(db, driver),
		FacilitySPSControllerSystemTypes: facilityrepo.NewSPSControllerSystemTypeRepository(db, driver),
		FacilityBacnetObjects:            facilityrepo.NewBacnetObjectRepository(db, driver),
		FacilityObjectData:               facilityrepo.NewObjectDataRepository(db, driver),
		FacilityObjectDataBacnetObjects:  facilityrepo.NewObjectDataBacnetObjectRepository(db, driver),
	}, nil
}
