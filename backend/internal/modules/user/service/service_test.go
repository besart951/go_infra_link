package user

import (
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/modules/user/domain"
	"github.com/google/uuid"
)

type fakeUserRepo struct {
	users   map[uuid.UUID]*domainUser.User
	updated *domainUser.User
}

func (r *fakeUserRepo) Create(entity *domainUser.User) error {
	r.users[entity.ID] = entity
	return nil
}

func (r *fakeUserRepo) GetByIds(ids []uuid.UUID) ([]*domainUser.User, error) {
	items := make([]*domainUser.User, 0, len(ids))
	for _, id := range ids {
		if usr, ok := r.users[id]; ok {
			copy := *usr
			items = append(items, &copy)
		}
	}
	return items, nil
}

func (r *fakeUserRepo) Update(entity *domainUser.User) error {
	copy := *entity
	r.users[entity.ID] = &copy
	r.updated = &copy
	return nil
}

func (r *fakeUserRepo) DeleteByIds(ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.users, id)
	}
	return nil
}

func (r *fakeUserRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainUser.User], error) {
	return &domain.PaginatedList[domainUser.User]{}, nil
}

type fakePasswordHasher struct{}

func (fakePasswordHasher) Hash(plain string) (string, error) {
	return "hashed:" + plain, nil
}

func (fakePasswordHasher) Compare(hash, plain string) error {
	return nil
}

func TestUpdateCurrentUserOnlyChangesSelfServiceFields(t *testing.T) {
	userID := uuid.New()
	repo := &fakeUserRepo{
		users: map[uuid.UUID]*domainUser.User{
			userID: {
				Base:      domain.Base{ID: userID},
				FirstName: "Old",
				LastName:  "Name",
				Email:     "old@example.test",
				Password:  "old-hash",
				IsActive:  true,
				Role:      domainUser.RoleAdminFZAG,
			},
		},
	}
	svc := New(repo, fakePasswordHasher{})

	updated, err := svc.UpdateCurrentUser(userID, domainUser.CurrentUserUpdate{
		FirstName: "New",
		LastName:  "Person",
		Email:     "new@example.test",
		Password:  "new-password",
	})
	if err != nil {
		t.Fatalf("UpdateCurrentUser returned error: %v", err)
	}

	if updated.FirstName != "New" || updated.LastName != "Person" || updated.Email != "new@example.test" {
		t.Fatalf("profile fields were not updated: %+v", updated)
	}
	if updated.Password != "hashed:new-password" {
		t.Fatalf("password was not hashed before update: %q", updated.Password)
	}
	if updated.Role != domainUser.RoleAdminFZAG {
		t.Fatalf("role changed through current-user interface: %q", updated.Role)
	}
	if !updated.IsActive {
		t.Fatal("active status changed through current-user interface")
	}
}

func TestUpdateCurrentUserKeepsPasswordWhenOmitted(t *testing.T) {
	userID := uuid.New()
	repo := &fakeUserRepo{
		users: map[uuid.UUID]*domainUser.User{
			userID: {
				Base:      domain.Base{ID: userID},
				FirstName: "Old",
				LastName:  "Name",
				Email:     "old@example.test",
				Password:  "old-hash",
				IsActive:  true,
				Role:      domainUser.RoleEnterpreneur,
			},
		},
	}
	svc := New(repo, fakePasswordHasher{})

	updated, err := svc.UpdateCurrentUser(userID, domainUser.CurrentUserUpdate{FirstName: "New"})
	if err != nil {
		t.Fatalf("UpdateCurrentUser returned error: %v", err)
	}

	if updated.Password != "old-hash" {
		t.Fatalf("password changed even though it was omitted: %q", updated.Password)
	}
	if updated.FirstName != "New" {
		t.Fatalf("first name was not updated: %q", updated.FirstName)
	}
}
