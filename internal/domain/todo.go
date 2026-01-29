package domain

import "time"

// Status represents the todo status
type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
)

// Urgency levels
const (
	UrgencyLow      = 1
	UrgencyMedium   = 2
	UrgencyHigh     = 3
	UrgencyCritical = 4
)

// Todo represents a todo item
type Todo struct {
	ID          string
	WorkspaceID string
	Description string
	Position    int
	Status      Status
	Urgency     int
	DueDate     *time.Time
	Depth       int // Calculated from closure table
	ParentID    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
	DeletedAt   *time.Time
	IsArchived  bool
}

// IsDeleted returns true if the todo is soft-deleted
func (t *Todo) IsDeleted() bool {
	return t.DeletedAt != nil
}

// IsCompleted returns true if the todo is completed
func (t *Todo) IsCompleted() bool {
	return t.Status == StatusCompleted
}

// IsPending returns true if the todo is pending
func (t *Todo) IsPending() bool {
	return t.Status == StatusPending
}

// IsOverdue returns true if the todo is overdue
func (t *Todo) IsOverdue() bool {
	if t.DueDate == nil || t.IsCompleted() {
		return false
	}
	return time.Now().After(*t.DueDate)
}

// IsDueToday returns true if the todo is due today
func (t *Todo) IsDueToday() bool {
	if t.DueDate == nil {
		return false
	}
	now := time.Now()
	return t.DueDate.Year() == now.Year() &&
		t.DueDate.Month() == now.Month() &&
		t.DueDate.Day() == now.Day()
}
