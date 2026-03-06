import { Link } from 'react-router-dom'
import { PageHeader } from '../../components/shared/PageHeader'

const includedTables = [
  'sources',
  'engagements',
  'inquiries',
  'engagement_inquiries',
  'claims',
  'claim_inquiries',
  'language_notes',
  'syntheses',
  'rediscovery_items',
]

export function ExportPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Settings"
        title="Export a JSON backup"
        description="Download a full JSON snapshot of the core Lectio tables, including join records and archived material, so the corpus is portable."
        actions={
          <a
            href="/api/export"
            className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
          >
            Download export
          </a>
        }
      />

      <section className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <article className="rounded-[1.5rem] border border-black/5 bg-white/72 p-5 shadow-card backdrop-blur">
          <p className="text-xs uppercase tracking-[0.22em] text-accent/80">What you get</p>
          <h3 className="mt-2 font-display text-[1.8rem] leading-tight text-ink">A table-oriented backup</h3>
          <p className="mt-3 text-sm leading-6 text-ink/74">
            The download is structured for durability rather than presentation. It includes the core records and join
            tables the MVP depends on, with timestamps preserved.
          </p>

          <div className="mt-5 grid gap-3 sm:grid-cols-2">
            {includedTables.map((tableName) => (
              <div key={tableName} className="rounded-[1.1rem] bg-black/[0.03] px-4 py-3">
                <p className="font-mono text-sm text-ink">{tableName}</p>
              </div>
            ))}
          </div>
        </article>

        <article className="rounded-[1.5rem] border border-black/5 bg-white/72 p-5 shadow-card backdrop-blur">
          <p className="text-xs uppercase tracking-[0.22em] text-accent/80">Usage</p>
          <h3 className="mt-2 font-display text-[1.8rem] leading-tight text-ink">Keep a portable snapshot</h3>
          <div className="mt-4 space-y-3 text-sm leading-6 text-ink/76">
            <p>The recommended filename is <code className="rounded bg-black/[0.04] px-1.5 py-0.5 text-[0.92em]">lectio-export-YYYY-MM-DD.json</code> and the API uses that convention automatically.</p>
            <p>Use this export for backups, migration work, or sanity checks before larger schema changes.</p>
            <p>Older v2 material can now be approximated with the <code className="rounded bg-black/[0.04] px-1.5 py-0.5 text-[0.92em]">lectio import-v2 &lt;file&gt;</code> CLI helper.</p>
          </div>

          <div className="mt-6 flex flex-wrap gap-3">
            <a
              href="/api/export"
              className="rounded-2xl bg-accent px-4 py-3 text-sm font-medium text-white transition hover:bg-accent/90"
            >
              Download now
            </a>
            <Link
              to="/search"
              className="rounded-2xl border border-black/10 bg-white/90 px-4 py-3 text-sm text-ink transition hover:bg-white"
            >
              Back to search
            </Link>
          </div>
        </article>
      </section>
    </div>
  )
}
