package domain

import "time"

// Workspace represents a workspace entity
type Workspace struct {
	ID         string
	Name       string
	Position   int
	IsExpanded bool
	Depth      int // Calculated from closure table
	ParentID   string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

// IsDeleted returns true if the workspace is soft-deleted
func (w *Workspace) IsDeleted() bool {
	return w.DeletedAt != nil
}

// IsSystem returns true if this is a system workspace (e.g., _archive)
func (w *Workspace) IsSystem() bool {
	return len(w.Name) > 0 && w.Name[0] == '_'
}
