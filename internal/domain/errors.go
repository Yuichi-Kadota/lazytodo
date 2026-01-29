package domain

import "errors"

// Fatal errors - application should exit
var (
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrMigrationFailed    = errors.New("migration failed")
)

// Recoverable errors - can be handled gracefully
var (
	ErrNotFound          = errors.New("not found")
	ErrIntegrityViolation = errors.New("data integrity violation")
	ErrOrphanNode        = errors.New("orphan node detected")
	ErrCircularReference = errors.New("circular reference detected")
	ErrWriteFailed       = errors.New("write operation failed")
	ErrInvalidOperation  = errors.New("invalid operation")
)

// Warning errors - operation continues with defaults
var (
	ErrConfigNotFound    = errors.New("config file not found")
	ErrInvalidTheme      = errors.New("invalid theme configuration")
	ErrInvalidKeybinding = errors.New("invalid keybinding configuration")
)
