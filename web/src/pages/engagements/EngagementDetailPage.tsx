import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import {
  archiveEngagement,
  getEngagement,
  listEngagementClaims,
  listEngagementInquiries,
  type Claim,
  type Engagement,
  type InquirySummary,
} from '../../api/client'
import { ClaimCard } from '../../components/claims/ClaimCard'
import { EmptyState } from '../../components/shared/EmptyState'
import { PageHeader } from '../../components/shared/PageHeader'

export function EngagementDetailPage() {
  const navigate = useNavigate()
  const { engagementId } = useParams()
  const [engagement, setEngagement] = useState<Engagement | null>(null)
  const [claims, setClaims] = useState<Claim[]>([])
  const [inquiries, setInquiries] = useState<InquirySummary[]>([])
  const [loading, setLoading] = useState(true)
  const [archivePending, setArchivePending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!engagementId) {
      setError('Engagement id is missing')
      setLoading(false)
      return
    }

    let cancelled = false
    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        const [nextEngagement, nextInquiries, nextClaims] = await Promise.all([
          getEngagement(engagementId),
          listEngagementInquiries(engagementId),
          listEngagementClaims(engagementId),
        ])
        if (!cancelled) {
          setEngagement(nextEngagement)
          setInquiries(nextInquiries)
          setClaims(nextClaims)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load engagement')
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    })()

    return () => {
      cancelled = true
    }
  }, [engagementId])

  const handleArchive = async () => {
    if (!engagement || archivePending) {
      return
    }

    const confirmed = window.confirm('Archive this engagement?')
    if (!confirmed) {
      return
    }

    setArchivePending(true)
    setError(null)

    try {
      await archiveEngagement(engagement.id)
      navigate('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to archive engagement')
      setArchivePending(false)
    }
  }

  if (loading) {
    return (
      <section className="rounded-[2rem] border border-black/5 bg-white/70 px-6 py-8 shadow-card backdrop-blur">
        Loading engagement...
      </section>
    )
  }

  if (!engagement) {
    return (
      <EmptyState
        title="Engagement not found"
        body={error ?? 'This engagement could not be loaded.'}
        action={
          <Link
            to="/"
            className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
          >
            Back to dashboard
          </Link>
        }
      />
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow={formatDateTime(engagement.engaged_at)}
        title={engagement.portion_label ?? 'Engagement detail'}
        description={`${engagement.source.title} • ${engagement.source.medium.toLowerCase().replace(/_/g, ' ')}`}
        actions={
          <>
            <Link
              to={`/sources/${engagement.source.id}`}
              className="rounded-2xl border border-black/10 bg-white/70 px-4 py-3 text-sm text-ink transition hover:bg-white"
            >
              Open source
            </Link>
            <Link
              to={`/engagements/${engagement.id}/edit`}
              className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
            >
              Edit engagement
            </Link>
            <button
              type="button"
              onClick={() => void handleArchive()}
              disabled={archivePending}
              className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 transition hover:bg-red-100 disabled:cursor-wait disabled:opacity-70"
            >
              {archivePending ? 'Archiving...' : 'Archive'}
            </button>
          </>
        }
      />

      {error ? (
        <section className="rounded-[2rem] border border-amber-200 bg-amber-50 px-6 py-5 text-amber-700 shadow-card">
          {error}
        </section>
      ) : null}

      <section className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
        <article className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
          <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Reflection</p>
          <p className="mt-4 whitespace-pre-wrap text-sm leading-7 text-ink/82">{engagement.reflection}</p>

          {engagement.why_it_matters ? (
            <div className="mt-6 rounded-2xl bg-black/[0.03] px-4 py-4">
              <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Why it matters</p>
              <p className="mt-3 whitespace-pre-wrap text-sm leading-6 text-ink/80">{engagement.why_it_matters}</p>
            </div>
          ) : null}
        </article>

        <article className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
          <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Metadata</p>
          <dl className="mt-5 space-y-4">
            <MetaItem label="Source" value={engagement.source.title} />
            <MetaItem label="Source language" value={engagement.source_language ?? 'Not set'} />
            <MetaItem label="Reflection language" value={engagement.reflection_language ?? 'Not set'} />
            <MetaItem label="Access mode" value={engagement.access_mode?.toLowerCase().replace(/_/g, ' ') ?? 'Not set'} />
            <MetaItem
              label="Revisit priority"
              value={engagement.revisit_priority ? String(engagement.revisit_priority) : 'Not set'}
            />
            <MetaItem
              label="Reread / rewatch"
              value={engagement.is_reread_or_rewatch ? 'Yes' : 'No'}
            />
            <MetaItem label="Updated" value={formatDateTime(engagement.updated_at)} />
          </dl>

          <div className="mt-6 rounded-2xl bg-black/[0.03] px-4 py-4">
            <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Linked inquiries</p>
            {inquiries.length === 0 ? (
              <p className="mt-3 text-sm leading-6 text-ink/78">No inquiries linked yet.</p>
            ) : (
              <div className="mt-3 flex flex-wrap gap-2">
                {inquiries.map((inquiry) => (
                  <Link
                    key={inquiry.id}
                    to={`/inquiries/${inquiry.id}`}
                    className="rounded-full bg-white/90 px-3 py-2 text-sm text-ink transition hover:bg-white"
                  >
                    {inquiry.title}
                  </Link>
                ))}
              </div>
            )}
          </div>
        </article>
      </section>

      {claims.length === 0 ? (
        <section className="rounded-[2rem] border border-dashed border-black/10 bg-white/55 px-6 py-8 text-center shadow-card backdrop-blur">
          <h3 className="font-display text-3xl text-ink">No claims from this engagement yet</h3>
          <p className="mx-auto mt-4 max-w-2xl text-base leading-7 text-ink/70">
            Claims arrive during capture now. Edit the engagement if this reflection is ready to sharpen into one to
            three explicit propositions or questions.
          </p>
        </section>
      ) : (
        <section className="space-y-4">
          <div>
            <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Claims</p>
            <h3 className="mt-2 font-display text-3xl text-ink">Takeaways extracted from this engagement</h3>
          </div>
          <div className="grid gap-5 xl:grid-cols-2">
            {claims.map((claim) => (
              <ClaimCard key={claim.id} claim={claim} />
            ))}
          </div>
        </section>
      )}
    </div>
  )
}

function MetaItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl bg-black/[0.03] px-4 py-4">
      <dt className="text-xs uppercase tracking-[0.2em] text-accent/75">{label}</dt>
      <dd className="mt-2 text-sm leading-6 text-ink/80">{value}</dd>
    </div>
  )
}

function formatDateTime(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}
