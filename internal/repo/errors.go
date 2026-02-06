package repo

import "errors"

var (
	ErrNotFound = errors.New("repo: requested entity was not found")
)
