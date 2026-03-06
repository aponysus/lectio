import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import type { Claim } from '../../api/client'

type ClaimCardProps = {
  claim: Claim
  actions?: ReactNode
}

export function ClaimCard({ claim, actions }: ClaimCardProps) {
  return (
    <article className="rounded-[1.75rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-accent/80">
            {claim.claim_type.toLowerCase().replace(/_/g, ' ')} • {claim.status.toLowerCase().replace(/_/g, ' ')}
          </p>
          <h3 className="mt-2 font-display text-2xl text-ink">{claim.text}</h3>
        </div>
        {actions ? <div className="flex flex-wrap gap-2">{actions}</div> : null}
      </div>

      <dl className="mt-5 grid gap-4 sm:grid-cols-2">
        <MetaItem label="Confidence" value={claim.confidence ? String(claim.confidence) : 'Not set'} />
        <MetaItem label="Updated" value={formatDateTime(claim.updated_at)} />
      </dl>

      {claim.notes ? (
        <div className="mt-4 rounded-2xl bg-black/[0.03] px-4 py-4">
          <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Notes</p>
          <p className="mt-3 whitespace-pre-wrap text-sm leading-6 text-ink/80">{claim.notes}</p>
        </div>
      ) : null}

      {claim.origin ? (
        <div className="mt-4 rounded-2xl bg-black/[0.03] px-4 py-4">
          <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Origin engagement</p>
          <div className="mt-3 flex items-center justify-between gap-4">
            <div className="min-w-0">
              <p className="text-sm font-medium text-ink">{claim.origin.source_title}</p>
              <p className="mt-1 text-sm leading-6 text-ink/76">
                {claim.origin.portion_label ?? claim.origin.source_medium.toLowerCase().replace(/_/g, ' ')}
              </p>
            </div>
            <Link
              to={`/engagements/${claim.origin.engagement_id}`}
              className="rounded-xl bg-white/90 px-3 py-2 text-sm text-ink transition hover:bg-white"
            >
              Open
            </Link>
          </div>
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
