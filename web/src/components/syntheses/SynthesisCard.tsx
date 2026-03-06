import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import type { Synthesis } from '../../api/client'

type SynthesisCardProps = {
  synthesis: Synthesis
  actions?: ReactNode
}

export function SynthesisCard({ synthesis, actions }: SynthesisCardProps) {
  return (
    <article className="rounded-[1.5rem] border border-black/5 bg-white/75 p-4 shadow-card backdrop-blur">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-accent/80">
            {synthesis.type.toLowerCase().replace(/_/g, ' ')}
          </p>
          <h3 className="mt-2 font-display text-[1.65rem] leading-tight text-ink">{synthesis.title}</h3>
          {synthesis.inquiry ? (
            <p className="mt-2 text-sm leading-6 text-ink/72">
              {synthesis.inquiry.title} • {synthesis.inquiry.status.toLowerCase().replace(/_/g, ' ')}
            </p>
          ) : null}
        </div>
        {actions ? (
          <div className="flex flex-wrap gap-2">{actions}</div>
        ) : (
          <Link
            to={`/syntheses/${synthesis.id}`}
            className="rounded-xl bg-pine px-3 py-2 text-sm text-white transition hover:bg-pine/90"
          >
            Open
          </Link>
        )}
      </div>

      <p className="mt-4 line-clamp-5 whitespace-pre-wrap text-sm leading-6 text-ink/80">{synthesis.body}</p>

      <dl className="mt-5 grid gap-4 sm:grid-cols-2">
        <MetaItem label="Created" value={formatDateTime(synthesis.created_at)} />
        <MetaItem label="Updated" value={formatDateTime(synthesis.updated_at)} />
      </dl>

      {synthesis.notes ? (
        <div className="mt-4 rounded-2xl bg-black/[0.03] px-4 py-4">
          <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Notes</p>
          <p className="mt-3 whitespace-pre-wrap text-sm leading-6 text-ink/80">{synthesis.notes}</p>
        </div>
      ) : null}
    </article>
  )
}

function MetaItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl bg-black/[0.03] px-4 py-3">
      <dt className="text-xs uppercase tracking-[0.2em] text-accent/75">{label}</dt>
      <dd className="mt-2 text-sm text-ink/80">{value}</dd>
    </div>
  )
}

function formatDateTime(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}
