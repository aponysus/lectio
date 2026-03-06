import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import {
  archiveClaim,
  archiveLanguageNote,
  createClaim,
  createEngagement,
  createInquiry,
  createLanguageNote,
  getEngagement,
  listEngagementClaims,
  listEngagementInquiries,
  listEngagementLanguageNotes,
  listInquiries,
  listSources,
  replaceClaimInquiries,
  type Engagement,
  type Claim,
  type EngagementCreateInput,
  type ClaimUpdateInput,
  type Inquiry,
  type InquiryInput,
  type LanguageNote,
  type LanguageNoteUpdateInput,
  type Source,
  replaceEngagementInquiries,
  updateClaim,
  updateEngagement,
  updateLanguageNote,
} from '../../api/client'
import { useToast } from '../../components/feedback/ToastProvider'
import {
  EngagementForm,
  type EngagementFormLanguageNoteSubmission,
  type EngagementFormSubmission,
} from '../../components/engagements/EngagementForm'
import { EmptyState } from '../../components/shared/EmptyState'
import { LoadingPanel } from '../../components/shared/LoadingPanel'
import { PageHeader } from '../../components/shared/PageHeader'

type EngagementFormPageProps = {
  mode: 'create' | 'edit'
}

const draftInquiryPrefix = 'draft-inquiry:'

