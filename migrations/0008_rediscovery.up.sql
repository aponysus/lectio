CREATE TABLE IF NOT EXISTS rediscovery_items (
    id TEXT PRIMARY KEY,
    kind TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    reason TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rediscovery_items_status ON rediscovery_items(status);
CREATE INDEX IF NOT EXISTS idx_rediscovery_items_kind ON rediscovery_items(kind);
CREATE INDEX IF NOT EXISTS idx_rediscovery_items_target ON rediscovery_items(target_type, target_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_rediscovery_items_active_dedupe
    ON rediscovery_items(kind, target_type, target_id)
    WHERE status IN ('NEW', 'SEEN', 'DISMISSED');

ALTER TABLE inquiries ADD COLUMN reactivated_at DATETIME NULL;

CREATE INDEX IF NOT EXISTS idx_inquiries_reactivated_at ON inquiries(reactivated_at);
