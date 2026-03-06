# Lectio User Guide

Last updated: 2026-03-06

This guide is for using Lectio as it exists today: a private single-user workspace for serious source work. The MVP is built around one loop:

1. record a source
2. log an engagement with that source
3. connect the engagement to one or more inquiries
4. extract a small number of claims
5. revisit what matters
6. compress the work into synthesis

If you are looking for import/export specifics or a short operator note, see:

- [Import and Export](./import-export.md)
- [Current State](./current-state.md)

## 1. What Lectio Is For

Lectio is not a general-purpose notes app and it is not trying to be a second brain with an open-ended ontology. It is for deliberate intellectual work around meaningful inputs such as:

- books
- essays and papers
- lectures and podcasts
- films and television
- conversations
- scripture and other primary texts

The app is organized around actual engagements with those sources, not around passive collection.

## 2. First Run

### Local development setup

If you are running Lectio locally, the simplest path is:

```sh
make dev
```

That starts:

- the Go API on `http://localhost:8080`
- the Vite frontend on `http://localhost:5173`

The frontend proxies `/api` requests to the local API during development.

### Bootstrap login

Lectio currently uses a single bootstrap password rather than multi-user accounts.

The default local credentials are defined in `.env.example`:

```text
LECTIO_BOOTSTRAP_PASSWORD=changeme
```

Sign in with that password unless you have changed it in your environment.

Important:

- change the bootstrap password and secrets before deploying anywhere real
- there is no user registration flow in the MVP
- the app is intentionally private and single-user

## 3. Workspace Layout

After login, the app is organized around six main areas:

- `Dashboard`: recent work, active inquiries, synthesis prompts, rediscovery
- `Search`: cross-section text search
- `Sources`: stable records for books, lectures, films, and other inputs
- `Engagements`: captured reading, viewing, or listening sessions
- `Inquiries`: the live questions that organize your work
- `Syntheses`: written compression of an inquiry’s current state

The left sidebar also provides quick actions:

- log engagement
- create inquiry
- create source
- export data
- sign out

## 4. Core Concepts

### Source

A source is the thing you engaged with: a book, article, film, lecture, podcast, conversation, and so on.

Supported source media:

- `BOOK`
- `ESSAY`
- `ARTICLE`
- `PAPER`
- `SCRIPTURE`
- `LECTURE`
- `PODCAST`
- `FILM`
- `TV`
- `CONVERSATION`
- `OTHER`

Recommended source fields:

- `title`
- `medium`
- `creator`
- `year`
- `original language`
- `culture / context`
- `notes`

Keep source records lean. The goal is stable identification, not cataloging for its own sake.

### Engagement

An engagement is a concrete encounter with a source on a specific date. This is the core record in Lectio.

Core engagement fields:

- `source`
- `engaged at`
- `reflection`

Optional engagement fields:

- `portion label`
- `why it matters`
- `source language`
- `reflection language`
- `access mode`
- `revisit priority`
- `reread / rewatch`

Supported access modes:

- `ORIGINAL`
- `TRANSLATION`
- `BILINGUAL`
- `SUBTITLED`
- `LOOKUP_HEAVY`
- `OTHER`

### Inquiry

An inquiry is a live question or line of thought that organizes multiple engagements.

Inquiry fields:

- `title`
- `question`
- `status`
- `why it matters`
- `current view`
- `open tensions`

Supported inquiry statuses:

- `ACTIVE`
- `DORMANT`
- `SYNTHESIZED`
- `ABANDONED`

### Claim

A claim is a sharpened statement, interpretation, question, or hypothesis extracted from an engagement.

Supported claim types:

- `OBSERVATION`
- `INTERPRETATION`
- `PERSONAL_VIEW`
- `QUESTION`
- `HYPOTHESIS`

Supported claim statuses:

- `ACTIVE`
- `TENTATIVE`
- `REVISED`
- `ABANDONED`

### Language Note

A language note captures wording, translation, register, idiom, or cultural nuance that mattered during an engagement.

Supported language note types:

- `TRANSLATION`
- `REGISTER`
- `IDIOM`
- `COLLOCATION`
- `CULTURAL_NUANCE`
- `OTHER`

### Synthesis

A synthesis is a written attempt to compress the current state of an inquiry.

Supported synthesis types:

- `CHECKPOINT`
- `COMPARISON`
- `POSITION`

Each synthesis belongs to a single inquiry in the MVP.

### Rediscovery Item

A rediscovery item is an automatically surfaced prompt telling you that something older may need another pass.

Rediscovery is intentionally small and rule-based in the MVP.

## 5. Recommended Workflow

If you are not sure where to begin, use this sequence:

1. Create a source.
2. Log an engagement the next time you read, watch, listen, or discuss with intent.
3. Link the engagement to an existing inquiry, or create an inquiry inline.
4. Extract one to three claims if the reflection is ready to sharpen.
5. Add language notes only if wording or translation genuinely mattered.
6. Revisit the inquiry workspace after several engagements.
7. Write a synthesis when the inquiry has enough density.

