package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yuichikadota/lazytodo/internal/domain"
)

// WorkspaceRepository implements domain.WorkspaceRepository
type WorkspaceRepository struct {
	db *DB
}

// NewWorkspaceRepository creates a new workspace repository
func NewWorkspaceRepository(db *DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

// Create creates a new workspace
func (r *WorkspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) error {
	if workspace.ID == "" {
		workspace.ID = uuid.New().String()
	}
	workspace.CreatedAt = time.Now()
	workspace.UpdatedAt = workspace.CreatedAt

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert workspace
	_, err = tx.ExecContext(ctx, `
		INSERT INTO workspaces (id, name, position, is_expanded, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, workspace.ID, workspace.Name, workspace.Position, workspace.IsExpanded,
		workspace.CreatedAt.Format(time.RFC3339), workspace.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to insert workspace: %w", err)
	}

	// Insert self-reference in closure table
	_, err = tx.ExecContext(ctx, `
		INSERT INTO workspace_closure (ancestor_id, descendant_id, depth)
		VALUES (?, ?, 0)
	`, workspace.ID, workspace.ID)
	if err != nil {
		return fmt.Errorf("failed to insert self-reference: %w", err)
	}

	// If has parent, insert closure relationships
	if workspace.ParentID != "" {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO workspace_closure (ancestor_id, descendant_id, depth)
			SELECT ancestor_id, ?, depth + 1
			FROM workspace_closure
			WHERE descendant_id = ?
		`, workspace.ID, workspace.ParentID)
		if err != nil {
			return fmt.Errorf("failed to insert closure relationships: %w", err)
		}
	}

	return tx.Commit()
}

// Update updates an existing workspace
func (r *WorkspaceRepository) Update(ctx context.Context, workspace *domain.Workspace) error {
	workspace.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		UPDATE workspaces
		SET name = ?, position = ?, is_expanded = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, workspace.Name, workspace.Position, workspace.IsExpanded,
		workspace.UpdatedAt.Format(time.RFC3339), workspace.ID)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	return nil
}

// Delete soft-deletes a workspace
func (r *WorkspaceRepository) Delete(ctx context.Context, id string) error {
	now := time.Now().Format(time.RFC3339)

	// Soft delete workspace and all descendants
	_, err := r.db.ExecContext(ctx, `
		UPDATE workspaces
		SET deleted_at = ?, updated_at = ?
		WHERE id IN (
			SELECT descendant_id FROM workspace_closure WHERE ancestor_id = ?
		) AND deleted_at IS NULL
	`, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	return nil
}

// GetByID retrieves a workspace by ID
func (r *WorkspaceRepository) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	var w domain.Workspace
	var createdAt, updatedAt string
	var deletedAt sql.NullString
	var parentID sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT w.id, w.name, w.position, w.is_expanded, w.created_at, w.updated_at, w.deleted_at,
			   (SELECT MAX(depth) FROM workspace_closure WHERE descendant_id = w.id) as depth,
			   (SELECT ancestor_id FROM workspace_closure WHERE descendant_id = w.id AND depth = 1) as parent_id
		FROM workspaces w
		WHERE w.id = ?
	`, id).Scan(&w.ID, &w.Name, &w.Position, &w.IsExpanded, &createdAt, &updatedAt, &deletedAt, &w.Depth, &parentID)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	w.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	w.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	if deletedAt.Valid {
		t, _ := time.Parse(time.RFC3339, deletedAt.String)
		w.DeletedAt = &t
	}
	if parentID.Valid {
		w.ParentID = parentID.String
	}

	return &w, nil
}

// GetAll retrieves all active workspaces
func (r *WorkspaceRepository) GetAll(ctx context.Context) ([]*domain.Workspace, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT w.id, w.name, w.position, w.is_expanded, w.created_at, w.updated_at,
			   COALESCE((SELECT MAX(depth) FROM workspace_closure WHERE descendant_id = w.id), 0) as depth,
			   (SELECT ancestor_id FROM workspace_closure WHERE descendant_id = w.id AND depth = 1) as parent_id
		FROM workspaces w
		WHERE w.deleted_at IS NULL
		ORDER BY w.position, w.name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var w domain.Workspace
		var createdAt, updatedAt string
		var parentID sql.NullString

		err := rows.Scan(&w.ID, &w.Name, &w.Position, &w.IsExpanded, &createdAt, &updatedAt, &w.Depth, &parentID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}

		w.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		w.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		if parentID.Valid {
			w.ParentID = parentID.String
		}

		workspaces = append(workspaces, &w)
	}

	return workspaces, nil
}

// GetChildren retrieves direct children of a workspace
func (r *WorkspaceRepository) GetChildren(ctx context.Context, parentID string) ([]*domain.Workspace, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT w.id, w.name, w.position, w.is_expanded, w.created_at, w.updated_at
		FROM workspaces w
		JOIN workspace_closure wc ON w.id = wc.descendant_id
		WHERE wc.ancestor_id = ? AND wc.depth = 1 AND w.deleted_at IS NULL
		ORDER BY w.position, w.name
	`, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var w domain.Workspace
		var createdAt, updatedAt string

		err := rows.Scan(&w.ID, &w.Name, &w.Position, &w.IsExpanded, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}

		w.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		w.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		w.ParentID = parentID
		w.Depth = 1

		workspaces = append(workspaces, &w)
	}

	return workspaces, nil
}

