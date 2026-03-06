# Import and Export

Lectio now has two MVP portability paths:

- `GET /api/export` for a full JSON snapshot of the current workspace
- `lectio import-v2 <path-to-legacy-json>` for a narrow migration path from the older v2 data shape

## Export

`GET /api/export` is an authenticated API route. It returns a JSON document with a `Content-Disposition` attachment header using this filename convention:

```text
lectio-export-YYYY-MM-DD.json
```

The payload currently includes:

- `sources`
- `engagements`
- `inquiries`
- `engagement_inquiries`
- `claims`
- `claim_inquiries`
- `language_notes`
- `syntheses`
- `rediscovery_items`

This export is table-oriented rather than presentation-oriented. It is meant for backup, inspection, and migration work. Archived rows are included, with `archived_at` populated when relevant.

Example:

```sh
curl -sS -b cookies.txt http://127.0.0.1:8080/api/export -o lectio-export.json
```

The frontend export page at `/settings/export` is just a thin wrapper around this same route.

## Import v2

The `import-v2` command is intentionally narrow. It exists to approximate the documented v2 dataset shape from `.plan/plan_v2.md`, not to ingest arbitrary legacy exports.

Run it against the same database the app uses:

```sh
go run ./cmd/lectio import-v2 /path/to/legacy.json
```

If you already have a compiled binary:

```sh
./lectio import-v2 /path/to/legacy.json
```

The importer expects top-level `sources` and `entries` arrays.

Supported source fields:

- `id`
- `title`
- `author`
- `language`
- `tradition`

Supported entry fields:

- `id`
- `source_id`
- `source`
- `reflection`
- `passage`
- `created_at`
- `updated_at`

Current mapping behavior:

- `source.author -> sources.creator`
- `source.language -> sources.original_language`
- `source.tradition -> sources.culture_or_context`
- `entry.created_at` or `entry.updated_at -> engagements.engaged_at`
- `entry.reflection -> engagements.reflection`
- `entry.passage -> appended to engagements.reflection`

Deliberately ignored during import:

- tags
- mood
- energy
- legacy fields not recognized by the narrow importer
- richer v2 concepts that do not exist in the lean MVP schema

If a legacy entry points at an unknown source, the importer falls back to creating a placeholder source so the engagement is still preserved.

## Operational Notes

- `lectio import-v2` runs migrations first, so it can target a fresh SQLite database.
- The importer only creates `sources` and `engagements`. It does not attempt to infer inquiries, claims, syntheses, or rediscovery state.
- The export format is a backup and migration format, not a stable public API contract yet. Treat field additions as possible while the MVP is still moving.
