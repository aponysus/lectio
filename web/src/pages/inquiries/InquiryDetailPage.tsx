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
import { useConfirm } from '../../components/feedback/ConfirmProvider'
import { useToast } from '../../components/feedback/ToastProvider'
import { EngagementCard } from '../../components/engagements/EngagementCard'
import { SynthesisCard } from '../../components/syntheses/SynthesisCard'
import { EmptyState } from '../../components/shared/EmptyState'
import { useWorkspaceSidebar } from '../../components/shared/AppShell'
import { formFieldClassName } from '../../components/shared/formStyles'
import { LoadingPanel } from '../../components/shared/LoadingPanel'
import { PageHeader } from '../../components/shared/PageHeader'

export function InquiryDetailPage() {
  const navigate = useNavigate()
  const { confirm } = useConfirm()
  const { showToast } = useToast()
  const setWorkspaceSidebar = useWorkspaceSidebar()
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

  useEffect(() => {
    if (!inquiry) {
      setWorkspaceSidebar(null)
      return
    }

    const latestActivity = formatDateTime(inquiry.latest_activity ?? inquiry.updated_at)
    const synthesisReady = inquiry.engagement_count >= 3 || inquiry.claim_count >= 2

    setWorkspaceSidebar({
      eyebrow: 'Active inquiry',
      title: inquiry.title,
      body: `Use the workspace in order: question, evidence, synthesis, then the next move. Latest activity ${latestActivity}.`,
      stats: [
        { label: 'Status', value: inquiry.status.toLowerCase().replace(/_/g, ' ') },
        { label: 'Engagements', value: String(inquiry.engagement_count) },
        { label: 'Claims', value: String(inquiry.claim_count) },
        { label: 'Syntheses', value: String(inquiry.synthesis_count) },
      ],
      links: [
        { label: 'Question and posture', href: '#workspace-compass', detail: 'orient' },
        { label: 'Evidence feed', href: '#evidence-feed', detail: `${engagements.length + claims.length} items` },
        { label: 'Synthesis rail', href: '#synthesis-rail', detail: synthesisReady ? 'ready' : 'collecting' },
        { label: 'Next moves', href: '#next-moves', detail: 'prompt' },
      ],
      actions: [
        { label: 'Log engagement', href: `/engagements/new?inquiryId=${inquiry.id}`, tone: 'primary' },
        { label: 'Write synthesis', href: `/syntheses/new?inquiryId=${inquiry.id}` },
      ],
    })

    return () => {
      setWorkspaceSidebar(null)
    }
  }, [
    claims.length,
    engagements.length,
    inquiry,
    setWorkspaceSidebar,
  ])

  const handleArchive = async () => {
    if (!inquiry || archivePending) {
      return
    }

    const confirmed = await confirm({
      title: 'Archive inquiry?',
      body: `Archive "${inquiry.title}"? Its linked claims, engagements, and syntheses will remain accessible through their own records.`,
      confirmLabel: 'Archive inquiry',
    })
    if (!confirmed) {
      return
    }

    setArchivePending(true)
    setError(null)

    try {
      await archiveInquiry(inquiry.id)
      showToast({ message: 'Inquiry archived.', tone: 'info' })
      navigate('/inquiries')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to archive inquiry')
      setArchivePending(false)
    }
  }

  if (loading) {
    return <LoadingPanel label="Loading inquiry" />
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
  const workspaceSummary = `${inquiry.engagement_count} engagements • ${inquiry.claim_count} claims • ${inquiry.synthesis_count} syntheses • Latest activity ${formatDateTime(
    inquiry.latest_activity ?? inquiry.updated_at,
  )}`
  const evidenceFeed = [
    ...engagements.map((engagement) => ({
      kind: 'engagement' as const,
      id: engagement.id,
      sortAt: engagement.engaged_at,
      engagement,
    })),
    ...claims.map((claim) => ({
      kind: 'claim' as const,
      id: claim.id,
      sortAt: claim.updated_at,
      claim,
    })),
  ].sort((left, right) => toTimestamp(right.sortAt) - toTimestamp(left.sortAt))
  const nextMoves = buildNextMoves({
    engagementCount: engagements.length,
    claimCount: claims.length,
    synthesisCount: syntheses.length,
    synthesisReady,
  })

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow={inquiry.status.toLowerCase().replace(/_/g, ' ')}
        title={inquiry.title}
        description={workspaceSummary}
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

      <section className="grid gap-6 xl:grid-cols-[0.92fr_1.24fr_0.84fr]">
        <aside className="space-y-5 xl:sticky xl:top-6 xl:self-start">
          <article
            id="workspace-compass"
            className="scroll-mt-6 rounded-[1.5rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur"
          >
            <p className="text-xs uppercase tracking-[0.24em] text-accent/80">Workspace compass</p>
            <h3 className="mt-2 font-display text-[1.75rem] leading-tight text-ink">The live question</h3>
            <p className="mt-4 whitespace-pre-wrap text-sm leading-7 text-ink/82">{inquiry.question}</p>

            <div className="mt-5 space-y-3">
              <ContextBlock label="Why it matters" value={inquiry.why_it_matters} empty="Not set yet." />
              <ContextBlock label="Current view" value={inquiry.current_view} empty="Still taking shape." />
              <ContextBlock label="Open tensions" value={inquiry.open_tensions} empty="No tensions recorded yet." />
            </div>
          </article>

          <article className="rounded-[1.5rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur">
            <p className="text-xs uppercase tracking-[0.24em] text-accent/80">Current posture</p>
            <div className="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-1">
              <MetaItem label="Status" value={inquiry.status.toLowerCase().replace(/_/g, ' ')} />
              <MetaItem label="Engagements" value={String(inquiry.engagement_count)} />
              <MetaItem label="Claims" value={String(inquiry.claim_count)} />
              <MetaItem label="Latest activity" value={formatDateTime(inquiry.latest_activity ?? inquiry.updated_at)} />
            </div>

            <div className="mt-5 flex flex-col gap-3">
              <Link
                to={`/engagements/new?inquiryId=${inquiry.id}`}
                className="rounded-2xl bg-accent px-4 py-3 text-sm font-medium text-white transition hover:bg-accent/90"
              >
                Log engagement into this inquiry
              </Link>
              <Link
                to={`/syntheses/new?inquiryId=${inquiry.id}`}
                className="rounded-2xl border border-black/10 bg-white/90 px-4 py-3 text-sm text-ink transition hover:bg-white"
              >
                Write synthesis
              </Link>
            </div>
          </article>
        </aside>

        <section className="space-y-5">
          <article
            id="evidence-feed"
            className="scroll-mt-6 rounded-[1.5rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur"
          >
            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
              <div className="max-w-3xl">
                <p className="text-xs uppercase tracking-[0.24em] text-accent/80">Evidence feed</p>
                <h3 className="mt-2 font-display text-[1.75rem] leading-tight text-ink">Claims and source work in one stream</h3>
                <p className="mt-3 text-sm leading-6 text-ink/74">
                  The center column stays chronological: recent engagements and claim revisions are kept together so the question, the evidence, and the revisions remain legible.
                </p>
              </div>
              <div className="rounded-[1.25rem] bg-black/[0.03] px-4 py-3 text-sm leading-6 text-ink/72">
                {engagements.length} engagements • {claims.length} claims
              </div>
            </div>

            {evidenceFeed.length === 0 ? (
              <div className="mt-5 rounded-[1.5rem] border border-dashed border-black/10 bg-white/55 px-5 py-8 text-center">
                <h4 className="font-display text-[1.6rem] text-ink">No evidence linked yet</h4>
                <p className="mx-auto mt-3 max-w-2xl text-sm leading-7 text-ink/72">
                  Attach this inquiry during engagement capture to turn it into a real workspace with evidence, tensions, and later claims.
                </p>
                <Link
                  to={`/engagements/new?inquiryId=${inquiry.id}`}
                  className="mt-5 inline-flex rounded-2xl bg-accent px-4 py-3 text-sm font-medium text-white transition hover:bg-accent/90"
                >
                  Log first engagement
                </Link>
              </div>
            ) : (
              <div className="mt-5 space-y-5">
                {evidenceFeed.map((item) => (
                  <div key={`${item.kind}-${item.id}`} className="space-y-3">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <p className="text-xs uppercase tracking-[0.2em] text-accent/75">
                        {item.kind === 'engagement' ? 'Engagement' : 'Claim revision'}
                      </p>
                      <p className="text-sm text-ink/64">
                        {item.kind === 'engagement'
                          ? formatDateTime(item.engagement.engaged_at)
                          : `Updated ${formatDateTime(item.claim.updated_at)}`}
                      </p>
                    </div>

                    {item.kind === 'engagement' ? (
                      <EngagementCard engagement={item.engagement} />
                    ) : (
                      <EditableClaimCard
                        claim={item.claim}
                        onSave={async (nextInput) => {
                          const updated = await updateClaim(item.claim.id, nextInput)
                          setClaims((current) => current.map((currentClaim) => (currentClaim.id === updated.id ? updated : currentClaim)))
                        }}
                        onArchive={async () => {
                          await archiveClaim(item.claim.id)
                          setClaims((current) => current.filter((currentClaim) => currentClaim.id !== item.claim.id))
                          setInquiry((current) =>
                            current
                              ? {
                                  ...current,
                                  claim_count: Math.max(0, current.claim_count - 1),
                                  latest_activity: new Date().toISOString(),
                                }
                              : current,
                          )
                        }}
                      />
                    )}
                  </div>
                ))}
              </div>
            )}
          </article>
        </section>

        <aside className="space-y-5 xl:sticky xl:top-6 xl:self-start">
          <article
            id="synthesis-rail"
            className="scroll-mt-6 rounded-[1.5rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur"
          >
            <div className="flex flex-col gap-4">
              <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div>
                  <p className="text-xs uppercase tracking-[0.24em] text-accent/80">Synthesis rail</p>
                  <h3 className="mt-2 font-display text-[1.75rem] leading-tight text-ink">Compression attempts and next moves</h3>
                </div>
                <Link
                  to={`/syntheses/new?inquiryId=${inquiry.id}`}
                  className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
                >
                  New synthesis
                </Link>
              </div>

              <div
                className={`rounded-[1.25rem] border px-4 py-4 ${
                  synthesisReady
                    ? 'border-pine/20 bg-pine/10'
                    : 'border-black/6 bg-black/[0.03]'
                }`}
              >
                <p className="text-xs uppercase tracking-[0.2em] text-accent/75">
                  {synthesisReady ? 'Ready for compression' : 'Still collecting'}
                </p>
                <p className="mt-3 text-sm leading-6 text-ink/78">
                  {synthesisReady
                    ? 'This inquiry has enough density to justify a synthesis pass now.'
                    : 'Keep gathering evidence. Once the inquiry reaches three engagements or two claims, synthesis becomes the right move.'}
                </p>
              </div>

              {latestSynthesis ? (
                <div className="space-y-3">
                  <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Latest synthesis</p>
                  <SynthesisCard synthesis={latestSynthesis} />
                </div>
              ) : (
                <div className="rounded-[1.25rem] border border-dashed border-black/10 bg-white/55 px-4 py-5">
                  <p className="text-xs uppercase tracking-[0.2em] text-accent/75">No synthesis yet</p>
                  <p className="mt-3 text-sm leading-6 text-ink/76">
                    The inquiry does not have a stored compression attempt yet. Use the button above when you can state the current view and the unresolved tension in one pass.
                  </p>
                </div>
              )}

              {priorSyntheses.length > 0 ? (
                <div className="space-y-3">
                  <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Earlier syntheses</p>
                  <div className="space-y-3">
                    {priorSyntheses.map((synthesis) => (
                      <SynthesisCard key={synthesis.id} synthesis={synthesis} />
                    ))}
                  </div>
                </div>
              ) : null}
            </div>
          </article>

          <article
            id="next-moves"
            className="scroll-mt-6 rounded-[1.5rem] border border-black/5 bg-white/75 p-5 shadow-card backdrop-blur"
          >
            <p className="text-xs uppercase tracking-[0.24em] text-accent/80">Next moves</p>
            <div className="mt-4 space-y-3">
              {nextMoves.map((move) => (
                <div key={move.title} className="rounded-[1.25rem] bg-black/[0.03] px-4 py-4">
                  <p className="text-sm font-medium text-ink">{move.title}</p>
                  <p className="mt-2 text-sm leading-6 text-ink/74">{move.body}</p>
                </div>
              ))}
            </div>
          </article>
        </aside>
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
  const { confirm } = useConfirm()
  const { showToast } = useToast()
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
      showToast({ message: 'Claim updated.' })
      setEditing(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update claim')
    } finally {
      setSaving(false)
    }
  }

  async function handleArchive() {
    const confirmed = await confirm({
      title: 'Archive claim?',
      body: 'Archive this claim? It will be removed from the active inquiry workspace and rediscovery prompts.',
      confirmLabel: 'Archive claim',
    })
    if (!confirmed) {
      return
    }

    setArchivePending(true)
    setError(null)

    try {
      await onArchive()
      showToast({ message: 'Claim archived.', tone: 'info' })
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
        className={`mt-4 ${formFieldClassName}`}
      />

      <div className="mt-4 grid gap-4 md:grid-cols-3">
        <label className="block">
          <span className="mb-2 block text-sm text-ink/75">Claim type</span>
          <select
            value={draft.claim_type}
            onChange={(event) => setDraft((current) => ({ ...current, claim_type: event.target.value as Claim['claim_type'] }))}
            className={formFieldClassName}
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
            className={formFieldClassName}
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
            className={formFieldClassName}
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
          className={formFieldClassName}
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

function toTimestamp(value: string) {
  const timestamp = new Date(value).getTime()
  return Number.isNaN(timestamp) ? 0 : timestamp
}

function buildNextMoves({
  engagementCount,
  claimCount,
  synthesisCount,
  synthesisReady,
}: {
  engagementCount: number
  claimCount: number
  synthesisCount: number
  synthesisReady: boolean
}) {
  const moves: Array<{ title: string; body: string }> = []

  if (engagementCount === 0) {
    moves.push({
      title: 'Collect source evidence',
      body: 'Log the next engagement and attach this inquiry so the question has real material feeding it.',
    })
  }

  if (engagementCount > 0 && claimCount === 0) {
    moves.push({
      title: 'Extract claims',
      body: 'Pull one to three propositions or questions out of the recent engagements so the inquiry starts sharpening.',
    })
  }

  if (synthesisReady && synthesisCount === 0) {
    moves.push({
      title: 'Write the first synthesis',
      body: 'State the current view, identify what remains unresolved, and decide what evidence the inquiry needs next.',
    })
  }

  if (synthesisCount > 0) {
    moves.push({
      title: 'Pressure-test the latest synthesis',
      body: 'Use the next engagement or claim revision to confirm, complicate, or overturn the current compressed position.',
    })
  }

  if (moves.length === 0) {
    moves.push({
      title: 'Keep the workspace live',
      body: 'Revisit the question after the next meaningful encounter and avoid letting the inquiry flatten into archive.',
    })
  }

  return moves
}
