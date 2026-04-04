package user

import "context"

type UserEmailRepository interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
}
