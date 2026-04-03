package notification

import "errors"

var ErrProviderNotConfigured = errors.New("notification provider not configured")
var ErrProviderDisabled = errors.New("notification provider disabled")
