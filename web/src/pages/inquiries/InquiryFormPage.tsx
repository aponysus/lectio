import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { createInquiry, getInquiry, type Inquiry, type InquiryInput, updateInquiry } from '../../api/client'
import { useToast } from '../../components/feedback/ToastProvider'
import { InquiryForm } from '../../components/inquiries/InquiryForm'
import { LoadingPanel } from '../../components/shared/LoadingPanel'
import { PageHeader } from '../../components/shared/PageHeader'

type InquiryFormPageProps = {
  mode: 'create' | 'edit'
}

export function InquiryFormPage({ mode }: InquiryFormPageProps) {
  const navigate = useNavigate()
  const { showToast } = useToast()
  const { inquiryId } = useParams()
  const [inquiry, setInquiry] = useState<Inquiry | null>(null)
  const [loading, setLoading] = useState(mode === 'edit')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (mode !== 'edit' || !inquiryId) {
      return
    }

    let cancelled = false
    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        const nextInquiry = await getInquiry(inquiryId)
        if (!cancelled) {
          setInquiry(nextInquiry)
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
  }, [inquiryId, mode])

  const handleSubmit = async (input: InquiryInput) => {
    setSaving(true)
    setError(null)

    try {
      const saved =
        mode === 'create'
          ? await createInquiry(input)
          : await updateInquiry(inquiryId ?? '', input)

      showToast({ message: mode === 'create' ? 'Inquiry created.' : 'Inquiry updated.' })
      navigate(`/inquiries/${saved.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save inquiry')
    } finally {
      setSaving(false)
    }
  }

  const actions =
    mode === 'edit' && inquiryId ? (
      <Link
        to={`/inquiries/${inquiryId}`}
        className="rounded-2xl border border-black/10 bg-white/70 px-4 py-3 text-sm text-ink transition hover:bg-white"
      >
        Back to inquiry
      </Link>
    ) : (
      <Link
        to="/inquiries"
        className="rounded-2xl border border-black/10 bg-white/70 px-4 py-3 text-sm text-ink transition hover:bg-white"
      >
        Back to inquiries
      </Link>
    )

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow={mode === 'create' ? 'New inquiry' : 'Edit inquiry'}
        title={mode === 'create' ? 'Create a live question workspace' : 'Refine the inquiry workspace'}
        description={
          mode === 'create'
            ? 'This should stay focused: name the question, set its status, and add only enough context to guide future engagement.'
            : 'Update the inquiry while it is still a live workspace, not a frozen archive entry.'
        }
        actions={actions}
      />

      {loading ? (
        <LoadingPanel label="Loading inquiry form" />
      ) : (
        <InquiryForm
          inquiry={inquiry}
          submitLabel={mode === 'create' ? 'Create inquiry' : 'Save changes'}
          submitting={saving}
          apiError={error}
          onSubmit={handleSubmit}
        />
      )}
    </div>
  )
}
