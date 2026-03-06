import { Link } from 'react-router-dom'
import type { Source } from '../../api/client'

export function SourceListRow({ source }: { source: Source }) {
  return (
    <article className="rounded-[1.25rem] border border-black/5 bg-white/78 px-4 py-4 shadow-card backdrop-blur">
      <div className="grid gap-4 lg:grid-cols-[minmax(0,2.4fr)_1fr_1.1fr_1fr_auto] lg:items-center">
        <div className="min-w-0">
          <p className="text-xs uppercase tracking-[0.2em] text-accent/78">{formatMedium(source.medium)}</p>
          <h3 className="mt-2 truncate font-display text-[1.35rem] leading-tight text-ink">{source.title}</h3>
          <p className="mt-2 truncate text-sm text-ink/72">
            {source.creator ?? 'Unknown creator'}
            {source.year ? ` • ${source.year}` : ''}
          </p>
          {source.culture_or_context ? (
            <p className="mt-2 line-clamp-1 text-sm leading-6 text-ink/68">{source.culture_or_context}</p>
          ) : null}
        </div>

        <MetaColumn label="Language" value={source.original_language ?? 'Not set'} />
        <MetaColumn label="Creator" value={source.creator ?? 'Unknown'} />
        <MetaColumn label="Updated" value={formatDate(source.updated_at)} />

        <div className="flex justify-start lg:justify-end">
          <Link
            to={`/sources/${source.id}`}
            className="rounded-xl bg-pine px-3 py-2 text-sm text-white transition hover:bg-pine/90"
          >
            Open
          </Link>
        </div>
      </div>
    </article>
  )
}

function MetaColumn({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0">
      <p className="text-xs uppercase tracking-[0.18em] text-accent/72">{label}</p>
      <p className="mt-2 truncate text-sm text-ink/78">{value}</p>
    </div>
  )
}

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString()
}

function formatMedium(value: string) {
  return value.toLowerCase().replace(/_/g, ' ')
}
