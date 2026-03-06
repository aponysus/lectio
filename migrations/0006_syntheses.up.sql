CREATE TABLE IF NOT EXISTS syntheses (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    type TEXT NOT NULL,
    inquiry_id TEXT NULL REFERENCES inquiries(id),
    notes TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at DATETIME NULL
);

CREATE INDEX IF NOT EXISTS idx_syntheses_inquiry_id ON syntheses(inquiry_id);
CREATE INDEX IF NOT EXISTS idx_syntheses_type ON syntheses(type);
CREATE INDEX IF NOT EXISTS idx_syntheses_archived_at ON syntheses(archived_at);
