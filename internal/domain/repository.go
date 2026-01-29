package domain

import (
	"context"
	"time"
)

// WorkspaceRepository defines the interface for workspace persistence
type WorkspaceRepository interface {
	// Create creates a new workspace
	Create(ctx context.Context, workspace *Workspace) error

	// Update updates an existing workspace
	Update(ctx context.Context, workspace *Workspace) error

	// Delete soft-deletes a workspace
	Delete(ctx context.Context, id string) error

	// GetByID retrieves a workspace by ID
	GetByID(ctx context.Context, id string) (*Workspace, error)

	// GetAll retrieves all active workspaces
	GetAll(ctx context.Context) ([]*Workspace, error)

	// GetChildren retrieves direct children of a workspace
	GetChildren(ctx context.Context, parentID string) ([]*Workspace, error)

	// GetDescendants retrieves all descendants of a workspace
	GetDescendants(ctx context.Context, ancestorID string) ([]*Workspace, error)

	// GetAncestors retrieves all ancestors of a workspace
	GetAncestors(ctx context.Context, descendantID string) ([]*Workspace, error)

	// Move moves a workspace to a new parent
	Move(ctx context.Context, id string, newParentID string) error

	// Reorder changes the position of a workspace among siblings
	Reorder(ctx context.Context, id string, newPosition int) error

	// GetOrCreateArchive ensures the _archive workspace exists
	GetOrCreateArchive(ctx context.Context) (*Workspace, error)
}

// TodoRepository defines the interface for todo persistence
type TodoRepository interface {
	// Create creates a new todo
	Create(ctx context.Context, todo *Todo) error

	// Update updates an existing todo
	Update(ctx context.Context, todo *Todo) error

	// Delete soft-deletes a todo
	Delete(ctx context.Context, id string) error

	// GetByID retrieves a todo by ID
	GetByID(ctx context.Context, id string) (*Todo, error)

	// GetByWorkspace retrieves all active todos in a workspace
	GetByWorkspace(ctx context.Context, workspaceID string, includeArchived bool) ([]*Todo, error)

	// GetChildren retrieves direct children of a todo
	GetChildren(ctx context.Context, parentID string) ([]*Todo, error)

	// GetDescendants retrieves all descendants of a todo
	GetDescendants(ctx context.Context, ancestorID string) ([]*Todo, error)

	// Move moves a todo to a new parent or workspace
	Move(ctx context.Context, id string, newParentID string, newWorkspaceID string) error

	// Reorder changes the position of a todo among siblings
	Reorder(ctx context.Context, id string, newPosition int) error

	// Search searches todos by description
	Search(ctx context.Context, query string, includeArchived bool) ([]*Todo, error)

	// Archive marks a todo as archived
	Archive(ctx context.Context, id string) error

	// GetCompletedBefore retrieves todos completed before a given time
	GetCompletedBefore(ctx context.Context, before time.Time) ([]*Todo, error)
}
