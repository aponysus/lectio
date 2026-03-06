CREATE TABLE IF NOT EXISTS app_meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO app_meta (key, value)
VALUES ('bootstrapped_at', CURRENT_TIMESTAMP)
ON CONFLICT(key) DO NOTHING;

INSERT INTO app_meta (key, value)
VALUES ('app_name', 'Lectio')
ON CONFLICT(key) DO NOTHING;
