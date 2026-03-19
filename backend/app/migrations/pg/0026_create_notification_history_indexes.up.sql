CREATE INDEX idx_notification_history_project_created ON notification_history(project_id, created_at DESC);
