package wire

import "errors"

// ErrUserRepoMissingEmailLookup is returned when the user repository
// does not implement the UserEmailRepository interface.
var ErrUserRepoMissingEmailLookup = errors.New("user repository does not implement email lookup")
