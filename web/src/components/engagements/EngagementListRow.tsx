import { Link } from 'react-router-dom'
import type { Engagement } from '../../api/client'

export function EngagementListRow({ engagement }: { engagement: Engagement }) {
  return (
    <article className="rounded-[1.25rem] border border-black/5 bg-white/78 px-4 py-4 shadow-card backdrop-blur">
      <div className="grid gap-4 lg:grid-cols-[minmax(0,2.2fr)_1fr_1fr_1fr_auto] lg:items-center">
        <div className="min-w-0">
          <p className="text-xs uppercase tracking-[0.2em] text-accent/78">{formatDateTime(engagement.engaged_at)}</p>
          <h3 className="mt-2 truncate font-display text-[1.35rem] leading-tight text-ink">
            {engagement.portion_label ?? 'Untitled engagement'}
          </h3>
          <p className="mt-2 truncate text-sm text-ink/72">
            {engagement.source.title} • {formatMedium(engagement.source.medium)}
          </p>
          <p className="mt-2 line-clamp-1 text-sm leading-6 text-ink/68">{engagement.reflection}</p>
        </div>

        <MetaColumn label="Access" value={engagement.access_mode ? formatMedium(engagement.access_mode) : 'Not set'} />
        <MetaColumn
          label="Revisit"
          value={engagement.revisit_priority ? String(engagement.revisit_priority) : 'Not set'}
        />
        <MetaColumn
          label="Notes"
          value={`${engagement.language_note_count} ${engagement.language_note_count === 1 ? 'language note' : 'language notes'}`}
        />

        <div className="flex justify-start lg:justify-end">
          <Link
            to={`/engagements/${engagement.id}`}
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

function formatDateTime(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

function formatMedium(value: string) {
  return value.toLowerCase().replace(/_/g, ' ')
}
