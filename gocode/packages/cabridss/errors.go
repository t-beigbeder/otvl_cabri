package cabridss

import "errors"

var (
	// ErrPasswordRequired is returned when accessing encrypted content
	// without access to the user's master password
	ErrPasswordRequired = errors.New("password required to perform this action")
)
