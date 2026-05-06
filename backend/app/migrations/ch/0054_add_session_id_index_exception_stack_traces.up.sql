ALTER TABLE exception_stack_traces ADD INDEX idx_exceptions_session_id session_id TYPE bloom_filter(0.001) GRANULARITY 1
