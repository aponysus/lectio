import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import {
  createEngagement,
  createInquiry,
  getEngagement,
  listEngagementInquiries,
  listInquiries,
  listSources,
  type Engagement,
  type Inquiry,
  type InquiryInput,
  type Source,
  replaceEngagementInquiries,
  updateEngagement,
} from '../../api/client'
import { EngagementForm, type EngagementFormSubmission } from '../../components/engagements/EngagementForm'
import { EmptyState } from '../../components/shared/EmptyState'
import { PageHeader } from '../../components/shared/PageHeader'

type EngagementFormPageProps = {
  mode: 'create' | 'edit'
}

export function EngagementFormPage({ mode }: EngagementFormPageProps) {
  const navigate = useNavigate()
  const { engagementId } = useParams()
  const [searchParams] = useSearchParams()
  const defaultSourceID = searchParams.get('sourceId') ?? undefined

  const [sources, setSources] = useState<Source[]>([])
  const [inquiries, setInquiries] = useState<Inquiry[]>([])
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
        const [nextSources, nextInquiries, nextEngagement, nextLinkedInquiries] = await Promise.all([
          listSources({ limit: 100, sort: 'title' }),
          listInquiries({ limit: 100 }),
          mode === 'edit' && engagementId ? getEngagement(engagementId) : Promise.resolve(null),
          mode === 'edit' && engagementId ? listEngagementInquiries(engagementId) : Promise.resolve([]),
        ])
        if (!cancelled) {
          setSources(nextSources)
          setInquiries(nextInquiries)
          setEngagement(nextEngagement)
          setLinkedInquiryIDs(nextLinkedInquiries.map((inquiry) => inquiry.id))
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
  }, [engagementId, mode])

  const handleCreateInquiry = async (input: InquiryInput) => {
    const saved = await createInquiry(input)
    setInquiries((current) => {
      const existing = current.some((inquiry) => inquiry.id === saved.id)
      if (existing) {
        return current
      }

      return [saved, ...current]
    })
    return saved
  }

  const handleSubmit = async (submission: EngagementFormSubmission) => {
    setSaving(true)
    setError(null)

    try {
      const saved =
        mode === 'create'
          ? await createEngagement(submission.engagement)
          : await updateEngagement(engagementId ?? '', submission.engagement)

      await replaceEngagementInquiries(saved.id, submission.inquiry_ids)

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
    return (
      <section className="rounded-[2rem] border border-black/5 bg-white/70 px-6 py-8 shadow-card backdrop-blur">
        Loading engagement form...
      </section>
    )
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
