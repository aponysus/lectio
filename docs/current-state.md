# Current State

Last updated: 2026-03-06

Lectio is now at the lean MVP line on the Go + chi + SQLite + React/Vite stack.

## What Is Live

- session auth, CSRF protection, SQLite migrations, and a working local dev shell
- source CRUD with archive behavior
- engagement capture with transactional create for inline inquiry links, claims, and language notes
- inquiry CRUD plus a workspace-style inquiry detail view
- claim capture, editing, archive behavior, and inquiry linkage
- synthesis creation, editing, archive behavior, and dashboard prompting
- language note capture and filtered engagement browse
- rediscovery v1 with dismiss and acted-on flows
- core search for sources, inquiries, engagements, and claims
- JSON export at `GET /api/export`
- narrow `import-v2` CLI support for legacy `sources` + `entries` JSON

## Supported Commands

- `make dev` starts the API and Vite dev server together
- `make test` runs the Go test suite with the pinned toolchain
- `make build-api` builds the Go server
- `cd web && npm run build` builds the frontend
- `go run ./cmd/lectio import-v2 /path/to/legacy.json` runs the v2 importer

## Verification Snapshot

The current closeout pass succeeded with:

- `make test`
- `cd web && npm run build`

The repo also has coverage for:

- archive regressions
- export payload generation
- transactional engagement capture
- golden-path API flow from source -> engagement -> synthesis

## Known Follow-Ups

- engagement edit is still a sequential client-orchestrated flow rather than a single backend transaction; create is hardened first because it is the MVP-critical path
- `import-v2` is intentionally narrow and does not attempt to recover tags, mood, energy, or higher-order v2 structures
- restore/unarchive flows are not part of the MVP; archived rows remain available through export
- the next product step should be real usage and friction capture, not more scope expansion

## Working Rule

Use the app for actual reading, viewing, and synthesis work before reopening the model or broadening the ontology. Any post-MVP changes should come from observed friction in that workflow.
