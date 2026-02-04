package password

import (
	"errors"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type bcryptService struct{}

// New creates a new password hasher backed by bcrypt.
func New() domainUser.PasswordHasher {
	return &bcryptService{}
}

func (s *bcryptService) Hash(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("empty_password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func (s *bcryptService) Compare(hash string, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
