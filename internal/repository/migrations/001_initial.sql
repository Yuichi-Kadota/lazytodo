-- lazytodo initial schema
-- Closure Table pattern for hierarchical data

PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

-- Workspaces
CREATE TABLE IF NOT EXISTS workspaces (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    is_expanded INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at TEXT  -- Soft delete (NULL = active)
);

-- Workspace hierarchy (Closure Table)
CREATE TABLE IF NOT EXISTS workspace_closure (
    ancestor_id TEXT NOT NULL,
    descendant_id TEXT NOT NULL,
    depth INTEGER NOT NULL DEFAULT 0,

    PRIMARY KEY (ancestor_id, descendant_id),
    FOREIGN KEY (ancestor_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    FOREIGN KEY (descendant_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    CHECK (depth >= 0)
);

CREATE INDEX IF NOT EXISTS idx_workspace_closure_ancestor ON workspace_closure(ancestor_id);
CREATE INDEX IF NOT EXISTS idx_workspace_closure_descendant ON workspace_closure(descendant_id);
CREATE INDEX IF NOT EXISTS idx_workspace_closure_depth ON workspace_closure(depth);

-- Todos
CREATE TABLE IF NOT EXISTS todos (
    id TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL,
    description TEXT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    urgency INTEGER NOT NULL DEFAULT 1,
    due_date TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    completed_at TEXT,
    deleted_at TEXT,  -- Soft delete (NULL = active)
    is_archived INTEGER NOT NULL DEFAULT 0,  -- For search optimization

    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    CHECK (status IN ('pending', 'completed')),
    CHECK (urgency BETWEEN 1 AND 4)
);

-- Todo hierarchy (Closure Table)
CREATE TABLE IF NOT EXISTS todo_closure (
    ancestor_id TEXT NOT NULL,
    descendant_id TEXT NOT NULL,
    depth INTEGER NOT NULL DEFAULT 0,

    PRIMARY KEY (ancestor_id, descendant_id),
    FOREIGN KEY (ancestor_id) REFERENCES todos(id) ON DELETE CASCADE,
    FOREIGN KEY (descendant_id) REFERENCES todos(id) ON DELETE CASCADE,
    CHECK (depth >= 0)
);

CREATE INDEX IF NOT EXISTS idx_todo_closure_ancestor ON todo_closure(ancestor_id);
CREATE INDEX IF NOT EXISTS idx_todo_closure_descendant ON todo_closure(descendant_id);
CREATE INDEX IF NOT EXISTS idx_todos_workspace ON todos(workspace_id);
CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status);
CREATE INDEX IF NOT EXISTS idx_todos_search ON todos(is_archived, deleted_at);

-- Write-Ahead Log + Undo history
CREATE TABLE IF NOT EXISTS operation_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    operation_type TEXT NOT NULL,  -- 'create', 'update', 'delete', 'move'
    entity_type TEXT NOT NULL,     -- 'workspace', 'todo'
    entity_id TEXT NOT NULL,
    payload TEXT NOT NULL,         -- JSON (includes previous state)
    applied INTEGER NOT NULL DEFAULT 0,
    is_undone INTEGER NOT NULL DEFAULT 0,
    undo_group_id TEXT,            -- Group related operations
    created_at TEXT NOT NULL DEFAULT (datetime('now')),

    CHECK (operation_type IN ('create', 'update', 'delete', 'move')),
    CHECK (entity_type IN ('workspace', 'todo'))
);

CREATE INDEX IF NOT EXISTS idx_operation_log_applied ON operation_log(applied);
CREATE INDEX IF NOT EXISTS idx_operation_log_undo ON operation_log(is_undone, created_at DESC);

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT OR IGNORE INTO schema_version (version) VALUES (1);
