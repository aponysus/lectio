CREATE TABLE IF NOT EXISTS engagements (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL REFERENCES sources(id),
    engaged_at DATETIME NOT NULL,
    portion_label TEXT NULL,
    reflection TEXT NOT NULL,
    why_it_matters TEXT NULL,
    source_language TEXT NULL,
    reflection_language TEXT NULL,
    access_mode TEXT NULL,
    revisit_priority INTEGER NULL,
    is_reread_or_rewatch BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at DATETIME NULL
);

CREATE INDEX IF NOT EXISTS idx_engagements_source_id ON engagements(source_id);
CREATE INDEX IF NOT EXISTS idx_engagements_engaged_at ON engagements(engaged_at);
CREATE INDEX IF NOT EXISTS idx_engagements_access_mode ON engagements(access_mode);
CREATE INDEX IF NOT EXISTS idx_engagements_archived_at ON engagements(archived_at);
