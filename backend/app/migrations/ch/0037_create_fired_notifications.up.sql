CREATE TABLE IF NOT EXISTS fired_notifications (
    id UUID DEFAULT generateUUIDv4(),
    project_id UUID,
    rule_id Int32,
    rule_type String,
    rule_name String,
    channel_type String,
    channel_name String,
    severity String,
    subject String,
    body String,
    status String DEFAULT 'sent',
    error_message String DEFAULT '',
    endpoint String DEFAULT '',
    fired_at DateTime64(3) DEFAULT now64(3)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(fired_at)
ORDER BY (project_id, fired_at)
