CREATE TABLE IF NOT EXISTS claims (
    id TEXT PRIMARY KEY,
    text TEXT NOT NULL,
    claim_type TEXT NOT NULL,
    confidence INTEGER NULL,
    status TEXT NOT NULL,
    origin_engagement_id TEXT NULL REFERENCES engagements(id),
    notes TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at DATETIME NULL
);

CREATE INDEX IF NOT EXISTS idx_claims_origin_engagement_id ON claims(origin_engagement_id);
CREATE INDEX IF NOT EXISTS idx_claims_status ON claims(status);
CREATE INDEX IF NOT EXISTS idx_claims_claim_type ON claims(claim_type);
CREATE INDEX IF NOT EXISTS idx_claims_archived_at ON claims(archived_at);

CREATE TABLE IF NOT EXISTS claim_inquiries (
    claim_id TEXT NOT NULL REFERENCES claims(id),
    inquiry_id TEXT NOT NULL REFERENCES inquiries(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (claim_id, inquiry_id)
);

CREATE INDEX IF NOT EXISTS idx_claim_inquiries_inquiry_id ON claim_inquiries(inquiry_id);
