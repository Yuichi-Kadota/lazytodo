package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yuichikadota/lazytodo/internal/domain"
)

// TodoRepository implements domain.TodoRepository
type TodoRepository struct {
	db *DB
}

// NewTodoRepository creates a new todo repository
func NewTodoRepository(db *DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// Create creates a new todo
func (r *TodoRepository) Create(ctx context.Context, todo *domain.Todo) error {
	if todo.ID == "" {
		todo.ID = uuid.New().String()
	}
	todo.CreatedAt = time.Now()
	todo.UpdatedAt = todo.CreatedAt

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert todo
	_, err = tx.ExecContext(ctx, `
		INSERT INTO todos (id, workspace_id, description, position, status, urgency, due_date, created_at, updated_at, is_archived)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, todo.ID, todo.WorkspaceID, todo.Description, todo.Position, todo.Status, todo.Urgency,
		formatNullableTime(todo.DueDate), todo.CreatedAt.Format(time.RFC3339), todo.UpdatedAt.Format(time.RFC3339), todo.IsArchived)
	if err != nil {
		return fmt.Errorf("failed to insert todo: %w", err)
	}

	// Insert self-reference in closure table
	_, err = tx.ExecContext(ctx, `
		INSERT INTO todo_closure (ancestor_id, descendant_id, depth)
		VALUES (?, ?, 0)
	`, todo.ID, todo.ID)
	if err != nil {
		return fmt.Errorf("failed to insert self-reference: %w", err)
	}

	// If has parent, insert closure relationships
	if todo.ParentID != "" {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO todo_closure (ancestor_id, descendant_id, depth)
			SELECT ancestor_id, ?, depth + 1
			FROM todo_closure
			WHERE descendant_id = ?
		`, todo.ID, todo.ParentID)
		if err != nil {
			return fmt.Errorf("failed to insert closure relationships: %w", err)
		}
	}

	return tx.Commit()
}

// Update updates an existing todo
func (r *TodoRepository) Update(ctx context.Context, todo *domain.Todo) error {
	todo.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		UPDATE todos
		SET description = ?, position = ?, status = ?, urgency = ?, due_date = ?,
			updated_at = ?, completed_at = ?, is_archived = ?
		WHERE id = ? AND deleted_at IS NULL
	`, todo.Description, todo.Position, todo.Status, todo.Urgency,
		formatNullableTime(todo.DueDate), todo.UpdatedAt.Format(time.RFC3339),
		formatNullableTime(todo.CompletedAt), todo.IsArchived, todo.ID)
	if err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	return nil
}

// Delete soft-deletes a todo
func (r *TodoRepository) Delete(ctx context.Context, id string) error {
	now := time.Now().Format(time.RFC3339)

	// Soft delete todo and all descendants
	_, err := r.db.ExecContext(ctx, `
		UPDATE todos
		SET deleted_at = ?, updated_at = ?
		WHERE id IN (
			SELECT descendant_id FROM todo_closure WHERE ancestor_id = ?
		) AND deleted_at IS NULL
	`, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	return nil
}

// GetByID retrieves a todo by ID
func (r *TodoRepository) GetByID(ctx context.Context, id string) (*domain.Todo, error) {
	var t domain.Todo
	var createdAt, updatedAt string
	var dueDate, completedAt, deletedAt sql.NullString
	var parentID sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT t.id, t.workspace_id, t.description, t.position, t.status, t.urgency,
			   t.due_date, t.created_at, t.updated_at, t.completed_at, t.deleted_at, t.is_archived,
			   (SELECT MAX(depth) FROM todo_closure WHERE descendant_id = t.id) as depth,
			   (SELECT ancestor_id FROM todo_closure WHERE descendant_id = t.id AND depth = 1) as parent_id
		FROM todos t
		WHERE t.id = ?
	`, id).Scan(&t.ID, &t.WorkspaceID, &t.Description, &t.Position, &t.Status, &t.Urgency,
		&dueDate, &createdAt, &updatedAt, &completedAt, &deletedAt, &t.IsArchived, &t.Depth, &parentID)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	t.DueDate = parseNullableTime(dueDate)
	t.CompletedAt = parseNullableTime(completedAt)
	t.DeletedAt = parseNullableTime(deletedAt)
	if parentID.Valid {
		t.ParentID = parentID.String
	}

	return &t, nil
}

// GetByWorkspace retrieves all active todos in a workspace
func (r *TodoRepository) GetByWorkspace(ctx context.Context, workspaceID string, includeArchived bool) ([]*domain.Todo, error) {
	query := `
		SELECT t.id, t.workspace_id, t.description, t.position, t.status, t.urgency,
			   t.due_date, t.created_at, t.updated_at, t.completed_at, t.is_archived,
			   COALESCE((SELECT MAX(depth) FROM todo_closure WHERE descendant_id = t.id), 0) as depth,
			   (SELECT ancestor_id FROM todo_closure WHERE descendant_id = t.id AND depth = 1) as parent_id
		FROM todos t
		WHERE t.workspace_id = ? AND t.deleted_at IS NULL
	`
	if !includeArchived {
		query += " AND t.is_archived = 0"
	}
	query += " ORDER BY t.status, t.position, t.created_at"

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}
	defer rows.Close()

	return scanTodos(rows)
}

