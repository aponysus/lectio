import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { createSource, getSource, type Source, type SourceInput, updateSource } from '../../api/client'
import { useToast } from '../../components/feedback/ToastProvider'
import { LoadingPanel } from '../../components/shared/LoadingPanel'
import { SourceForm } from '../../components/sources/SourceForm'
import { PageHeader } from '../../components/shared/PageHeader'

type SourceFormPageProps = {
  mode: 'create' | 'edit'
}

export function SourceFormPage({ mode }: SourceFormPageProps) {
  const navigate = useNavigate()
  const { showToast } = useToast()
  const { sourceId } = useParams()
  const [source, setSource] = useState<Source | null>(null)
  const [loading, setLoading] = useState(mode === 'edit')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (mode !== 'edit' || !sourceId) {
      return
    }

    let cancelled = false
    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        const nextSource = await getSource(sourceId)
        if (!cancelled) {
          setSource(nextSource)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load source')
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
  }, [mode, sourceId])

  const handleSubmit = async (input: SourceInput) => {
    setSaving(true)
    setError(null)

    try {
      const saved =
        mode === 'create'
          ? await createSource(input)
          : await updateSource(sourceId ?? '', input)

      showToast({ message: mode === 'create' ? 'Source created.' : 'Source updated.' })
      navigate(`/sources/${saved.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save source')
    } finally {
      setSaving(false)
    }
  }

  const actions =
    mode === 'edit' && sourceId ? (
      <Link
        to={`/sources/${sourceId}`}
        className="rounded-2xl border border-black/10 bg-white/70 px-4 py-3 text-sm text-ink transition hover:bg-white"
      >
        Back to source
      </Link>
    ) : (
      <Link
        to="/sources"
        className="rounded-2xl border border-black/10 bg-white/70 px-4 py-3 text-sm text-ink transition hover:bg-white"
      >
        Back to sources
      </Link>
    )

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow={mode === 'create' ? 'New source' : 'Edit source'}
        title={mode === 'create' ? 'Create a stable source record' : 'Refine source metadata'}
        description={
          mode === 'create'
            ? 'Keep sources clean and intentional so engagement capture never turns into duplicate cleanup.'
            : 'Update the source metadata now so later inquiry and engagement work rests on the right foundation.'
        }
        actions={actions}
      />

      {loading ? (
        <LoadingPanel label="Loading source form" />
      ) : (
        <SourceForm
          source={source}
          submitLabel={mode === 'create' ? 'Create source' : 'Save changes'}
          submitting={saving}
          apiError={error}
          onSubmit={handleSubmit}
        />
      )}
    </div>
  )
}
