PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    display_name  TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'employee',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workflow_templates (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    business_type TEXT NOT NULL,
    is_active     INTEGER NOT NULL DEFAULT 1,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workflow_nodes (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id   INTEGER NOT NULL REFERENCES workflow_templates(id) ON DELETE CASCADE,
    code          TEXT NOT NULL,
    name          TEXT NOT NULL,
    node_type     TEXT NOT NULL,             -- START / APPROVAL / END
    approve_type  TEXT NOT NULL DEFAULT 'ANY',
    join_type     TEXT NOT NULL DEFAULT 'ANY',
    UNIQUE(template_id, code)
);

CREATE TABLE IF NOT EXISTS workflow_node_approvers (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id       INTEGER NOT NULL REFERENCES workflow_nodes(id) ON DELETE CASCADE,
    approver_type TEXT NOT NULL,             -- USER / ROLE
    approver_ref  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS workflow_edges (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id   INTEGER NOT NULL REFERENCES workflow_templates(id) ON DELETE CASCADE,
    from_node_id  INTEGER NOT NULL REFERENCES workflow_nodes(id) ON DELETE CASCADE,
    to_node_id    INTEGER NOT NULL REFERENCES workflow_nodes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS workflow_instances (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id   INTEGER NOT NULL REFERENCES workflow_templates(id),
    business_type TEXT NOT NULL,
    business_id   TEXT NOT NULL,
    title         TEXT NOT NULL,
    form_data     TEXT NOT NULL DEFAULT '{}',
    initiator_id  INTEGER NOT NULL REFERENCES users(id),
    status        TEXT NOT NULL DEFAULT 'RUNNING',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workflow_instance_nodes (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    instance_id   INTEGER NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    node_id       INTEGER NOT NULL REFERENCES workflow_nodes(id),
    status        TEXT NOT NULL DEFAULT 'PENDING',
    activated_at  DATETIME,
    completed_at  DATETIME,
    UNIQUE(instance_id, node_id)
);

CREATE TABLE IF NOT EXISTS workflow_tasks (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    instance_id       INTEGER NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    instance_node_id  INTEGER NOT NULL REFERENCES workflow_instance_nodes(id) ON DELETE CASCADE,
    approver_id       INTEGER NOT NULL REFERENCES users(id),
    status            TEXT NOT NULL DEFAULT 'PENDING',
    comment           TEXT NOT NULL DEFAULT '',
    acted_at          DATETIME,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workflow_logs (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    instance_id   INTEGER NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    actor_id      INTEGER,
    action        TEXT NOT NULL,
    detail        TEXT NOT NULL DEFAULT '',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tasks_approver_status ON workflow_tasks(approver_id, status);
CREATE INDEX IF NOT EXISTS idx_instance_nodes_instance ON workflow_instance_nodes(instance_id);
CREATE INDEX IF NOT EXISTS idx_edges_template_from ON workflow_edges(template_id, from_node_id);
CREATE INDEX IF NOT EXISTS idx_edges_template_to ON workflow_edges(template_id, to_node_id);
