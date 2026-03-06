import { Link } from 'react-router-dom'
import type { Inquiry } from '../../api/client'

export function InquiryCard({ inquiry }: { inquiry: Inquiry }) {
  return (
    <article className="rounded-[1.75rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-accent/80">{formatStatus(inquiry.status)}</p>
          <h3 className="mt-2 font-display text-2xl text-ink">{inquiry.title}</h3>
          <p className="mt-3 line-clamp-3 text-sm leading-6 text-ink/78">{inquiry.question}</p>
        </div>
        <Link
          to={`/inquiries/${inquiry.id}`}
          className="rounded-xl bg-pine px-3 py-2 text-sm text-white transition hover:bg-pine/90"
        >
          Open
        </Link>
      </div>

      <dl className="mt-5 grid gap-4 sm:grid-cols-3">
        <MetaItem label="Engagements" value={String(inquiry.engagement_count)} />
        <MetaItem label="Claims" value={String(inquiry.claim_count)} />
        <MetaItem label="Latest activity" value={formatDate(inquiry.latest_activity ?? inquiry.updated_at)} />
      </dl>

      {inquiry.why_it_matters ? (
        <p className="mt-4 line-clamp-3 text-sm leading-6 text-ink/72">{inquiry.why_it_matters}</p>
      ) : null}
    </article>
  )
}

function MetaItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl bg-black/[0.03] px-4 py-3">
      <dt className="text-xs uppercase tracking-[0.2em] text-accent/70">{label}</dt>
      <dd className="mt-2 text-sm text-ink/80">{value}</dd>
    </div>
  )
}

function formatStatus(value: string) {
  return value.toLowerCase().replace(/_/g, ' ')
}

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}
