package wal

import (
	"fmt"
)

// RecoveryResult contains the result of recovery process
type RecoveryResult struct {
	RecoveredOps int
	Errors       []error
}

// RunRecovery performs crash recovery on startup
func (w *WAL) RunRecovery() (*RecoveryResult, error) {
	result := &RecoveryResult{}

	// Get unapplied operations
	ops, err := w.Recovery()
	if err != nil {
		return nil, fmt.Errorf("failed to get unapplied operations: %w", err)
	}

	if len(ops) == 0 {
		return result, nil
	}

	// Apply each operation
	if w.applyFunc != nil {
		if err := w.applyFunc(ops); err != nil {
			result.Errors = append(result.Errors, err)
		} else {
			result.RecoveredOps = len(ops)
		}
	}

	// Mark all as applied
	for _, op := range ops {
		if _, err := w.db.Exec(`UPDATE operation_log SET applied = 1 WHERE id = ?`, op.ID); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to mark op %d as applied: %w", op.ID, err))
		}
	}

	return result, nil
}
