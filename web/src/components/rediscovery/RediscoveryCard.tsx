import { Link } from 'react-router-dom'
import type { RediscoveryItem } from '../../api/client'

type RediscoveryCardProps = {
  item: RediscoveryItem
  pendingAction?: 'dismiss' | 'act' | null
  onDismiss: (item: RediscoveryItem) => Promise<void> | void
  onAct: (item: RediscoveryItem) => Promise<void> | void
}

export function RediscoveryCard({ item, pendingAction, onDismiss, onAct }: RediscoveryCardProps) {
  const primaryAction = getPrimaryAction(item)
  const secondaryAction = getSecondaryAction(item)

  return (
    <article className="rounded-[1.5rem] bg-black/[0.03] px-4 py-4">
      <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
        <div className="min-w-0">
          <p className="text-xs uppercase tracking-[0.2em] text-accent/75">{formatKind(item.kind)}</p>
          <h4 className="mt-2 font-display text-2xl text-ink">{getTitle(item)}</h4>
          <p className="mt-2 text-sm leading-6 text-ink/78">{getSummary(item)}</p>
          <p className="mt-3 text-xs uppercase tracking-[0.18em] text-ink/60">{getMeta(item)}</p>
        </div>

        <div className="flex flex-wrap gap-2">
          {primaryAction ? (
            <Link
              to={primaryAction.to}
              className="rounded-xl bg-pine px-3 py-2 text-sm text-white transition hover:bg-pine/90"
            >
              {primaryAction.label}
            </Link>
          ) : null}
          {secondaryAction ? (
            <Link
              to={secondaryAction.to}
              className="rounded-xl border border-black/10 bg-white/80 px-3 py-2 text-sm text-ink transition hover:bg-white"
            >
              {secondaryAction.label}
            </Link>
          ) : null}
          <button
            type="button"
            onClick={() => void onAct(item)}
            disabled={pendingAction !== null}
            className="rounded-xl border border-black/10 bg-white/80 px-3 py-2 text-sm text-ink transition hover:bg-white disabled:cursor-wait disabled:opacity-70"
          >
            {pendingAction === 'act' ? 'Saving...' : 'Mark acted on'}
          </button>
          <button
            type="button"
            onClick={() => void onDismiss(item)}
            disabled={pendingAction !== null}
            className="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 transition hover:bg-red-100 disabled:cursor-wait disabled:opacity-70"
          >
            {pendingAction === 'dismiss' ? 'Dismissing...' : 'Dismiss'}
          </button>
        </div>
      </div>
    </article>
  )
}

function getTitle(item: RediscoveryItem) {
  if (item.claim) {
    return truncate(item.claim.text, 110)
  }
  if (item.engagement) {
    return item.engagement.portion_label ?? item.engagement.source.title
  }
  if (item.inquiry) {
    return item.inquiry.title
  }
  return 'Rediscovery item'
}

function getSummary(item: RediscoveryItem) {
  if (item.kind === 'stale_tentative_claim' && item.claim) {
    const origin = item.claim.origin?.source_title ? ` Origin: ${item.claim.origin.source_title}.` : ''
    return `${item.reason}${origin}`
  }
  if (item.kind === 'active_inquiry_old_engagement' && item.engagement) {
    return `${item.reason} ${truncate(item.engagement.reflection, 160)}`
  }
  if (item.inquiry) {
    return `${item.reason} ${truncate(item.inquiry.question, 150)}`
  }
  return item.reason
}

function getMeta(item: RediscoveryItem) {
  if (item.claim) {
    const inquiryTitle = item.linked_inquiry?.title ? ` • ${item.linked_inquiry.title}` : ''
    return `${item.claim.claim_type.toLowerCase().replace(/_/g, ' ')} • ${item.claim.status.toLowerCase().replace(/_/g, ' ')}${inquiryTitle}`
  }
  if (item.engagement) {
    const revisit = item.engagement.revisit_priority ? `revisit ${item.engagement.revisit_priority}` : 'revisit not set'
    const inquiryTitle = item.linked_inquiry?.title ? ` • ${item.linked_inquiry.title}` : ''
    return `${item.engagement.source.title} • ${revisit}${inquiryTitle}`
  }
  if (item.inquiry) {
    return `${item.inquiry.engagement_count} engagements • ${item.inquiry.claim_count} claims • ${item.inquiry.status.toLowerCase().replace(/_/g, ' ')}`
  }
  return formatTimestamp(item.created_at)
}

function getPrimaryAction(item: RediscoveryItem) {
  if (item.kind === 'stale_tentative_claim') {
    if (item.linked_inquiry) {
      return { to: `/inquiries/${item.linked_inquiry.id}`, label: 'Open inquiry' }
    }
    if (item.claim?.origin_engagement_id) {
      return { to: `/engagements/${item.claim.origin_engagement_id}`, label: 'Open engagement' }
    }
    return null
  }

  if (item.kind === 'active_inquiry_old_engagement' && item.engagement) {
    return { to: `/engagements/${item.engagement.id}`, label: 'Open engagement' }
  }

  if (item.kind === 'unsynthesized_inquiry' && item.inquiry) {
    return { to: `/syntheses/new?inquiryId=${item.inquiry.id}`, label: 'Write synthesis' }
  }

  if (item.inquiry) {
    return { to: `/inquiries/${item.inquiry.id}`, label: 'Open inquiry' }
  }

  return null
}

function getSecondaryAction(item: RediscoveryItem) {
  if (item.kind === 'active_inquiry_old_engagement' && item.linked_inquiry) {
    return { to: `/inquiries/${item.linked_inquiry.id}`, label: 'Open inquiry' }
  }

  if (item.kind === 'unsynthesized_inquiry' && item.inquiry) {
    return { to: `/inquiries/${item.inquiry.id}`, label: 'Open inquiry' }
  }

  return null
}

function formatKind(kind: RediscoveryItem['kind']) {
  return kind.replace(/_/g, ' ')
}

function truncate(value: string, length: number) {
  if (value.length <= length) {
    return value
  }
  return `${value.slice(0, length - 1).trimEnd()}...`
}

function formatTimestamp(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}