This order matches the app’s design. Sources support engagements. Engagements feed inquiries. Inquiries accumulate claims and eventually produce synthesis.

## 6. Using the Dashboard

The dashboard is the default landing page after sign-in.

It shows:

- quick actions for new source, new inquiry, and new engagement
- counts for active inquiries, sources, synthesis-ready inquiries, and rediscovery items
- recent engagements
- active inquiries
- synthesis prompts
- rediscovery prompts

Use the dashboard for daily orientation, not for deep editing. The typical moves from here are:

- resume a recent engagement or inquiry
- log a new engagement
- open a synthesis prompt
- act on or dismiss a rediscovery item

## 7. Working with Sources

### Create a source

Go to `Sources` and choose `New source`.

You can also create a source from the dashboard or sidebar.

Use sources for inputs you expect to revisit or connect to future work. If something is only incidental and will never matter again, you probably do not need a source record for it.

### Browse and filter sources

The source list supports:

- search by title or creator
- filter by medium
- filter by original language
- sort by recently updated or title
- list view or card view

### Source detail page

A source detail page shows:

- source metadata
- source notes
- all visible engagements linked to that source

From the source page, you can:

- log an engagement
- edit the source
- archive the source

If a source has no engagements yet, the detail page shows a zero-state with a direct link into the engagement flow.

## 8. Logging an Engagement

### Why this matters

Engagement logging is the center of Lectio. This is where the app stops being a library and becomes a thinking tool.

### Start points

You can start a new engagement from:

- the dashboard
- the sidebar quick action
- the source detail page
- the inquiry detail page

If you start from a source or inquiry, Lectio preselects that context in the form.

### Form structure

The engagement form is intentionally sectioned:

- core capture
- inquiry links
- claims
- advanced metadata

The shortest useful path is:

1. choose source
2. set date
3. write reflection
4. save

Then add inquiry links, claims, and language notes when they add real value.

### Inquiry links during capture

You can:

- link the engagement to existing inquiries
- create a new inquiry inline while logging the engagement

On create, the engagement and its inline inquiry data are saved transactionally. That means the main engagement record and the nested inquiry, claim, and language-note records are created together rather than as a fragile chain of separate requests.

### Claim capture during engagement logging

The MVP keeps claim extraction inside the engagement flow. You can add up to three claims during capture.

Each claim can include:

- text
- claim type
- confidence
- claim status
- notes

Use this section when the reflection contains something worth sharpening into a more explicit statement or question.

### Language notes during engagement logging

Use language notes only when wording genuinely affected the encounter. They are optional.

Good reasons to add one:

- a translation choice changed the force of a passage
- a phrase carried a register or idiomatic meaning you want to preserve
- a cultural nuance matters to later interpretation

## 9. Browsing Engagements

The engagement list supports:

- text search on reflection content
- filter by access mode
- filter to engagements with language notes only
- list view or card view

Use this page when you want to scan the body of your captured encounters rather than enter through a source or inquiry.

### Engagement detail page

An engagement detail page shows:

- the full reflection
- why it matters
- source and metadata
- linked inquiries
- extracted claims
- language notes

From here, you can:

- open the source
- edit the engagement
- archive the engagement

## 10. Working with Inquiries

### Create an inquiry

Use `Inquiries -> New inquiry` when you already know the question you want to pursue.

Use inline inquiry creation during engagement logging when the question emerges from the engagement itself.

### Browse and filter inquiries

The inquiry list supports:

- search by title or question
- filter by status
- list view or card view

### Inquiry workspace

The inquiry detail page is the app’s main thinking workspace.

It is organized into three functional areas:

- `Workspace compass`: the question, why it matters, current view, open tensions
- `Evidence feed`: linked engagements and claim revisions in one chronological stream
- `Synthesis rail`: syntheses, readiness, and next moves

Use inquiry pages to:

- understand what the inquiry is actually about
- see the evidence attached to it
- revise claims in context
- decide whether the inquiry is ready for synthesis

### Inquiry status guidance

Use statuses roughly this way:

- `ACTIVE`: currently being worked
- `DORMANT`: still meaningful, but not under active pressure
- `SYNTHESIZED`: has been compressed into a current position
- `ABANDONED`: intentionally dropped

These statuses affect filtering and rediscovery behavior, so it is worth keeping them current.

## 11. Claims

Claims are visible in two main places:

- the engagement they came from
- the inquiry they are linked to

In the MVP, claim capture is tied closely to engagement logging and inquiry work. Claims are not the primary top-level object; they serve the inquiry loop.

Use claims for statements that need to survive beyond one reflection paragraph, especially when you expect to revisit or revise them later.

Good claim examples:

- an interpretation worth testing
- a hypothesis you are not ready to assert strongly
- a question that should stay visible across future engagements

Weak claim examples:

- vague summaries that belong in the reflection itself
- raw quotes without interpretation
- duplicate fragments that do not change later work

## 12. Syntheses

