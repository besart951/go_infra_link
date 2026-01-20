package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Hash(plain string) (string, error)
	Compare(hash string, plain string) error
}

type bcryptService struct{}

func New() Service {
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
