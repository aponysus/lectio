import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  type Engagement,
  getSystemStatus,
  type Inquiry,
  listEngagements,
  listSynthesisEligibleInquiries,
  type SystemStatus,
} from '../api/client'
import { EngagementCard } from '../components/engagements/EngagementCard'
import { EmptyState } from '../components/shared/EmptyState'

export function DashboardPage() {
  const [status, setStatus] = useState<SystemStatus | null>(null)
  const [recentEngagements, setRecentEngagements] = useState<Engagement[]>([])
  const [eligibleInquiries, setEligibleInquiries] = useState<Inquiry[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    ;(async () => {
      try {
        const [nextStatus, nextRecentEngagements, nextEligibleInquiries] = await Promise.all([
          getSystemStatus(),
          listEngagements({ limit: 4 }),
          listSynthesisEligibleInquiries(4),
        ])
        if (!cancelled) {
          setStatus(nextStatus)
          setRecentEngagements(nextRecentEngagements)
          setEligibleInquiries(nextEligibleInquiries)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load system status')
        }
      }
    })()

    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="space-y-6">
      <section className="rounded-[2rem] border border-black/5 bg-white/70 px-6 py-8 shadow-card backdrop-blur lg:px-8">
        <p className="text-xs uppercase tracking-[0.3em] text-accent/80">Sharpening loop</p>
        <h2 className="mt-3 font-display text-4xl text-ink">M6 synthesis is live</h2>
        <p className="mt-4 max-w-3xl text-base leading-7 text-ink/72">
          The core MVP path now moves past collection into sharper thinking: protected sessions, source records,
          engagement capture, inquiry workspaces, explicit claims extracted from reflection, and inquiry-linked
          syntheses that compress what the work now seems to say.
        </p>
      </section>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatusCard title="App" value={status?.app_name ?? 'Loading...'} tone="pine" />
        <StatusCard title="Environment" value={status?.environment ?? 'Loading...'} tone="accent" />
        <StatusCard title="Migrations" value={status ? String(status.applied_migrations) : 'Loading...'} tone="ink" />
        <StatusCard title="Database Time" value={status?.database_time ?? 'Loading...'} tone="neutral" />
      </section>

      <section className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
        <article className="rounded-[2rem] border border-black/5 bg-white/70 px-6 py-7 shadow-card backdrop-blur">
          <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Synthesis prompts</p>
          <h3 className="mt-3 font-display text-3xl text-ink">Inquiries ready for compression</h3>
          {eligibleInquiries.length === 0 ? (
            <p className="mt-5 text-sm leading-7 text-ink/74">
              Nothing is over the synthesis threshold right now. Once an inquiry reaches three linked engagements or
              two linked claims without a synthesis, it will surface here.
            </p>
          ) : (
            <div className="mt-5 space-y-4">
              {eligibleInquiries.map((inquiry) => (
                <SynthesisPromptCard key={inquiry.id} inquiry={inquiry} />
              ))}
            </div>
          )}
        </article>

        <article className="rounded-[2rem] border border-black/5 bg-stone-950 px-6 py-7 text-stone-100 shadow-card">
          <p className="text-xs uppercase tracking-[0.25em] text-stone-400">System status</p>
          {error ? (
            <p className="mt-4 text-sm leading-6 text-amber-300">{error}</p>
          ) : (
            <dl className="mt-4 space-y-4 text-sm">
              <div>
                <dt className="text-stone-400">Bootstrapped At</dt>
                <dd className="mt-1 text-stone-100">{status?.bootstrapped_at ?? 'Loading...'}</dd>
              </div>
              <div>
                <dt className="text-stone-400">Database Clock</dt>
                <dd className="mt-1 text-stone-100">{status?.database_time ?? 'Loading...'}</dd>
              </div>
              <div>
                <dt className="text-stone-400">Authenticated Surface</dt>
                <dd className="mt-1 text-stone-100">Dashboard shell, protected API route, and session-aware nav</dd>
              </div>
            </dl>
          )}
        </article>
      </section>

      {recentEngagements.length === 0 ? (
        <EmptyState
          title="No engagements yet"
          body="The source slice is stable and the engagement flow is now live. Log the first meaningful encounter with a source from here or from any source detail page."
          action={
            <Link
              to="/engagements/new"
              className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
            >
              Log first engagement
            </Link>
          }
        />
      ) : (
        <section className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Recent engagements</p>
              <h3 className="mt-2 font-display text-3xl text-ink">The latest captured work</h3>
            </div>
            <Link
              to="/engagements/new"
              className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
            >
              New engagement
            </Link>
          </div>
          <div className="grid gap-5 xl:grid-cols-2">
            {recentEngagements.map((engagement) => (
              <EngagementCard key={engagement.id} engagement={engagement} />
            ))}
          </div>
        </section>
      )}
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