### What synthesis is for

Synthesis is the payoff layer. It is where multiple engagements and claims get compressed into a clearer current position.

### When an inquiry is probably ready

In the current MVP, an inquiry is treated as synthesis-ready when it has either:

- at least 3 engagements, or
- at least 2 claims

Those thresholds drive dashboard prompting and inquiry-page readiness cues.

### Creating a synthesis

You can begin a synthesis from:

- the dashboard synthesis prompt
- the inquiry workspace
- the syntheses section after choosing an inquiry

Each synthesis includes:

- title
- body
- type
- inquiry
- optional notes

The form preloads inquiry context so the writing starts from the existing inquiry material rather than from scratch.

### Browsing syntheses

The syntheses page supports:

- list view
- card view
- browsing recent syntheses across all inquiries

There is no separate synthesis search page in the MVP. Use inquiry pages or the syntheses list when you want to browse them.

## 13. Language Notes

Language notes are attached to engagements, not stored independently.

Use them when:

- you read in translation
- you compare original and translated wording
- you notice idiom, register, or collocation that matters
- cultural context changes interpretation

You can browse engagements with language notes by enabling the `With language notes only` filter on the engagement list.

## 14. Rediscovery

Rediscovery is the lightweight resurfacing layer on the dashboard.

The MVP currently surfaces exactly four prompt types:

1. tentative claims older than 14 days
2. an old engagement linked to an active inquiry, older than 30 days
3. an inquiry with at least 4 engagements and no synthesis
4. an inquiry reactivated from dormant to active in the last 7 days

For each rediscovery item, you can:

- follow the prompt into the relevant record
- mark it as acted on
- dismiss it

Dismissal removes it from the active feed. If the underlying condition disappears naturally, the app also resolves the prompt automatically.

## 15. Search

The `Search` page searches across:

- sources
- inquiries
- engagements
- claims

Search works well with:

- source titles
- creator names
- question fragments
- memorable reflection phrases
- claim wording

The page shows up to six results per section. If you want to go deeper, use `Open full list` from the search results to jump into the corresponding browse page with the query applied.

## 16. Export and Import

### Export

The MVP includes a full JSON export surface at `Settings -> Export`.

Export is intended for:

- backup
- inspection
- migration work

The export includes the core tables and join tables. Archived rows are included too, with `archived_at` populated when relevant.

### Import v2

There is also a narrow CLI importer for older v2 JSON dumps:

```sh
go run ./cmd/lectio import-v2 /path/to/legacy.json
```

This importer is intentionally limited. It mainly maps older `sources` and `entries` into the current MVP’s `sources` and `engagements`.

For full details, see [Import and Export](./import-export.md).

## 17. Archive Behavior

Lectio archives records rather than deleting them through the UI.

What archive means in practice:

- archived rows disappear from normal browse lists
- archived rows stop participating in most active workflows
- archived rows still remain in the database
- archived rows are still included in full export

Important archive notes:

- archiving a source does not delete its engagements
- archiving an inquiry does not delete linked engagements, claims, or syntheses
- archiving an engagement removes it from active lists and related prompts
- there is currently no restore or unarchive flow in the MVP

Archive deliberately. It is a visibility control, not a reversible trash can in the current UI.

## 18. Suggested Working Habits

The app works best if you keep the records small and honest.

Recommended habits:

- create sources only for inputs you expect to matter again
- log engagements soon after the encounter
- write reflections in your own words before extracting claims
- keep claims few and distinct
- use inquiry pages as the center of active work
- write syntheses as checkpoints, not final monuments
- act on rediscovery items quickly so the feed stays meaningful

## 19. Current MVP Limits

The MVP is strong enough for real use, but it still has boundaries.

Current limits:

- single-user bootstrap-password auth
- no restore/unarchive flow
- no rich import pipeline beyond the narrow v2 helper
- no separate top-level claim workspace beyond inquiry and engagement contexts
- no ontology layer for concepts, projects, or arbitrary graph relationships

These are deliberate constraints, not omissions by accident.

## 20. Troubleshooting

### I cannot sign in

Check the bootstrap password in your environment. For local development, the default in `.env.example` is `changeme`.

### I created a source but do not see it anymore

It may have been archived. Archived rows do not appear in normal list views and there is no restore UI yet.

### I expected a synthesis prompt but do not see one

The inquiry usually needs at least:

- 3 engagements, or
- 2 claims

The dashboard also only shows a small number of prompts at once.

### I expected a rediscovery prompt but do not see one

Rediscovery is rule-based and threshold-driven. Make sure the underlying condition is actually true and that the item was not already dismissed or marked acted on.

### I want to recover everything outside the UI

Use the export surface. The full JSON export is the current backup path.

## 21. Short Version

If you ignore everything else, use Lectio like this:

1. create the source
2. log the encounter
3. attach it to the real question
4. extract one to three claims
5. revisit the inquiry when it thickens
6. write synthesis when the inquiry is ready

That is the intended path through the product.
