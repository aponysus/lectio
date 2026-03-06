import { Link } from 'react-router-dom'
import type { Inquiry } from '../../api/client'

export function InquiryListRow({ inquiry }: { inquiry: Inquiry }) {
  return (
    <article className="rounded-[1.25rem] border border-black/5 bg-white/78 px-4 py-4 shadow-card backdrop-blur">
      <div className="grid gap-4 lg:grid-cols-[minmax(0,2.3fr)_0.9fr_1.1fr_1fr_auto] lg:items-center">
        <div className="min-w-0">
          <p className="text-xs uppercase tracking-[0.2em] text-accent/78">{formatStatus(inquiry.status)}</p>
          <h3 className="mt-2 truncate font-display text-[1.35rem] leading-tight text-ink">{inquiry.title}</h3>
          <p className="mt-2 line-clamp-2 text-sm leading-6 text-ink/72">{inquiry.question}</p>
        </div>

        <MetaColumn label="Status" value={formatStatus(inquiry.status)} />
        <MetaColumn
          label="Counts"
          value={`${inquiry.engagement_count} eng • ${inquiry.claim_count} claims • ${inquiry.synthesis_count} synth`}
        />
        <MetaColumn label="Latest activity" value={formatDate(inquiry.latest_activity ?? inquiry.updated_at)} />

        <div className="flex justify-start lg:justify-end">
          <Link
            to={`/inquiries/${inquiry.id}`}
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

function formatStatus(value: string) {
  return value.toLowerCase().replace(/_/g, ' ')
}

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}
