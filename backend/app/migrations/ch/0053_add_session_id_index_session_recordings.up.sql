ALTER TABLE session_recordings ADD INDEX idx_recordings_session_id session_id TYPE bloom_filter(0.001) GRANULARITY 1
