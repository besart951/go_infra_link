package user

import "errors"

var ErrPasswordHashingFailed = errors.New("password_hashing_failed")
var ErrForbiddenUserDirectory = errors.New("forbidden_user_directory")
