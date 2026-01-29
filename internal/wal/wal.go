package wal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

const (
	defaultDebounceInterval = 100 * time.Millisecond
	defaultMaxOperations    = 100
)

// WAL manages the Write-Ahead Log
type WAL struct {
	db              *sql.DB
	mu              sync.Mutex
	pending         []*Operation
	debounceTimer   *time.Timer
	debounceInterval time.Duration
	applyFunc       func([]*Operation) error
}

// Config holds WAL configuration
type Config struct {
	DebounceInterval time.Duration
	ApplyFunc        func([]*Operation) error
}

// New creates a new WAL instance
func New(db *sql.DB, cfg Config) *WAL {
	interval := cfg.DebounceInterval
	if interval == 0 {
		interval = defaultDebounceInterval
	}

	return &WAL{
		db:               db,
		debounceInterval: interval,
		applyFunc:        cfg.ApplyFunc,
	}
}

// Append adds a new operation to the WAL
func (w *WAL) Append(op *Operation) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Marshal payload
	payloadJSON, err := json.Marshal(op.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Insert into operation_log
	result, err := w.db.Exec(`
		INSERT INTO operation_log (operation_type, entity_type, entity_id, payload, undo_group_id)
		VALUES (?, ?, ?, ?, ?)
	`, op.OperationType, op.EntityType, op.EntityID, string(payloadJSON), op.UndoGroupID)
	if err != nil {
		return fmt.Errorf("failed to insert operation: %w", err)
	}

	id, _ := result.LastInsertId()
	op.ID = id
	op.CreatedAt = time.Now()

	// Add to pending list
	w.pending = append(w.pending, op)

	// Reset debounce timer
	if w.debounceTimer != nil {
		w.debounceTimer.Stop()
	}
	w.debounceTimer = time.AfterFunc(w.debounceInterval, func() {
		w.flush()
	})

	return nil
}

// flush applies pending operations to main tables
func (w *WAL) flush() {
	w.mu.Lock()
	pending := w.pending
	w.pending = nil
	w.mu.Unlock()

	if len(pending) == 0 {
		return
	}

	// Apply operations
	if w.applyFunc != nil {
		if err := w.applyFunc(pending); err != nil {
			// Log error but don't fail - operations are already in WAL
			fmt.Printf("WAL apply error: %v\n", err)
			return
		}
	}

	// Mark as applied
	for _, op := range pending {
		w.db.Exec(`UPDATE operation_log SET applied = 1 WHERE id = ?`, op.ID)
	}
}

// Flush forces immediate flush of pending operations
func (w *WAL) Flush() {
	if w.debounceTimer != nil {
		w.debounceTimer.Stop()
	}
	w.flush()
}

// Recovery replays unapplied operations on startup
func (w *WAL) Recovery() ([]*Operation, error) {
	rows, err := w.db.Query(`
		SELECT id, operation_type, entity_type, entity_id, payload, is_undone, undo_group_id, created_at
		FROM operation_log
		WHERE applied = 0
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query unapplied operations: %w", err)
	}
	defer rows.Close()

	var ops []*Operation
	for rows.Next() {
		var op Operation
		var payloadStr string
		var createdAtStr string
		var undoGroupID sql.NullString

		err := rows.Scan(
			&op.ID,
			&op.OperationType,
			&op.EntityType,
			&op.EntityID,
			&payloadStr,
			&op.IsUndone,
			&undoGroupID,
			&createdAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan operation: %w", err)
		}

		if err := json.Unmarshal([]byte(payloadStr), &op.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		if undoGroupID.Valid {
			op.UndoGroupID = undoGroupID.String
		}

		op.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		ops = append(ops, &op)
	}

	return ops, nil
}

// GetUndoOperations returns operations that can be undone
func (w *WAL) GetUndoOperations(limit int) ([]*Operation, error) {
	if limit <= 0 {
		limit = defaultMaxOperations
	}

	rows, err := w.db.Query(`
		SELECT id, operation_type, entity_type, entity_id, payload, undo_group_id, created_at
		FROM operation_log
		WHERE applied = 1 AND is_undone = 0
		ORDER BY id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query undo operations: %w", err)
	}
	defer rows.Close()

	var ops []*Operation
	for rows.Next() {
		var op Operation
		var payloadStr string
		var createdAtStr string
		var undoGroupID sql.NullString

		err := rows.Scan(
			&op.ID,
			&op.OperationType,
			&op.EntityType,
			&op.EntityID,
			&payloadStr,
			&undoGroupID,
			&createdAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan operation: %w", err)
		}

		if err := json.Unmarshal([]byte(payloadStr), &op.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		if undoGroupID.Valid {
			op.UndoGroupID = undoGroupID.String
		}

		op.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		ops = append(ops, &op)
	}

	return ops, nil
}

// MarkUndone marks an operation as undone
func (w *WAL) MarkUndone(opID int64) error {
	_, err := w.db.Exec(`UPDATE operation_log SET is_undone = 1 WHERE id = ?`, opID)
	return err
}

// MarkRedone marks an operation as redone (not undone)
func (w *WAL) MarkRedone(opID int64) error {
	_, err := w.db.Exec(`UPDATE operation_log SET is_undone = 0 WHERE id = ?`, opID)
	return err
}

// Cleanup removes old applied operations
func (w *WAL) Cleanup(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan).Format("2006-01-02 15:04:05")
	_, err := w.db.Exec(`
		DELETE FROM operation_log
		WHERE applied = 1 AND is_undone = 1 AND created_at < ?
	`, cutoff)
	return err
}
