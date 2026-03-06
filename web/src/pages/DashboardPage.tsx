import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  dismissRediscoveryItem,
  type Engagement,
  type Inquiry,
  listInquiries,
  listRediscoveryItems,
  listEngagements,
  listSources,
  listSynthesisEligibleInquiries,
  markRediscoveryItemActedOn,
  type RediscoveryItem,
} from '../api/client'
import { useToast } from '../components/feedback/ToastProvider'
import { RediscoveryCard } from '../components/rediscovery/RediscoveryCard'
import { EmptyState } from '../components/shared/EmptyState'

export function DashboardPage() {
  const { showToast } = useToast()
  const [recentEngagements, setRecentEngagements] = useState<Engagement[]>([])
  const [activeInquiries, setActiveInquiries] = useState<Inquiry[]>([])
  const [eligibleInquiries, setEligibleInquiries] = useState<Inquiry[]>([])
  const [rediscoveryItems, setRediscoveryItems] = useState<RediscoveryItem[]>([])
  const [sourceCount, setSourceCount] = useState(0)
  const [pendingRediscoveryID, setPendingRediscoveryID] = useState<string | null>(null)
  const [pendingRediscoveryAction, setPendingRediscoveryAction] = useState<'dismiss' | 'act' | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    ;(async () => {
      try {
        const [nextSources, nextActiveInquiries, nextRecentEngagements, nextEligibleInquiries, nextRediscoveryItems] = await Promise.all([
          listSources({ limit: 100, sort: 'recent' }),
          listInquiries({ status: 'ACTIVE', limit: 6 }),
          listEngagements({ limit: 4 }),
          listSynthesisEligibleInquiries(4),
          listRediscoveryItems(6),
        ])
        if (!cancelled) {
          setSourceCount(nextSources.length)
          setActiveInquiries(nextActiveInquiries)
          setRecentEngagements(nextRecentEngagements)
          setEligibleInquiries(nextEligibleInquiries)
          setRediscoveryItems(nextRediscoveryItems)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load dashboard')
        }
      }
    })()

    return () => {
      cancelled = true
    }
  }, [])

  const handleDismissRediscoveryItem = async (item: RediscoveryItem) => {
    setPendingRediscoveryID(item.id)
    setPendingRediscoveryAction('dismiss')
    setError(null)

    try {
      await dismissRediscoveryItem(item.id)
      setRediscoveryItems((current) => current.filter((entry) => entry.id !== item.id))
      showToast({ message: 'Rediscovery item dismissed.', tone: 'info' })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to dismiss rediscovery item')
    } finally {
      setPendingRediscoveryID(null)
      setPendingRediscoveryAction(null)
    }
  }

  const handleActOnRediscoveryItem = async (item: RediscoveryItem) => {
    setPendingRediscoveryID(item.id)
    setPendingRediscoveryAction('act')
    setError(null)

    try {
      await markRediscoveryItemActedOn(item.id)
      setRediscoveryItems((current) => current.filter((entry) => entry.id !== item.id))
      showToast({ message: 'Rediscovery item marked acted on.' })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update rediscovery item')
    } finally {
      setPendingRediscoveryID(null)
      setPendingRediscoveryAction(null)
    }
  }

  return (
    <div className="space-y-6">
      <section className="rounded-[1.5rem] border border-black/5 bg-white/70 px-5 py-6 shadow-card backdrop-blur lg:px-6">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div className="max-w-3xl">
            <p className="text-[0.7rem] uppercase tracking-[0.24em] text-accent/80">Today&apos;s desk</p>
            <h2 className="mt-2 font-display text-[2.25rem] leading-tight text-ink">Resume the work that still matters</h2>
            <p className="mt-3 text-sm leading-6 text-ink/72">
              Pick up an active inquiry, log the next engagement, and surface what is ready for synthesis or another pass.
            </p>
          </div>
          <div className="flex flex-wrap gap-3">
            <Link
              to="/engagements/new"
              className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
            >
              Log engagement
            </Link>
            <Link
              to="/inquiries/new"
              className="rounded-2xl border border-black/10 bg-white/90 px-4 py-3 text-sm text-ink transition hover:bg-white"
            >
              New inquiry
            </Link>
            <Link
              to="/sources/new"
              className="rounded-2xl border border-black/10 bg-white/90 px-4 py-3 text-sm text-ink transition hover:bg-white"
            >
              New source
            </Link>
          </div>
        </div>
      </section>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatusCard title="Active inquiries" value={String(activeInquiries.length)} tone="pine" />
        <StatusCard title="Sources" value={String(sourceCount)} tone="accent" />
        <StatusCard title="Ready for synthesis" value={String(eligibleInquiries.length)} tone="ink" />
        <StatusCard title="Resurfacing now" value={String(rediscoveryItems.length)} tone="neutral" />
      </section>

      <section className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
        <article className="rounded-[2rem] border border-black/5 bg-white/70 px-6 py-7 shadow-card backdrop-blur">
          <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Continue where you left off</p>
          <h3 className="mt-3 font-display text-3xl text-ink">Recent work and live questions</h3>

          <div className="mt-5 grid gap-6 lg:grid-cols-2">
            <div>
              <p className="text-xs uppercase tracking-[0.18em] text-ink/60">Recent engagements</p>
              {recentEngagements.length === 0 ? (
                <p className="mt-3 text-sm leading-6 text-ink/72">No engagements yet. Start by logging the next serious encounter.</p>
              ) : (
                <div className="mt-3 space-y-3">
                  {recentEngagements.slice(0, 3).map((engagement) => (
                    <WorkspaceLinkCard
                      key={engagement.id}
                      to={`/engagements/${engagement.id}`}
                      label={engagement.source.title}
                      title={engagement.portion_label ?? 'Untitled engagement'}
                      meta={formatTimestamp(engagement.engaged_at)}
                    />
                  ))}
                </div>
              )}
            </div>

            <div>
              <p className="text-xs uppercase tracking-[0.18em] text-ink/60">Active inquiries</p>
              {activeInquiries.length === 0 ? (
                <p className="mt-3 text-sm leading-6 text-ink/72">No active inquiries yet. Open one before the archive starts drifting.</p>
              ) : (
                <div className="mt-3 space-y-3">
                  {activeInquiries.slice(0, 3).map((inquiry) => (
                    <WorkspaceLinkCard
                      key={inquiry.id}
                      to={`/inquiries/${inquiry.id}`}
                      label={`${inquiry.engagement_count} engagements • ${inquiry.claim_count} claims`}
                      title={inquiry.title}
                      meta={inquiry.latest_activity ? `Latest activity ${formatTimestamp(inquiry.latest_activity)}` : 'No recent activity'}
                    />
                  ))}
                </div>
              )}
            </div>
          </div>
        </article>

        <article className="rounded-[2rem] border border-black/5 bg-white/70 px-6 py-7 shadow-card backdrop-blur">
          <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Synthesis prompts</p>
          <h3 className="mt-3 font-display text-3xl text-ink">Inquiries ready for compression</h3>
          {eligibleInquiries.length === 0 ? (
            <p className="mt-5 text-sm leading-7 text-ink/74">
              Nothing is at the synthesis threshold right now. Once an inquiry reaches enough density, it will surface here.
            </p>
          ) : (
            <div className="mt-5 space-y-4">
              {eligibleInquiries.map((inquiry) => (
                <SynthesisPromptCard key={inquiry.id} inquiry={inquiry} />
              ))}
            </div>
          )}
        </article>
      </section>

      {error ? (
        <section className="rounded-[1.5rem] border border-amber-200 bg-amber-50 px-5 py-4 text-amber-700 shadow-card">
          {error}
        </section>
      ) : null}

      <section className="rounded-[2rem] border border-black/5 bg-white/70 px-6 py-7 shadow-card backdrop-blur">
        <div className="flex items-center justify-between gap-4">
          <div>
            <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Bring back into view</p>
            <h3 className="mt-3 font-display text-3xl text-ink">Sparse prompts for work worth resurfacing</h3>
          </div>
          <p className="max-w-sm text-sm leading-6 text-ink/70">
            The feed stays intentionally small: stale tentative claims, older inquiry-linked engagements, unsynthesized
            inquiry clusters, and recent reactivations.
          </p>
        </div>

        {rediscoveryItems.length === 0 ? (
          <p className="mt-5 text-sm leading-7 text-ink/74">
            Nothing currently needs resurfacing. Once material crosses the v1 thresholds, it will appear here.
          </p>
        ) : (
          <div className="mt-5 space-y-4">
            {rediscoveryItems.map((item) => (
              <RediscoveryCard
                key={item.id}
                item={item}
                pendingAction={pendingRediscoveryID === item.id ? pendingRediscoveryAction : null}
                onDismiss={handleDismissRediscoveryItem}
                onAct={handleActOnRediscoveryItem}
              />
            ))}
          </div>
        )}
      </section>

      {recentEngagements.length === 0 ? (
        <EmptyState
          title="No engagements yet"
          body="Start with one deliberate encounter. Source records are ready; the next useful move is to log what you actually engaged."
          action={
            <Link
              to="/engagements/new"
              className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
            >
              Log first engagement
            </Link>
          }
        />
      ) : null}
    </div>
  )
}

