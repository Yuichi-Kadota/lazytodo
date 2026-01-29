package wal

import (
	"encoding/json"
	"time"
)

// OperationType represents the type of operation
type OperationType string

const (
	OpCreate OperationType = "create"
	OpUpdate OperationType = "update"
	OpDelete OperationType = "delete"
	OpMove   OperationType = "move"
)

// EntityType represents the type of entity
type EntityType string

const (
	EntityWorkspace EntityType = "workspace"
	EntityTodo      EntityType = "todo"
)

// Operation represents a single operation in the WAL
type Operation struct {
	ID            int64         `json:"id,omitempty"`
	OperationType OperationType `json:"operation_type"`
	EntityType    EntityType    `json:"entity_type"`
	EntityID      string        `json:"entity_id"`
	Payload       Payload       `json:"payload"`
	Applied       bool          `json:"applied"`
	IsUndone      bool          `json:"is_undone"`
	UndoGroupID   string        `json:"undo_group_id,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

// Payload contains the operation data
type Payload struct {
	Before json.RawMessage `json:"before,omitempty"` // State before operation (for undo)
	After  json.RawMessage `json:"after,omitempty"`  // State after operation
	Extra  json.RawMessage `json:"extra,omitempty"`  // Additional data (e.g., position changes)
}

// MarshalPayload converts the payload to JSON
func (p Payload) MarshalJSON() ([]byte, error) {
	type payloadAlias Payload
	return json.Marshal(payloadAlias(p))
}

// UnmarshalPayload parses JSON into payload
func (p *Payload) UnmarshalJSON(data []byte) error {
	type payloadAlias Payload
	return json.Unmarshal(data, (*payloadAlias)(p))
}
