package incidents

import "errors"

var (
	ErrNotFound     = errors.New("incident not found")
	ErrInvalidState = errors.New("invalid incident state")
)
