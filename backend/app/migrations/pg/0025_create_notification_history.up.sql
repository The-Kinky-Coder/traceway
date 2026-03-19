CREATE TABLE IF NOT EXISTS notification_history (
    id SERIAL PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id),
    rule_id INTEGER REFERENCES notification_rules(id) ON DELETE SET NULL,
    channel_id INTEGER REFERENCES notification_channels(id) ON DELETE SET NULL,
    rule_type VARCHAR(50) NOT NULL,
    rule_name VARCHAR(200) NOT NULL,
    channel_name VARCHAR(200) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'sent',
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
