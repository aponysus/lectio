CREATE TABLE IF NOT EXISTS language_notes (
    id TEXT PRIMARY KEY,
    engagement_id TEXT NOT NULL REFERENCES engagements(id),
    term TEXT NULL,
    language TEXT NULL,
    note_type TEXT NULL,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at DATETIME NULL
);

CREATE INDEX IF NOT EXISTS idx_language_notes_engagement_id ON language_notes(engagement_id);
CREATE INDEX IF NOT EXISTS idx_language_notes_language ON language_notes(language);
CREATE INDEX IF NOT EXISTS idx_language_notes_note_type ON language_notes(note_type);
CREATE INDEX IF NOT EXISTS idx_language_notes_archived_at ON language_notes(archived_at);
