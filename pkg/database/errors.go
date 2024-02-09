package database

import "github.com/pkg/errors"

// Error codes for database operations.
const (
	ErrCodeWrite  = "failed to write"
	ErrCodeRead   = "failed to read"
	ErrCodeDelete = "failed to delete"
)

// Predefined errors for database operations.
var (
	ErrWrite  = errors.New(ErrCodeWrite)
	ErrRead   = errors.New(ErrCodeRead)
	ErrDelete = errors.New(ErrCodeDelete)
)
