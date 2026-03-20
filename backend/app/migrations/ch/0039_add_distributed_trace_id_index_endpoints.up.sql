ALTER TABLE endpoints ADD INDEX idx_distributed_trace_id distributed_trace_id TYPE bloom_filter(0.001) GRANULARITY 1
