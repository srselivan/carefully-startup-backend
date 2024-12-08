package repo

import "errors"

var (
	ErrNothingUpdated = errors.New("nothing updated")
	ErrNotFound       = errors.New("not found")
)
