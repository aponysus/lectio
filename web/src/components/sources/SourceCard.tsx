import { Link } from 'react-router-dom'
import type { Source } from '../../api/client'

export function SourceCard({ source }: { source: Source }) {
  return (
    <article className="rounded-[1.5rem] border border-black/5 bg-white/75 p-4 shadow-card backdrop-blur">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-accent/80">{formatMedium(source.medium)}</p>
          <h3 className="mt-2 font-display text-[1.65rem] leading-tight text-ink">{source.title}</h3>
          <p className="mt-2 text-sm leading-6 text-ink/70">
            {source.creator ?? 'Unknown creator'}
            {source.year ? ` • ${source.year}` : ''}
          </p>
        </div>
        <Link
          to={`/sources/${source.id}`}
          className="rounded-xl bg-pine px-3 py-2 text-sm text-white transition hover:bg-pine/90"
        >
          Open
        </Link>
      </div>

      <dl className="mt-5 grid gap-4 sm:grid-cols-2">
        <div className="rounded-2xl bg-black/[0.03] px-4 py-3">
          <dt className="text-xs uppercase tracking-[0.2em] text-accent/70">Language</dt>
          <dd className="mt-2 text-sm text-ink/80">{source.original_language ?? 'Not set'}</dd>
        </div>
        <div className="rounded-2xl bg-black/[0.03] px-4 py-3">
          <dt className="text-xs uppercase tracking-[0.2em] text-accent/70">Updated</dt>
          <dd className="mt-2 text-sm text-ink/80">{formatDate(source.updated_at)}</dd>
        </div>
      </dl>

      {source.culture_or_context ? (
        <p className="mt-4 text-sm leading-6 text-ink/72">{source.culture_or_context}</p>
      ) : null}
    </article>
  )
}

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString()
}

function formatMedium(value: string) {
  return value.toLowerCase().replace(/_/g, ' ')
}
