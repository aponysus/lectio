import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import {
  createSynthesis,
  getInquiry,
  getSynthesis,
  listInquiryClaims,
  listInquiryEngagements,
  type Claim,
  type Engagement,
  type Inquiry,
  type Synthesis,
  type SynthesisInput,
  updateSynthesis,
} from '../../api/client'
import { SynthesisForm } from '../../components/syntheses/SynthesisForm'
import { EmptyState } from '../../components/shared/EmptyState'
import { PageHeader } from '../../components/shared/PageHeader'

type SynthesisFormPageProps = {
  mode: 'create' | 'edit'
}

export function SynthesisFormPage({ mode }: SynthesisFormPageProps) {
  const navigate = useNavigate()
  const { synthesisId } = useParams()
  const [searchParams] = useSearchParams()
  const queryInquiryID = searchParams.get('inquiryId') ?? undefined

  const [synthesis, setSynthesis] = useState<Synthesis | null>(null)
  const [inquiry, setInquiry] = useState<Inquiry | null>(null)
  const [claims, setClaims] = useState<Claim[]>([])
  const [engagements, setEngagements] = useState<Engagement[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        if (mode === 'edit' && synthesisId) {
          const nextSynthesis = await getSynthesis(synthesisId)
          const linkedInquiryID = nextSynthesis.inquiry_id
          const [nextInquiry, nextClaims, nextEngagements] = await Promise.all([
            getInquiry(linkedInquiryID),
            listInquiryClaims(linkedInquiryID),
            listInquiryEngagements(linkedInquiryID, 6),
          ])
          if (!cancelled) {
            setSynthesis(nextSynthesis)
            setInquiry(nextInquiry)
            setClaims(nextClaims)
            setEngagements(nextEngagements)
          }
          return
        }

        if (!queryInquiryID) {
          if (!cancelled) {
            setInquiry(null)
            setClaims([])
            setEngagements([])
          }
          return
        }

        const [nextInquiry, nextClaims, nextEngagements] = await Promise.all([
          getInquiry(queryInquiryID),
          listInquiryClaims(queryInquiryID),
          listInquiryEngagements(queryInquiryID, 6),
        ])
        if (!cancelled) {
          setInquiry(nextInquiry)
          setClaims(nextClaims)
          setEngagements(nextEngagements)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load synthesis form')
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
  }, [mode, queryInquiryID, synthesisId])

  const handleSubmit = async (input: SynthesisInput) => {
    setSaving(true)
    setError(null)

    try {
      const saved =
        mode === 'create'
          ? await createSynthesis(input)
          : await updateSynthesis(synthesisId ?? '', input)

      navigate(`/syntheses/${saved.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save synthesis')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <section className="rounded-[2rem] border border-black/5 bg-white/70 px-6 py-8 shadow-card backdrop-blur">
        Loading synthesis form...
      </section>
    )
  }

  if (!inquiry) {
    return (
      <EmptyState
        title="Choose an inquiry first"
        body="In MVP, every synthesis is inquiry-linked. Start from an inquiry page or from a dashboard prompt so the compression stays grounded in a real workspace."
        action={
          <Link
            to="/inquiries"
            className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
          >
            Open inquiries
          </Link>
        }
      />
    )
  }

  const actions =
    mode === 'edit' && synthesisId ? (
      <Link
        to={`/syntheses/${synthesisId}`}
        className="rounded-2xl border border-black/10 bg-white/70 px-4 py-3 text-sm text-ink transition hover:bg-white"
      >
        Back to synthesis
      </Link>
    ) : (
      <Link
        to={`/inquiries/${inquiry.id}`}
        className="rounded-2xl border border-black/10 bg-white/70 px-4 py-3 text-sm text-ink transition hover:bg-white"
      >
        Back to inquiry
      </Link>
    )

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow={mode === 'create' ? 'New synthesis' : 'Edit synthesis'}
        title={mode === 'create' ? 'Compress the inquiry into a sharper current view' : 'Refine the synthesis'}
        description="Use the linked claims and recent engagements as raw material, then write the smallest synthesis that actually changes how the inquiry is understood."
        actions={actions}
      />

      {error ? (
        <section className="rounded-[2rem] border border-amber-200 bg-amber-50 px-6 py-5 text-amber-700 shadow-card">
          {error}
        </section>
      ) : null}

      <SynthesisForm
        synthesis={synthesis}
        inquiry={inquiry}
        submitLabel={mode === 'create' ? 'Create synthesis' : 'Save changes'}
        submitting={saving}
        apiError={error}
        onSubmit={handleSubmit}
      />

      <section className="grid gap-6 lg:grid-cols-2">
        <article className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
          <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Linked claims</p>
          <h3 className="mt-2 font-display text-3xl text-ink">What has been asserted so far</h3>
          {claims.length === 0 ? (
            <p className="mt-4 text-sm leading-7 text-ink/74">No claims are linked yet. This synthesis may still be useful, but the inquiry probably needs sharper extraction.</p>
          ) : (
            <div className="mt-5 space-y-3">
              {claims.slice(0, 5).map((claim) => (
                <div key={claim.id} className="rounded-2xl bg-black/[0.03] px-4 py-4">
                  <p className="text-xs uppercase tracking-[0.2em] text-accent/75">
                    {claim.claim_type.toLowerCase().replace(/_/g, ' ')} • {claim.status.toLowerCase().replace(/_/g, ' ')}
                  </p>
                  <p className="mt-3 text-sm leading-6 text-ink/82">{claim.text}</p>
                </div>
              ))}
            </div>
          )}
        </article>

        <article className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
          <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Recent engagements</p>
          <h3 className="mt-2 font-display text-3xl text-ink">What fed this inquiry</h3>
          {engagements.length === 0 ? (
            <p className="mt-4 text-sm leading-7 text-ink/74">No linked engagements yet. The inquiry exists, but it has not accumulated much raw material.</p>
          ) : (
            <div className="mt-5 space-y-3">
              {engagements.map((engagement) => (
                <div key={engagement.id} className="rounded-2xl bg-black/[0.03] px-4 py-4">
                  <p className="text-xs uppercase tracking-[0.2em] text-accent/75">{formatDateTime(engagement.engaged_at)}</p>
                  <p className="mt-2 text-sm font-medium text-ink">{engagement.source.title}</p>
                  <p className="mt-2 line-clamp-4 text-sm leading-6 text-ink/78">{engagement.reflection}</p>
                </div>
              ))}
            </div>
          )}
        </article>
      </section>
    </div>
  )
}

function formatDateTime(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}
