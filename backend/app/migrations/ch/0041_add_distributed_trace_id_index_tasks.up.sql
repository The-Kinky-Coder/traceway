ALTER TABLE tasks ADD INDEX idx_distributed_trace_id_tasks distributed_trace_id TYPE bloom_filter(0.001) GRANULARITY 1
