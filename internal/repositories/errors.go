package repositories

import "errors"

var (
	ErrNotFound = errors.New("repo: requested entity was not found")
)
