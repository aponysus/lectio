import { Link } from 'react-router-dom'
import type { Synthesis } from '../../api/client'

export function SynthesisListRow({ synthesis }: { synthesis: Synthesis }) {
  return (
    <article className="rounded-[1.25rem] border border-black/5 bg-white/78 px-4 py-4 shadow-card backdrop-blur">
      <div className="grid gap-4 lg:grid-cols-[minmax(0,2.2fr)_1fr_1.2fr_1fr_auto] lg:items-center">
        <div className="min-w-0">
          <p className="text-xs uppercase tracking-[0.2em] text-accent/78">{formatType(synthesis.type)}</p>
          <h3 className="mt-2 truncate font-display text-[1.35rem] leading-tight text-ink">{synthesis.title}</h3>
          {synthesis.inquiry ? (
            <p className="mt-2 truncate text-sm text-ink/72">
              {synthesis.inquiry.title} • {formatType(synthesis.inquiry.status)}
            </p>
          ) : null}
          <p className="mt-2 line-clamp-1 whitespace-pre-wrap text-sm leading-6 text-ink/68">{synthesis.body}</p>
        </div>

        <MetaColumn label="Type" value={formatType(synthesis.type)} />
        <MetaColumn label="Linked inquiry" value={synthesis.inquiry ? synthesis.inquiry.title : 'Unavailable'} />
        <MetaColumn label="Updated" value={formatDateTime(synthesis.updated_at)} />

        <div className="flex justify-start lg:justify-end">
          <Link
            to={`/syntheses/${synthesis.id}`}
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

function formatType(value: string) {
  return value.toLowerCase().replace(/_/g, ' ')
}

function formatDateTime(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}