// GetChildren retrieves direct children of a todo
func (r *TodoRepository) GetChildren(ctx context.Context, parentID string) ([]*domain.Todo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.workspace_id, t.description, t.position, t.status, t.urgency,
			   t.due_date, t.created_at, t.updated_at, t.completed_at, t.is_archived
		FROM todos t
		JOIN todo_closure tc ON t.id = tc.descendant_id
		WHERE tc.ancestor_id = ? AND tc.depth = 1 AND t.deleted_at IS NULL
		ORDER BY t.status, t.position, t.created_at
	`, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	defer rows.Close()

	return scanTodosSimple(rows, parentID)
}

// GetDescendants retrieves all descendants of a todo
func (r *TodoRepository) GetDescendants(ctx context.Context, ancestorID string) ([]*domain.Todo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.workspace_id, t.description, t.position, t.status, t.urgency,
			   t.due_date, t.created_at, t.updated_at, t.completed_at, t.is_archived, tc.depth
		FROM todos t
		JOIN todo_closure tc ON t.id = tc.descendant_id
		WHERE tc.ancestor_id = ? AND tc.depth > 0 AND t.deleted_at IS NULL
		ORDER BY tc.depth, t.position, t.created_at
	`, ancestorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get descendants: %w", err)
	}
	defer rows.Close()

	var todos []*domain.Todo
	for rows.Next() {
		var t domain.Todo
		var createdAt, updatedAt string
		var dueDate, completedAt sql.NullString

		err := rows.Scan(&t.ID, &t.WorkspaceID, &t.Description, &t.Position, &t.Status, &t.Urgency,
			&dueDate, &createdAt, &updatedAt, &completedAt, &t.IsArchived, &t.Depth)
		if err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}

		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		t.DueDate = parseNullableTime(dueDate)
		t.CompletedAt = parseNullableTime(completedAt)

		todos = append(todos, &t)
	}

	return todos, nil
}

// Move moves a todo to a new parent or workspace
func (r *TodoRepository) Move(ctx context.Context, id string, newParentID string, newWorkspaceID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove old closure relationships (except self-reference)
	_, err = tx.ExecContext(ctx, `
		DELETE FROM todo_closure
		WHERE descendant_id IN (SELECT descendant_id FROM todo_closure WHERE ancestor_id = ?)
		  AND ancestor_id IN (SELECT ancestor_id FROM todo_closure WHERE descendant_id = ? AND depth > 0)
	`, id, id)
	if err != nil {
		return fmt.Errorf("failed to remove old closure relationships: %w", err)
	}

	// Add new closure relationships
	if newParentID != "" {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO todo_closure (ancestor_id, descendant_id, depth)
			SELECT p.ancestor_id, c.descendant_id, p.depth + c.depth + 1
			FROM todo_closure p
			CROSS JOIN todo_closure c
			WHERE p.descendant_id = ? AND c.ancestor_id = ?
		`, newParentID, id)
		if err != nil {
			return fmt.Errorf("failed to add new closure relationships: %w", err)
		}
	}

	// Update workspace_id if changed
	if newWorkspaceID != "" {
		_, err = tx.ExecContext(ctx, `
			UPDATE todos SET workspace_id = ?, updated_at = ?
			WHERE id IN (SELECT descendant_id FROM todo_closure WHERE ancestor_id = ?)
		`, newWorkspaceID, time.Now().Format(time.RFC3339), id)
		if err != nil {
			return fmt.Errorf("failed to update workspace_id: %w", err)
		}
	}

	return tx.Commit()
}

// Reorder changes the position of a todo among siblings
func (r *TodoRepository) Reorder(ctx context.Context, id string, newPosition int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE todos SET position = ?, updated_at = ? WHERE id = ?
	`, newPosition, time.Now().Format(time.RFC3339), id)
	return err
}

// Search searches todos by description
func (r *TodoRepository) Search(ctx context.Context, query string, includeArchived bool) ([]*domain.Todo, error) {
	sqlQuery := `
		SELECT t.id, t.workspace_id, t.description, t.position, t.status, t.urgency,
			   t.due_date, t.created_at, t.updated_at, t.completed_at, t.is_archived,
			   COALESCE((SELECT MAX(depth) FROM todo_closure WHERE descendant_id = t.id), 0) as depth,
			   (SELECT ancestor_id FROM todo_closure WHERE descendant_id = t.id AND depth = 1) as parent_id
		FROM todos t
		WHERE t.deleted_at IS NULL AND t.description LIKE ?
	`
	if !includeArchived {
		sqlQuery += " AND t.is_archived = 0"
	}
	sqlQuery += " ORDER BY t.created_at DESC"

	rows, err := r.db.QueryContext(ctx, sqlQuery, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search todos: %w", err)
	}
	defer rows.Close()

	return scanTodos(rows)
}

