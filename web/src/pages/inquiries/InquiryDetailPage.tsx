import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import {
  archiveClaim,
  archiveInquiry,
  listInquirySyntheses,
  listInquiryClaims,
  getInquiry,
  listInquiryEngagements,
  type Claim,
  type ClaimUpdateInput,
  type Engagement,
  type Inquiry,
  type Synthesis,
  updateClaim,
} from '../../api/client'
import { ClaimCard } from '../../components/claims/ClaimCard'
import { EngagementCard } from '../../components/engagements/EngagementCard'
import { SynthesisCard } from '../../components/syntheses/SynthesisCard'
import { EmptyState } from '../../components/shared/EmptyState'
import { PageHeader } from '../../components/shared/PageHeader'

export function InquiryDetailPage() {
  const navigate = useNavigate()
  const { inquiryId } = useParams()
  const [inquiry, setInquiry] = useState<Inquiry | null>(null)
  const [claims, setClaims] = useState<Claim[]>([])
  const [engagements, setEngagements] = useState<Engagement[]>([])
  const [syntheses, setSyntheses] = useState<Synthesis[]>([])
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
        const [nextInquiry, nextClaims, nextEngagements, nextSyntheses] = await Promise.all([
          getInquiry(inquiryId),
          listInquiryClaims(inquiryId),
          listInquiryEngagements(inquiryId, 20),
          listInquirySyntheses(inquiryId, 20),
        ])
        if (!cancelled) {
          setInquiry(nextInquiry)
          setClaims(nextClaims)
          setEngagements(nextEngagements)
          setSyntheses(nextSyntheses)
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

  const synthesisReady = inquiry.engagement_count >= 3 || inquiry.claim_count >= 2
  const latestSynthesis = syntheses[0]
  const priorSyntheses = syntheses.slice(1)

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
            <MetaItem label="Syntheses" value={String(inquiry.synthesis_count)} />
            <MetaItem label="Latest activity" value={formatDateTime(inquiry.latest_activity ?? inquiry.updated_at)} />
          </dl>
        </article>
      </section>

      {claims.length === 0 ? (
        <section className="rounded-[2rem] border border-dashed border-black/10 bg-white/55 px-6 py-8 text-center shadow-card backdrop-blur">
          <h3 className="font-display text-3xl text-ink">No claims linked yet</h3>
          <p className="mx-auto mt-4 max-w-2xl text-base leading-7 text-ink/70">
            The inquiry workspace is live. The next sharpening move is to log an engagement and extract one to three
            claims from it.
          </p>
        </section>
      ) : (
        <section className="space-y-4">
          <div>
            <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Claims</p>
            <h3 className="mt-2 font-display text-3xl text-ink">Working propositions and open questions</h3>
          </div>
          <div className="space-y-4">
            {claims.map((claim) => (
              <EditableClaimCard
                key={claim.id}
                claim={claim}
                onSave={async (nextInput) => {
                  const updated = await updateClaim(claim.id, nextInput)
                  setClaims((current) => current.map((currentClaim) => (currentClaim.id === updated.id ? updated : currentClaim)))
                }}
                onArchive={async () => {
                  await archiveClaim(claim.id)
                  setClaims((current) => current.filter((currentClaim) => currentClaim.id !== claim.id))
                }}
              />
            ))}
          </div>
        </section>
      )}

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

      {syntheses.length === 0 ? (
        <article className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div className="max-w-3xl">
              <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Syntheses</p>
              <h3 className="mt-2 font-display text-3xl text-ink">
                {synthesisReady ? 'This inquiry is ready for synthesis' : 'No synthesis yet'}
              </h3>
              <p className="mt-4 text-sm leading-7 text-ink/76">
                {synthesisReady
                  ? 'The inquiry has enough density to justify compression. Write a synthesis that states the current view, names the unresolved tension, and changes what should happen next.'
                  : 'Synthesis is live, but this inquiry may still need more material. Once it reaches three linked engagements or two linked claims, the dashboard will prompt it too.'}
              </p>
            </div>
            <Link
              to={`/syntheses/new?inquiryId=${inquiry.id}`}
              className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
            >
              Write synthesis
            </Link>
          </div>
        </article>
      ) : (
        <section className="space-y-4">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Syntheses</p>
              <h3 className="mt-2 font-display text-3xl text-ink">Compression attempts tied to this inquiry</h3>
            </div>
            <Link
              to={`/syntheses/new?inquiryId=${inquiry.id}`}
              className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
            >
              New synthesis
            </Link>
          </div>

          {latestSynthesis ? (
            <div className="space-y-3">
              <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Latest synthesis</p>
              <SynthesisCard synthesis={latestSynthesis} />
            </div>
          ) : null}

          {priorSyntheses.length > 0 ? (
            <div className="space-y-3">
              <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Earlier syntheses</p>
              <div className="grid gap-5 xl:grid-cols-2">
                {priorSyntheses.map((synthesis) => (
                  <SynthesisCard key={synthesis.id} synthesis={synthesis} />
                ))}
              </div>
            </div>
          ) : null}
        </section>
      )}
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

function formatDateTime(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

function EditableClaimCard({
  claim,
  onSave,
  onArchive,
}: {
  claim: Claim
  onSave: (input: ClaimUpdateInput) => Promise<void>
  onArchive: () => Promise<void>
}) {
  const [editing, setEditing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [archivePending, setArchivePending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [draft, setDraft] = useState(() => toClaimDraft(claim))

  useEffect(() => {
    setDraft(toClaimDraft(claim))
  }, [claim])

  if (!editing) {
    return (
      <ClaimCard
        claim={claim}
        actions={
          <>
            <button
              type="button"
              onClick={() => {
                setEditing(true)
                setError(null)
              }}
              className="rounded-xl bg-pine px-3 py-2 text-sm text-white transition hover:bg-pine/90"
            >
              Edit
            </button>
            <button
              type="button"
              disabled={archivePending}
              onClick={() => void handleArchive()}
              className="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 transition hover:bg-red-100 disabled:cursor-wait disabled:opacity-70"
            >
              {archivePending ? 'Archiving...' : 'Archive'}
            </button>
          </>
        }
      />
    )
  }

  async function handleSave() {
    setSaving(true)
    setError(null)

    try {
      await onSave({
        text: draft.text.trim(),
        claim_type: draft.claim_type,
        confidence: draft.confidence === '' ? null : Number(draft.confidence),
        status: draft.status,
        origin_engagement_id: claim.origin_engagement_id ?? '',
        notes: draft.notes.trim(),
      })
      setEditing(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update claim')
    } finally {
      setSaving(false)
    }
  }

  async function handleArchive() {
    const confirmed = window.confirm('Archive this claim?')
    if (!confirmed) {
      return
    }

    setArchivePending(true)
    setError(null)

    try {
      await onArchive()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to archive claim')
      setArchivePending(false)
    }
  }

  return (
    <article className="rounded-[1.75rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur">
      <p className="text-xs uppercase tracking-[0.2em] text-accent/80">Edit claim</p>
      <textarea
        value={draft.text}
        onChange={(event) => setDraft((current) => ({ ...current, text: event.target.value }))}
        rows={4}
        className="mt-4 w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
      />

      <div className="mt-4 grid gap-4 md:grid-cols-3">
        <label className="block">
          <span className="mb-2 block text-sm text-ink/75">Claim type</span>
          <select
            value={draft.claim_type}
            onChange={(event) => setDraft((current) => ({ ...current, claim_type: event.target.value as Claim['claim_type'] }))}
            className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
          >
            {['OBSERVATION', 'INTERPRETATION', 'PERSONAL_VIEW', 'QUESTION', 'HYPOTHESIS'].map((claimType) => (
              <option key={claimType} value={claimType}>
                {claimType.toLowerCase().replace(/_/g, ' ')}
              </option>
            ))}
          </select>
        </label>

        <label className="block">
          <span className="mb-2 block text-sm text-ink/75">Status</span>
          <select
            value={draft.status}
            onChange={(event) => setDraft((current) => ({ ...current, status: event.target.value as Claim['status'] }))}
            className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
          >
            {['ACTIVE', 'TENTATIVE', 'REVISED', 'ABANDONED'].map((status) => (
              <option key={status} value={status}>
                {status.toLowerCase().replace(/_/g, ' ')}
              </option>
            ))}
          </select>
        </label>

        <label className="block">
          <span className="mb-2 block text-sm text-ink/75">Confidence</span>
          <select
            value={draft.confidence}
            onChange={(event) => setDraft((current) => ({ ...current, confidence: event.target.value }))}
            className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
          >
            <option value="">Not set</option>
            {[1, 2, 3, 4, 5].map((confidence) => (
              <option key={confidence} value={confidence}>
                {confidence}
              </option>
            ))}
          </select>
        </label>
      </div>

      <label className="mt-4 block">
        <span className="mb-2 block text-sm text-ink/75">Notes</span>
        <textarea
          value={draft.notes}
          onChange={(event) => setDraft((current) => ({ ...current, notes: event.target.value }))}
          rows={3}
          className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
        />
      </label>

      {claim.origin ? (
        <p className="mt-4 text-sm leading-6 text-ink/72">
          Origin: <Link to={`/engagements/${claim.origin.engagement_id}`} className="text-pine underline-offset-2 hover:underline">
            {claim.origin.source_title}
          </Link>
        </p>
      ) : null}

      {error ? (
        <p className="mt-4 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">{error}</p>
      ) : null}

      <div className="mt-4 flex justify-end gap-3">
        <button
          type="button"
          onClick={() => {
            setDraft(toClaimDraft(claim))
            setEditing(false)
            setError(null)
          }}
          className="rounded-2xl border border-black/10 bg-white/90 px-4 py-3 text-sm text-ink transition hover:bg-white"
        >
          Cancel
        </button>
        <button
          type="button"
          disabled={saving}
          onClick={() => void handleSave()}
          className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90 disabled:cursor-wait disabled:opacity-70"
        >
          {saving ? 'Saving...' : 'Save claim'}
        </button>
      </div>
    </article>
  )
}

function toClaimDraft(claim: Claim) {
  return {
    text: claim.text,
    claim_type: claim.claim_type,
    status: claim.status,
    confidence: claim.confidence ? String(claim.confidence) : '',
    notes: claim.notes ?? '',
  }
}