export function EngagementFormPage({ mode }: EngagementFormPageProps) {
  const navigate = useNavigate()
  const { showToast } = useToast()
  const { engagementId } = useParams()
  const [searchParams] = useSearchParams()
  const defaultSourceID = searchParams.get('sourceId') ?? undefined
  const defaultInquiryID = searchParams.get('inquiryId') ?? undefined

  const [sources, setSources] = useState<Source[]>([])
  const [claims, setClaims] = useState<Claim[]>([])
  const [languageNotes, setLanguageNotes] = useState<LanguageNote[]>([])
  const [inquiries, setInquiries] = useState<Inquiry[]>([])
  const [draftInquiries, setDraftInquiries] = useState<Record<string, InquiryInput>>({})
  const [linkedInquiryIDs, setLinkedInquiryIDs] = useState<string[]>([])
  const [engagement, setEngagement] = useState<Engagement | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        const [nextSources, nextInquiries, nextEngagement, nextLinkedInquiries, nextClaims, nextLanguageNotes] = await Promise.all([
          listSources({ limit: 100, sort: 'title' }),
          listInquiries({ limit: 100 }),
          mode === 'edit' && engagementId ? getEngagement(engagementId) : Promise.resolve(null),
          mode === 'edit' && engagementId ? listEngagementInquiries(engagementId) : Promise.resolve([]),
          mode === 'edit' && engagementId ? listEngagementClaims(engagementId) : Promise.resolve([]),
          mode === 'edit' && engagementId ? listEngagementLanguageNotes(engagementId) : Promise.resolve([]),
        ])
        if (!cancelled) {
          setSources(nextSources)
          setInquiries(nextInquiries)
          setDraftInquiries({})
          setEngagement(nextEngagement)
          setLinkedInquiryIDs(
            mode === 'edit'
              ? nextLinkedInquiries.map((inquiry) => inquiry.id)
              : defaultInquiryID
                ? [defaultInquiryID]
                : [],
          )
          setClaims(nextClaims)
          setLanguageNotes(nextLanguageNotes)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load engagement form')
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
  }, [defaultInquiryID, engagementId, mode])

  const handleCreateInquiry = async (input: InquiryInput) => {
    if (mode === 'create') {
      const draftID = `${draftInquiryPrefix}${crypto.randomUUID()}`
      const now = new Date().toISOString()
      setDraftInquiries((current) => ({ ...current, [draftID]: input }))
      setInquiries((current) => [
        {
          id: draftID,
          title: input.title,
          question: input.question,
          status: input.status,
          why_it_matters: input.why_it_matters || undefined,
          current_view: input.current_view || undefined,
          open_tensions: input.open_tensions || undefined,
          created_at: now,
          updated_at: now,
          engagement_count: 0,
          claim_count: 0,
          synthesis_count: 0,
        },
        ...current,
      ])
      showToast({ message: 'Inquiry queued and will be created with the engagement.' })
      return {
        id: draftID,
        title: input.title,
        question: input.question,
        status: input.status,
        why_it_matters: input.why_it_matters || undefined,
        current_view: input.current_view || undefined,
        open_tensions: input.open_tensions || undefined,
        created_at: now,
        updated_at: now,
        engagement_count: 0,
        claim_count: 0,
        synthesis_count: 0,
      }
    }

    const saved = await createInquiry(input)
    setInquiries((current) => {
      const existing = current.some((inquiry) => inquiry.id === saved.id)
      if (existing) {
        return current
      }

      return [saved, ...current]
    })
    showToast({ message: 'Inquiry created and ready to link.' })
    return saved
  }

  const handleSubmit = async (submission: EngagementFormSubmission) => {
    setSaving(true)
    setError(null)

    try {
      if (mode === 'create') {
        const inlineInquiries = submission.inquiry_ids
          .filter((inquiryID) => isDraftInquiryID(inquiryID))
          .map((inquiryID) => draftInquiries[inquiryID])
          .filter((inquiry): inquiry is InquiryInput => Boolean(inquiry))
        const inquiryIDs = submission.inquiry_ids.filter((inquiryID) => !isDraftInquiryID(inquiryID))

        const saved = await createEngagement(toEngagementCreateInput(submission, inquiryIDs, inlineInquiries))
        showToast({ message: 'Engagement saved.' })
        navigate(`/engagements/${saved.id}`)
        return
      }

      const saved = await updateEngagement(engagementId ?? '', submission.engagement)

      await replaceEngagementInquiries(saved.id, submission.inquiry_ids)

      const retainedClaimIDs = new Set<string>()
      for (const claim of submission.claims) {
        if (claim.claim_id) {
          const updated = await updateClaim(claim.claim_id, toClaimUpdateInput(claim, saved.id))
          await replaceClaimInquiries(updated.id, submission.inquiry_ids)
          retainedClaimIDs.add(updated.id)
          continue
        }

        const created = await createClaim(toClaimCreateInput(claim, saved.id, submission.inquiry_ids))
        retainedClaimIDs.add(created.id)
      }

      for (const existingClaim of claims) {
        if (!retainedClaimIDs.has(existingClaim.id)) {
          await archiveClaim(existingClaim.id)
        }
      }

      const retainedLanguageNoteIDs = new Set<string>()
      for (const note of submission.language_notes) {
        if (note.note_id) {
          const updated = await updateLanguageNote(note.note_id, toLanguageNoteUpdateInput(note))
          retainedLanguageNoteIDs.add(updated.id)
          continue
        }

        const created = await createLanguageNote(toLanguageNoteCreateInput(note, saved.id))
        retainedLanguageNoteIDs.add(created.id)
      }

      for (const existingNote of languageNotes) {
        if (!retainedLanguageNoteIDs.has(existingNote.id)) {
          await archiveLanguageNote(existingNote.id)
        }
      }

      showToast({ message: 'Engagement updated.' })
      navigate(`/engagements/${saved.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save engagement')
    } finally {
      setSaving(false)
    }
  }

  const actions =
    mode === 'edit' && engagementId ? (
      <Link
        to={`/engagements/${engagementId}`}
        className="rounded-2xl border border-black/10 bg-white/70 px-4 py-3 text-sm text-ink transition hover:bg-white"
      >
        Back to engagement
      </Link>
    ) : (
      <Link
        to={defaultSourceID ? `/sources/${defaultSourceID}` : '/sources'}
        className="rounded-2xl border border-black/10 bg-white/70 px-4 py-3 text-sm text-ink transition hover:bg-white"
      >
        {defaultSourceID ? 'Back to source' : 'Back to sources'}
      </Link>
    )

  if (loading) {
    return <LoadingPanel label="Loading engagement form" />
  }

  if (sources.length === 0) {
    return (
      <EmptyState
        title="Create a source first"
        body="Engagement capture depends on a stable source record. Add a source, then come back to log the engagement."
        action={
          <Link
            to="/sources/new"
            className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
          >
            Create source
          </Link>
        }
      />
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow={mode === 'create' ? 'New engagement' : 'Edit engagement'}
        title={mode === 'create' ? 'Log a meaningful encounter with a source' : 'Refine the engagement record'}
        description={
          mode === 'create'
            ? 'This flow should stay fast: source, timestamp, reflection, and just enough context to make the note useful later.'
            : 'Tighten the record without turning the capture flow into bureaucracy.'
        }
        actions={actions}
      />

      {error ? (
        <section className="rounded-[2rem] border border-amber-200 bg-amber-50 px-6 py-5 text-amber-700 shadow-card">
          {error}
        </section>
      ) : null}

      <EngagementForm
        engagement={engagement}
        claims={claims}
        languageNotes={languageNotes}
        sources={sources}
        inquiries={inquiries}
        linkedInquiryIDs={linkedInquiryIDs}
        defaultSourceID={defaultSourceID}
        submitLabel={mode === 'create' ? 'Create engagement' : 'Save changes'}
        submitting={saving}
        apiError={error}
        onCreateInquiry={handleCreateInquiry}
        onSubmit={handleSubmit}
      />
    </div>
  )
}

function toClaimUpdateInput(
  claim: EngagementFormSubmission['claims'][number],
  engagementID: string,
): ClaimUpdateInput {
  return {
    text: claim.text,
    claim_type: claim.claim_type,
    confidence: claim.confidence,
    status: claim.status,
    origin_engagement_id: engagementID,
    notes: claim.notes,
  }
}

function toClaimCreateInput(
  claim: EngagementFormSubmission['claims'][number],
  engagementID: string,
  inquiryIDs: string[],
) {
  return {
    text: claim.text,
    claim_type: claim.claim_type,
    confidence: claim.confidence,
    status: claim.status,
    origin_engagement_id: engagementID,
    notes: claim.notes,
    inquiry_ids: inquiryIDs,
  }
}

function toLanguageNoteUpdateInput(note: EngagementFormLanguageNoteSubmission): LanguageNoteUpdateInput {
  return {
    term: note.term,
    language: note.language,
    note_type: note.note_type,
    content: note.content,
  }
}

function toLanguageNoteCreateInput(note: EngagementFormLanguageNoteSubmission, engagementID: string) {
  return {
    engagement_id: engagementID,
    term: note.term,
    language: note.language,
    note_type: note.note_type,
    content: note.content,
  }
}

function toEngagementCreateInput(
  submission: EngagementFormSubmission,
  inquiryIDs: string[],
  inlineInquiries: InquiryInput[],
): EngagementCreateInput {
  return {
    ...submission.engagement,
    inquiry_ids: inquiryIDs,
    inline_inquiries: inlineInquiries,
    claims: submission.claims.map((claim) => ({
      text: claim.text,
      claim_type: claim.claim_type,
      confidence: claim.confidence,
      status: claim.status,
      notes: claim.notes,
    })),
    language_notes: submission.language_notes.map((note) => ({
      term: note.term,
      language: note.language,
      note_type: note.note_type,
      content: note.content,
    })),
  }
}

function isDraftInquiryID(inquiryID: string): boolean {
  return inquiryID.startsWith(draftInquiryPrefix)
}
