CREATE TABLE IF NOT EXISTS sources (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    medium TEXT NOT NULL,
    creator TEXT NULL,
    year INTEGER NULL,
    original_language TEXT NULL,
    culture_or_context TEXT NULL,
    notes TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at DATETIME NULL
);

CREATE INDEX IF NOT EXISTS idx_sources_title ON sources(title);
CREATE INDEX IF NOT EXISTS idx_sources_medium ON sources(medium);
CREATE INDEX IF NOT EXISTS idx_sources_original_language ON sources(original_language);
CREATE INDEX IF NOT EXISTS idx_sources_archived_at ON sources(archived_at);
