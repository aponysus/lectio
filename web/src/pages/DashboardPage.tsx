import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { type Engagement, getSystemStatus, listEngagements, type SystemStatus } from '../api/client'
import { EngagementCard } from '../components/engagements/EngagementCard'
import { EmptyState } from '../components/shared/EmptyState'

export function DashboardPage() {
  const [status, setStatus] = useState<SystemStatus | null>(null)
  const [recentEngagements, setRecentEngagements] = useState<Engagement[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    ;(async () => {
      try {
        const [nextStatus, nextRecentEngagements] = await Promise.all([
          getSystemStatus(),
          listEngagements({ limit: 4 }),
        ])
        if (!cancelled) {
          setStatus(nextStatus)
          setRecentEngagements(nextRecentEngagements)
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
        <p className="text-xs uppercase tracking-[0.3em] text-accent/80">Inquiry loop</p>
        <h2 className="mt-3 font-display text-4xl text-ink">M4 inquiry workspaces are live</h2>
        <p className="mt-4 max-w-3xl text-base leading-7 text-ink/72">
          The core MVP path now works through inquiry organization: protected sessions, source records, engagement
          capture, inquiry pages, and the ability to attach source work to live questions during capture.
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
          <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Next build step</p>
          <h3 className="mt-3 font-display text-3xl text-ink">Claim extraction is next</h3>
          <ul className="mt-5 space-y-3 text-sm leading-6 text-ink/75">
            <li>Let each engagement produce one to three explicit claims or open questions.</li>
            <li>Connect those claims back to the inquiry workspace they belong to.</li>
            <li>Sharpen reflection without turning capture into bureaucracy.</li>
            <li>Keep synthesis downstream so there is real material to compress.</li>
          </ul>
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
