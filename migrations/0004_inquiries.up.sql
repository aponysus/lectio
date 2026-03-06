CREATE TABLE IF NOT EXISTS inquiries (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    question TEXT NOT NULL,
    status TEXT NOT NULL,
    why_it_matters TEXT NULL,
    current_view TEXT NULL,
    open_tensions TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at DATETIME NULL
);

CREATE INDEX IF NOT EXISTS idx_inquiries_status ON inquiries(status);
CREATE INDEX IF NOT EXISTS idx_inquiries_archived_at ON inquiries(archived_at);
CREATE INDEX IF NOT EXISTS idx_inquiries_title ON inquiries(title);

CREATE TABLE IF NOT EXISTS engagement_inquiries (
    engagement_id TEXT NOT NULL REFERENCES engagements(id),
    inquiry_id TEXT NOT NULL REFERENCES inquiries(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (engagement_id, inquiry_id)
);

CREATE INDEX IF NOT EXISTS idx_engagement_inquiries_inquiry_id ON engagement_inquiries(inquiry_id);