// Archive marks a todo as archived
func (r *TodoRepository) Archive(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE todos SET is_archived = 1, updated_at = ? WHERE id = ?
	`, time.Now().Format(time.RFC3339), id)
	return err
}

// GetCompletedBefore retrieves todos completed before a given time
func (r *TodoRepository) GetCompletedBefore(ctx context.Context, before time.Time) ([]*domain.Todo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.workspace_id, t.description, t.position, t.status, t.urgency,
			   t.due_date, t.created_at, t.updated_at, t.completed_at, t.is_archived,
			   COALESCE((SELECT MAX(depth) FROM todo_closure WHERE descendant_id = t.id), 0) as depth,
			   (SELECT ancestor_id FROM todo_closure WHERE descendant_id = t.id AND depth = 1) as parent_id
		FROM todos t
		WHERE t.deleted_at IS NULL AND t.status = 'completed'
			  AND t.completed_at IS NOT NULL AND t.completed_at < ?
			  AND t.is_archived = 0
		ORDER BY t.completed_at
	`, before.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to get completed todos: %w", err)
	}
	defer rows.Close()

	return scanTodos(rows)
}

// AutoArchive marks completed todos older than the specified duration as archived
func (r *TodoRepository) AutoArchive(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan).Format(time.RFC3339)

	result, err := r.db.ExecContext(ctx, `
		UPDATE todos
		SET is_archived = 1, updated_at = ?
		WHERE status = 'completed'
			AND completed_at IS NOT NULL
			AND completed_at < ?
			AND is_archived = 0
			AND deleted_at IS NULL
	`, time.Now().Format(time.RFC3339), cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to auto-archive todos: %w", err)
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// CheckAndRepairIntegrity verifies and repairs closure table integrity
func (r *TodoRepository) CheckAndRepairIntegrity(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Fix missing self-references
	_, err = tx.ExecContext(ctx, `
		INSERT OR IGNORE INTO todo_closure (ancestor_id, descendant_id, depth)
		SELECT id, id, 0 FROM todos WHERE deleted_at IS NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to fix missing self-references: %w", err)
	}

	// Remove orphaned closure entries (entries referencing deleted todos)
	_, err = tx.ExecContext(ctx, `
		DELETE FROM todo_closure
		WHERE ancestor_id NOT IN (SELECT id FROM todos WHERE deleted_at IS NULL)
		   OR descendant_id NOT IN (SELECT id FROM todos WHERE deleted_at IS NULL)
	`)
	if err != nil {
		return fmt.Errorf("failed to remove orphaned closure entries: %w", err)
	}

	return tx.Commit()
}

// Helper functions

func formatNullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339)
}

func parseNullableTime(ns sql.NullString) *time.Time {
	if !ns.Valid {
		return nil
	}
	t, err := time.Parse(time.RFC3339, ns.String)
	if err != nil {
		return nil
	}
	return &t
}

func scanTodos(rows *sql.Rows) ([]*domain.Todo, error) {
	var todos []*domain.Todo
	for rows.Next() {
		var t domain.Todo
		var createdAt, updatedAt string
		var dueDate, completedAt sql.NullString
		var parentID sql.NullString

		err := rows.Scan(&t.ID, &t.WorkspaceID, &t.Description, &t.Position, &t.Status, &t.Urgency,
			&dueDate, &createdAt, &updatedAt, &completedAt, &t.IsArchived, &t.Depth, &parentID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}

		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		t.DueDate = parseNullableTime(dueDate)
		t.CompletedAt = parseNullableTime(completedAt)
		if parentID.Valid {
			t.ParentID = parentID.String
		}

		todos = append(todos, &t)
	}

	return todos, nil
}

func scanTodosSimple(rows *sql.Rows, parentID string) ([]*domain.Todo, error) {
	var todos []*domain.Todo
	for rows.Next() {
		var t domain.Todo
		var createdAt, updatedAt string
		var dueDate, completedAt sql.NullString

		err := rows.Scan(&t.ID, &t.WorkspaceID, &t.Description, &t.Position, &t.Status, &t.Urgency,
			&dueDate, &createdAt, &updatedAt, &completedAt, &t.IsArchived)
		if err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}

		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		t.DueDate = parseNullableTime(dueDate)
		t.CompletedAt = parseNullableTime(completedAt)
		t.ParentID = parentID
		t.Depth = 1

		todos = append(todos, &t)
	}

	return todos, nil
}
