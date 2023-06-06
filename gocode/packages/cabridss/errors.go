package cabridss

import (
	"errors"
	"fmt"
)

// ErrBadParameter is returned when REST API receives a bad parameter
type ErrBadParameter struct {
	Key   string
	Value fmt.Stringer
	Err   error
}

func (e *ErrBadParameter) Error() string {
	return fmt.Sprintf("bad parameter %s %s: %s", e.Key, e.Value, e.Err)
}

func (e *ErrBadParameter) Unwrap() error { return e.Err }

var (
	// ErrPasswordRequired is returned when accessing encrypted content
	// without access to the user's master password
	ErrPasswordRequired = errors.New("password required to perform this action")
)
