import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { archiveInquiry, getInquiry, listInquiryEngagements, type Engagement, type Inquiry } from '../../api/client'
import { EngagementCard } from '../../components/engagements/EngagementCard'
import { EmptyState } from '../../components/shared/EmptyState'
import { PageHeader } from '../../components/shared/PageHeader'

export function InquiryDetailPage() {
  const navigate = useNavigate()
  const { inquiryId } = useParams()
  const [inquiry, setInquiry] = useState<Inquiry | null>(null)
  const [engagements, setEngagements] = useState<Engagement[]>([])
  const [loading, setLoading] = useState(true)
  const [archivePending, setArchivePending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!inquiryId) {
      setError('Inquiry id is missing')
      setLoading(false)
      return
    }

    let cancelled = false
    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        const [nextInquiry, nextEngagements] = await Promise.all([
          getInquiry(inquiryId),
          listInquiryEngagements(inquiryId, 20),
        ])
        if (!cancelled) {
          setInquiry(nextInquiry)
          setEngagements(nextEngagements)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load inquiry')
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
  }, [inquiryId])

  const handleArchive = async () => {
    if (!inquiry || archivePending) {
      return
    }

    const confirmed = window.confirm(`Archive "${inquiry.title}"?`)
    if (!confirmed) {
      return
    }

    setArchivePending(true)
    setError(null)

    try {
      await archiveInquiry(inquiry.id)
      navigate('/inquiries')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to archive inquiry')
      setArchivePending(false)
    }
  }

  if (loading) {
    return (
      <section className="rounded-[2rem] border border-black/5 bg-white/70 px-6 py-8 shadow-card backdrop-blur">
        Loading inquiry...
      </section>
    )
  }

  if (!inquiry) {
    return (
      <EmptyState
        title="Inquiry not found"
        body={error ?? 'This inquiry could not be loaded.'}
        action={
          <Link
            to="/inquiries"
            className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
          >
            Back to inquiries
          </Link>
        }
      />
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow={inquiry.status.toLowerCase().replace(/_/g, ' ')}
        title={inquiry.title}
        description={inquiry.question}
        actions={
          <>
            <Link
              to={`/inquiries/${inquiry.id}/edit`}
              className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
            >
              Edit inquiry
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

      <section className="grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
        <article className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
          <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Inquiry context</p>
          <div className="mt-5 space-y-4">
            <ContextBlock label="Why it matters" value={inquiry.why_it_matters} empty="Not set yet." />
            <ContextBlock label="Current view" value={inquiry.current_view} empty="Still taking shape." />
            <ContextBlock label="Open tensions" value={inquiry.open_tensions} empty="No tensions recorded yet." />
          </div>
        </article>

        <article className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
          <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Workspace status</p>
          <dl className="mt-5 space-y-4">
            <MetaItem label="Status" value={inquiry.status.toLowerCase().replace(/_/g, ' ')} />
            <MetaItem label="Engagements" value={String(inquiry.engagement_count)} />
            <MetaItem label="Claims" value={String(inquiry.claim_count)} />
            <MetaItem label="Latest activity" value={formatDateTime(inquiry.latest_activity ?? inquiry.updated_at)} />
          </dl>
        </article>
      </section>

      {engagements.length === 0 ? (
        <EmptyState
          title="No linked engagements yet"
          body="Attach this inquiry during engagement capture to turn it into a real workspace with evidence, tensions, and later claims."
          action={
            <Link
              to="/engagements/new"
              className="rounded-2xl bg-accent px-4 py-3 text-sm font-medium text-white transition hover:bg-accent/90"
            >
              Log an engagement
            </Link>
          }
        />
      ) : (
        <section className="space-y-4">
          <div>
            <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Linked engagements</p>
            <h3 className="mt-2 font-display text-3xl text-ink">The source work feeding this inquiry</h3>
          </div>
          <div className="grid gap-5 xl:grid-cols-2">
            {engagements.map((engagement) => (
              <EngagementCard key={engagement.id} engagement={engagement} />
            ))}
          </div>
        </section>
      )}

      <section className="grid gap-6 lg:grid-cols-2">
        <PlaceholderSection
          eyebrow="Claims"
          title="Claim extraction arrives next"
          body="M5 will turn this inquiry into a sharper workspace by letting you extract and revise claims directly from engagement reflection."
        />
        <PlaceholderSection
          eyebrow="Syntheses"
          title="Synthesis comes after claims"
          body="M6 will add the compression layer. This page is already usable without it because the inquiry and its linked engagements are live now."
        />
      </section>
    </div>
  )
}

function ContextBlock({ label, value, empty }: { label: string; value?: string; empty: string }) {
  return (
    <div className="rounded-2xl bg-black/[0.03] px-4 py-4">
      <p className="text-xs uppercase tracking-[0.2em] text-accent/75">{label}</p>
      <p className="mt-3 whitespace-pre-wrap text-sm leading-6 text-ink/80">{value ?? empty}</p>
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

function PlaceholderSection({
  eyebrow,
  title,
  body,
}: {
  eyebrow: string
  title: string
  body: string
}) {
  return (
    <article className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
      <p className="text-xs uppercase tracking-[0.25em] text-accent/80">{eyebrow}</p>
      <h3 className="mt-3 font-display text-3xl text-ink">{title}</h3>
      <p className="mt-4 text-sm leading-7 text-ink/76">{body}</p>
    </article>
  )
}

function formatDateTime(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}