function SynthesisPromptCard({ inquiry }: { inquiry: Inquiry }) {
  return (
    <div className="rounded-[1.5rem] bg-black/[0.03] px-4 py-4">
      <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div className="min-w-0">
          <p className="text-xs uppercase tracking-[0.2em] text-accent/75">{inquiry.status.toLowerCase().replace(/_/g, ' ')}</p>
          <h4 className="mt-2 font-display text-2xl text-ink">{inquiry.title}</h4>
          <p className="mt-2 line-clamp-3 text-sm leading-6 text-ink/78">{inquiry.question}</p>
          <p className="mt-3 text-xs uppercase tracking-[0.18em] text-ink/60">
            {inquiry.engagement_count} engagements • {inquiry.claim_count} claims
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Link
            to={`/syntheses/new?inquiryId=${inquiry.id}`}
            className="rounded-xl bg-pine px-3 py-2 text-sm text-white transition hover:bg-pine/90"
          >
            Write synthesis
          </Link>
          <Link
            to={`/inquiries/${inquiry.id}`}
            className="rounded-xl border border-black/10 bg-white/80 px-3 py-2 text-sm text-ink transition hover:bg-white"
          >
            Open inquiry
          </Link>
        </div>
      </div>
    </div>
  )
}

function StatusCard({
  title,
  value,
  tone,
}: {
  title: string
  value: string
  tone: 'pine' | 'accent' | 'ink' | 'neutral'
}) {
  const toneClasses = {
    pine: 'bg-pine text-white',
    accent: 'bg-accent text-white',
    ink: 'bg-ink text-white',
    neutral: 'bg-white/70 text-ink',
  }

  return (
    <article className={`rounded-[1.75rem] px-5 py-5 shadow-card ${toneClasses[tone]}`}>
      <p className="text-xs uppercase tracking-[0.2em] opacity-75">{title}</p>
      <p className="mt-3 text-lg font-medium">{value}</p>
    </article>
  )
}

function WorkspaceLinkCard({
  to,
  label,
  title,
  meta,
}: {
  to: string
  label: string
  title: string
  meta: string
}) {
  return (
    <Link
      to={to}
      className="block rounded-[1.25rem] border border-black/5 bg-black/[0.03] px-4 py-4 transition hover:border-black/10 hover:bg-black/[0.045]"
    >
      <p className="text-xs uppercase tracking-[0.18em] text-accent/75">{label}</p>
      <h4 className="mt-2 text-base font-medium text-ink">{title}</h4>
      <p className="mt-2 text-sm leading-6 text-ink/72">{meta}</p>
    </Link>
  )
}

function formatTimestamp(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}
