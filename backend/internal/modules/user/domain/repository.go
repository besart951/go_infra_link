package user

type UserEmailRepository interface {
	GetByEmail(email string) (*User, error)
}
