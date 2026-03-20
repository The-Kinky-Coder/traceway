CREATE TABLE IF NOT EXISTS notification_channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    channel_type TEXT NOT NULL DEFAULT '',
    config TEXT NOT NULL DEFAULT '{}',
    enabled INTEGER NOT NULL DEFAULT 1,
    created_by INTEGER,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS notification_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    channel_id INTEGER NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    rule_type TEXT NOT NULL DEFAULT '',
    config TEXT NOT NULL DEFAULT '{}',
    enabled INTEGER NOT NULL DEFAULT 1,
    cooldown_minutes INTEGER NOT NULL DEFAULT 15,
    severity TEXT NOT NULL DEFAULT '',
    snoozed_until DATETIME,
    created_by INTEGER,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS notification_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id TEXT NOT NULL,
    rule_id INTEGER REFERENCES notification_rules(id) ON DELETE SET NULL,
    channel_id INTEGER REFERENCES notification_channels(id) ON DELETE SET NULL,
    rule_type TEXT NOT NULL DEFAULT '',
    rule_name TEXT NOT NULL DEFAULT '',
    channel_name TEXT NOT NULL DEFAULT '',
    severity TEXT NOT NULL DEFAULT '',
    subject TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'sent',
    error_message TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