// GetDescendants retrieves all descendants of a workspace
func (r *WorkspaceRepository) GetDescendants(ctx context.Context, ancestorID string) ([]*domain.Workspace, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT w.id, w.name, w.position, w.is_expanded, w.created_at, w.updated_at, wc.depth
		FROM workspaces w
		JOIN workspace_closure wc ON w.id = wc.descendant_id
		WHERE wc.ancestor_id = ? AND wc.depth > 0 AND w.deleted_at IS NULL
		ORDER BY wc.depth, w.position, w.name
	`, ancestorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get descendants: %w", err)
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var w domain.Workspace
		var createdAt, updatedAt string

		err := rows.Scan(&w.ID, &w.Name, &w.Position, &w.IsExpanded, &createdAt, &updatedAt, &w.Depth)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}

		w.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		w.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		workspaces = append(workspaces, &w)
	}

	return workspaces, nil
}

// GetAncestors retrieves all ancestors of a workspace
func (r *WorkspaceRepository) GetAncestors(ctx context.Context, descendantID string) ([]*domain.Workspace, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT w.id, w.name, w.position, w.is_expanded, w.created_at, w.updated_at, wc.depth
		FROM workspaces w
		JOIN workspace_closure wc ON w.id = wc.ancestor_id
		WHERE wc.descendant_id = ? AND wc.depth > 0 AND w.deleted_at IS NULL
		ORDER BY wc.depth DESC
	`, descendantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestors: %w", err)
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var w domain.Workspace
		var createdAt, updatedAt string

		err := rows.Scan(&w.ID, &w.Name, &w.Position, &w.IsExpanded, &createdAt, &updatedAt, &w.Depth)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}

		w.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		w.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		workspaces = append(workspaces, &w)
	}

	return workspaces, nil
}

// Move moves a workspace to a new parent
func (r *WorkspaceRepository) Move(ctx context.Context, id string, newParentID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove old closure relationships (except self-reference)
	_, err = tx.ExecContext(ctx, `
		DELETE FROM workspace_closure
		WHERE descendant_id IN (SELECT descendant_id FROM workspace_closure WHERE ancestor_id = ?)
		  AND ancestor_id IN (SELECT ancestor_id FROM workspace_closure WHERE descendant_id = ? AND depth > 0)
	`, id, id)
	if err != nil {
		return fmt.Errorf("failed to remove old closure relationships: %w", err)
	}

	// Add new closure relationships
	if newParentID != "" {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO workspace_closure (ancestor_id, descendant_id, depth)
			SELECT p.ancestor_id, c.descendant_id, p.depth + c.depth + 1
			FROM workspace_closure p
			CROSS JOIN workspace_closure c
			WHERE p.descendant_id = ? AND c.ancestor_id = ?
		`, newParentID, id)
		if err != nil {
			return fmt.Errorf("failed to add new closure relationships: %w", err)
		}
	}

	// Update timestamp
	_, err = tx.ExecContext(ctx, `
		UPDATE workspaces SET updated_at = ? WHERE id = ?
	`, time.Now().Format(time.RFC3339), id)
	if err != nil {
		return fmt.Errorf("failed to update timestamp: %w", err)
	}

	return tx.Commit()
}

// Reorder changes the position of a workspace among siblings
func (r *WorkspaceRepository) Reorder(ctx context.Context, id string, newPosition int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE workspaces SET position = ?, updated_at = ? WHERE id = ?
	`, newPosition, time.Now().Format(time.RFC3339), id)
	return err
}

// CheckAndRepairIntegrity verifies and repairs closure table integrity
func (r *WorkspaceRepository) CheckAndRepairIntegrity(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Fix missing self-references
	_, err = tx.ExecContext(ctx, `
		INSERT OR IGNORE INTO workspace_closure (ancestor_id, descendant_id, depth)
		SELECT id, id, 0 FROM workspaces WHERE deleted_at IS NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to fix missing self-references: %w", err)
	}

	// Remove orphaned closure entries (entries referencing deleted workspaces)
	_, err = tx.ExecContext(ctx, `
		DELETE FROM workspace_closure
		WHERE ancestor_id NOT IN (SELECT id FROM workspaces WHERE deleted_at IS NULL)
		   OR descendant_id NOT IN (SELECT id FROM workspaces WHERE deleted_at IS NULL)
	`)
	if err != nil {
		return fmt.Errorf("failed to remove orphaned closure entries: %w", err)
	}

	return tx.Commit()
}

// GetOrCreateArchive ensures the _archive workspace exists
func (r *WorkspaceRepository) GetOrCreateArchive(ctx context.Context) (*domain.Workspace, error) {
	// Try to get existing _archive workspace
	var w domain.Workspace
	var createdAt, updatedAt string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, position, is_expanded, created_at, updated_at
		FROM workspaces
		WHERE name = '_archive' AND deleted_at IS NULL
	`).Scan(&w.ID, &w.Name, &w.Position, &w.IsExpanded, &createdAt, &updatedAt)

	if err == nil {
		w.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		w.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		return &w, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get _archive workspace: %w", err)
	}

	// Create _archive workspace
	archive := &domain.Workspace{
		Name:       "_archive",
		Position:   999999, // Always at the end
		IsExpanded: false,
	}

	if err := r.Create(ctx, archive); err != nil {
		return nil, fmt.Errorf("failed to create _archive workspace: %w", err)
	}

	return archive, nil
}
