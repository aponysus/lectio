import { Link } from 'react-router-dom'
import type { Engagement } from '../../api/client'

type EngagementCardProps = {
  engagement: Engagement
  showSource?: boolean
}

export function EngagementCard({ engagement, showSource = true }: EngagementCardProps) {
  return (
    <article className="rounded-[1.75rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-accent/80">{formatDateTime(engagement.engaged_at)}</p>
          <h3 className="mt-2 font-display text-2xl text-ink">
            {engagement.portion_label ?? 'Untitled engagement'}
          </h3>
          {showSource ? (
            <p className="mt-2 text-sm leading-6 text-ink/72">
              {engagement.source.title} • {engagement.source.medium.toLowerCase().replace(/_/g, ' ')}
            </p>
          ) : null}
        </div>
        <Link
          to={`/engagements/${engagement.id}`}
          className="rounded-xl bg-pine px-3 py-2 text-sm text-white transition hover:bg-pine/90"
        >
          Open
        </Link>
      </div>

      <p className="mt-4 line-clamp-4 whitespace-pre-wrap text-sm leading-6 text-ink/80">{engagement.reflection}</p>

      <dl className="mt-5 grid gap-4 sm:grid-cols-2">
        <MetaItem label="Access mode" value={engagement.access_mode?.toLowerCase().replace(/_/g, ' ') ?? 'Not set'} />
        <MetaItem
          label="Revisit priority"
          value={engagement.revisit_priority ? String(engagement.revisit_priority) : 'Not set'}
        />
      </dl>
    </article>
  )
}

function MetaItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl bg-black/[0.03] px-4 py-3">
      <dt className="text-xs uppercase tracking-[0.2em] text-accent/75">{label}</dt>
      <dd className="mt-2 text-sm text-ink/78">{value}</dd>
    </div>
  )
}

function formatDateTime(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}
