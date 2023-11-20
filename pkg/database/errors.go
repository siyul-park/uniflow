package database

import "github.com/pkg/errors"

var (
	ErrCodeWrite  = "failed to write"
	ErrCodeRead   = "failed to read"
	ErrCodeDelete = "failed to delete"

	ErrWrite  = errors.New(ErrCodeWrite)
	ErrRead   = errors.New(ErrCodeRead)
	ErrDelete = errors.New(ErrCodeDelete)
)
