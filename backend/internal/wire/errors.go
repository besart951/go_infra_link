package wire

import "errors"

var (
	ErrUserRepoMissingEmailLookup = errors.New("user repository does not implement user email lookup")
	ErrUserRepoMissingLifecycle   = errors.New("user repository does not implement user lifecycle store")
)
