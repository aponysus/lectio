import { Link } from 'react-router-dom'
import type { Claim } from '../../api/client'

export function ClaimListRow({ claim }: { claim: Claim }) {
  return (
    <article className="rounded-[1.25rem] border border-black/5 bg-white/78 px-4 py-4 shadow-card backdrop-blur">
      <div className="grid gap-4 lg:grid-cols-[minmax(0,2.2fr)_1fr_1fr_1fr_auto] lg:items-center">
        <div className="min-w-0">
          <p className="text-xs uppercase tracking-[0.2em] text-accent/78">
            {formatLabel(claim.claim_type)} • {formatLabel(claim.status)}
          </p>
          <h3 className="mt-2 line-clamp-2 font-display text-[1.35rem] leading-tight text-ink">{claim.text}</h3>
          <p className="mt-2 truncate text-sm text-ink/72">
            {claim.origin?.source_title ?? 'Unlinked claim'}
            {claim.origin?.portion_label ? ` • ${claim.origin.portion_label}` : ''}
          </p>
        </div>

        <MetaColumn label="Confidence" value={claim.confidence ? String(claim.confidence) : 'Not set'} />
        <MetaColumn label="Updated" value={formatDateTime(claim.updated_at)} />
        <MetaColumn label="Origin" value={claim.origin ? formatLabel(claim.origin.source_medium) : 'No direct link'} />

        <div className="flex justify-start lg:justify-end">
          {claim.origin ? (
            <Link
              to={`/engagements/${claim.origin.engagement_id}`}
              className="rounded-xl bg-pine px-3 py-2 text-sm text-white transition hover:bg-pine/90"
            >
              Open
            </Link>
          ) : (
            <span className="rounded-xl bg-black/[0.03] px-3 py-2 text-sm text-ink/58">No target</span>
          )}
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

function formatDateTime(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

function formatLabel(value: string) {
  return value.toLowerCase().replace(/_/g, ' ')
}
