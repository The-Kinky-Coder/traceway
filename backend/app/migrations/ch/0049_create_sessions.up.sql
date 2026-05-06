CREATE TABLE IF NOT EXISTS sessions
(
    `id` UUID,
    `project_id` UUID,
    `started_at` DateTime,
    `ended_at` Nullable(DateTime) DEFAULT NULL,
    `duration` Int64 DEFAULT 0,
    `client_ip` String DEFAULT '',
    `attributes` String DEFAULT '{}',
    `app_version` LowCardinality(String) DEFAULT '',
    `server_name` LowCardinality(String) DEFAULT '',
    `distributed_trace_id` Nullable(UUID) DEFAULT NULL,
    `version` DateTime64(3) DEFAULT now64(3),
    INDEX idx_session_id id TYPE bloom_filter(0.001) GRANULARITY 1
)
ENGINE = ReplacingMergeTree(version)
PARTITION BY toYYYYMMDD(started_at)
ORDER BY (project_id, id)
SETTINGS index_granularity = 8192
