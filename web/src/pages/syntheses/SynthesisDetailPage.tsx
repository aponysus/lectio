import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { archiveSynthesis, getSynthesis, type Synthesis } from '../../api/client'
import { EmptyState } from '../../components/shared/EmptyState'
import { PageHeader } from '../../components/shared/PageHeader'

export function SynthesisDetailPage() {
  const navigate = useNavigate()
  const { synthesisId } = useParams()
  const [synthesis, setSynthesis] = useState<Synthesis | null>(null)
  const [loading, setLoading] = useState(true)
  const [archivePending, setArchivePending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!synthesisId) {
      setError('Synthesis id is missing')
      setLoading(false)
      return
    }

    let cancelled = false
    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        const nextSynthesis = await getSynthesis(synthesisId)
        if (!cancelled) {
          setSynthesis(nextSynthesis)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load synthesis')
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
  }, [synthesisId])

  const handleArchive = async () => {
    if (!synthesis || archivePending) {
      return
    }

    const confirmed = window.confirm(`Archive "${synthesis.title}"?`)
    if (!confirmed) {
      return
    }

    setArchivePending(true)
    setError(null)

    try {
      await archiveSynthesis(synthesis.id)
      navigate('/syntheses')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to archive synthesis')
      setArchivePending(false)
    }
  }

  if (loading) {
    return (
      <section className="rounded-[2rem] border border-black/5 bg-white/70 px-6 py-8 shadow-card backdrop-blur">
        Loading synthesis...
      </section>
    )
  }

  if (!synthesis) {
    return (
      <EmptyState
        title="Synthesis not found"
        body={error ?? 'This synthesis could not be loaded.'}
        action={
          <Link
            to="/syntheses"
            className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
          >
            Back to syntheses
          </Link>
        }
      />
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow={synthesis.type.toLowerCase().replace(/_/g, ' ')}
        title={synthesis.title}
        description={synthesis.inquiry ? synthesis.inquiry.question : 'Inquiry summary unavailable.'}
        actions={
          <>
            {synthesis.inquiry ? (
              <Link
                to={`/inquiries/${synthesis.inquiry.id}`}
                className="rounded-2xl border border-black/10 bg-white/70 px-4 py-3 text-sm text-ink transition hover:bg-white"
              >
                Open inquiry
              </Link>
            ) : null}
            <Link
              to={`/syntheses/${synthesis.id}/edit`}
              className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
            >
              Edit synthesis
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
          <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Synthesis body</p>
          <p className="mt-4 whitespace-pre-wrap text-sm leading-7 text-ink/82">{synthesis.body}</p>
        </article>

        <article className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
          <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Metadata</p>
          <dl className="mt-5 space-y-4">
            <MetaItem label="Type" value={synthesis.type.toLowerCase().replace(/_/g, ' ')} />
            <MetaItem label="Created" value={formatDateTime(synthesis.created_at)} />
            <MetaItem label="Updated" value={formatDateTime(synthesis.updated_at)} />
            <MetaItem
              label="Linked inquiry"
              value={synthesis.inquiry ? synthesis.inquiry.title : 'Unavailable'}
            />
          </dl>

          {synthesis.notes ? (
            <div className="mt-6 rounded-2xl bg-black/[0.03] px-4 py-4">
              <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Notes</p>
              <p className="mt-3 whitespace-pre-wrap text-sm leading-6 text-ink/80">{synthesis.notes}</p>
            </div>
          ) : null}
        </article>
      </section>
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
