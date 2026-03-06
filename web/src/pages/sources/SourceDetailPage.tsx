import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { archiveSource, getSource, listEngagements, type Engagement, type Source } from '../../api/client'
import { EngagementCard } from '../../components/engagements/EngagementCard'
import { useConfirm } from '../../components/feedback/ConfirmProvider'
import { useToast } from '../../components/feedback/ToastProvider'
import { EmptyState } from '../../components/shared/EmptyState'
import { LoadingPanel } from '../../components/shared/LoadingPanel'
import { PageHeader } from '../../components/shared/PageHeader'

export function SourceDetailPage() {
  const navigate = useNavigate()
  const { confirm } = useConfirm()
  const { showToast } = useToast()
  const { sourceId } = useParams()
  const [source, setSource] = useState<Source | null>(null)
  const [engagements, setEngagements] = useState<Engagement[]>([])
  const [loading, setLoading] = useState(true)
  const [archivePending, setArchivePending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!sourceId) {
      setError('Source id is missing')
      setLoading(false)
      return
    }

    let cancelled = false
    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        const [nextSource, nextEngagements] = await Promise.all([
          getSource(sourceId),
          listEngagements({ source_id: sourceId, limit: 20 }),
        ])
        if (!cancelled) {
          setSource(nextSource)
          setEngagements(nextEngagements)
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
  }, [sourceId])

  const handleArchive = async () => {
    if (!source || archivePending) {
      return
    }

    const confirmed = await confirm({
      title: 'Archive source?',
      body: `Archive "${source.title}"? Its engagements will remain, but this source will disappear from active lists.`,
      confirmLabel: 'Archive source',
    })
    if (!confirmed) {
      return
    }

    setArchivePending(true)
    setError(null)
    try {
      await archiveSource(source.id)
      showToast({ message: 'Source archived.', tone: 'info' })
      navigate('/sources')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to archive source')
      setArchivePending(false)
    }
  }

  if (loading) {
    return <LoadingPanel label="Loading source" />
  }

  if (!source) {
    return (
      <EmptyState
        title="Source not found"
        body={error ?? 'This source could not be loaded.'}
        action={
          <Link
            to="/sources"
            className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
          >
            Back to sources
          </Link>
        }
      />
    )
  }

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow={source.medium.toLowerCase().replace(/_/g, ' ')}
        title={source.title}
        description={
          source.creator
            ? `${source.creator}${source.year ? ` • ${source.year}` : ''}`
            : source.year
              ? String(source.year)
              : 'Creator and year not set yet.'
        }
        actions={
          <>
            <Link
              to={`/engagements/new?sourceId=${source.id}`}
              className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
            >
              Log engagement
            </Link>
            <Link
              to={`/sources/${source.id}/edit`}
              className="rounded-2xl border border-black/10 bg-white/75 px-4 py-3 text-sm text-ink transition hover:bg-white"
            >
              Edit source
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

      <section className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
        <article className="rounded-[2rem] border border-black/5 bg-white/75 p-6 shadow-card backdrop-blur">
          <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Metadata</p>
          <dl className="mt-5 grid gap-4 sm:grid-cols-2">
            <MetaItem label="Medium" value={source.medium.toLowerCase().replace(/_/g, ' ')} />
            <MetaItem label="Original language" value={source.original_language ?? 'Not set'} />
            <MetaItem label="Culture / context" value={source.culture_or_context ?? 'Not set'} />
            <MetaItem label="Updated" value={formatDate(source.updated_at)} />
          </dl>
          {source.notes ? (
            <div className="mt-6 rounded-2xl bg-black/[0.03] px-4 py-4">
              <p className="text-xs uppercase tracking-[0.2em] text-accent/75">Notes</p>
              <p className="mt-3 whitespace-pre-wrap text-sm leading-6 text-ink/78">{source.notes}</p>
            </div>
          ) : null}
        </article>

        {engagements.length === 0 ? (
          <EmptyState
            title="No engagements yet"
            body="This source record is ready. Use it as the entrypoint for the engagement flow whenever you next read, watch, or listen with intent."
            action={
              <Link
                to={`/engagements/new?sourceId=${source.id}`}
                className="rounded-2xl bg-accent px-4 py-3 text-sm font-medium text-white transition hover:bg-accent/90"
              >
                Open engagement flow
              </Link>
            }
          />
        ) : (
          <section className="space-y-4">
            <div>
              <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Engagements</p>
              <h3 className="mt-2 font-display text-3xl text-ink">Captured encounters with this source</h3>
            </div>
            <div className="space-y-4">
              {engagements.map((engagement) => (
                <EngagementCard key={engagement.id} engagement={engagement} showSource={false} />
              ))}
            </div>
          </section>
        )}
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

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}
