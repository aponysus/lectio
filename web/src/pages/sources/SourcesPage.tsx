import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { type ListSourcesFilters, listSources, SOURCE_MEDIA, type Source } from '../../api/client'
import { SourceCard } from '../../components/sources/SourceCard'
import { EmptyState } from '../../components/shared/EmptyState'
import { PageHeader } from '../../components/shared/PageHeader'

export function SourcesPage() {
  const [filters, setFilters] = useState<ListSourcesFilters>({
    q: '',
    medium: '',
    original_language: '',
    sort: 'recent',
  })
  const [sources, setSources] = useState<Source[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        const nextSources = await listSources(filters)
        if (!cancelled) {
          setSources(nextSources)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load sources')
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
  }, [filters])

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Sources"
        title="Serious inputs start with stable source records"
        description="Capture books, lectures, films, podcasts, and other meaningful inputs before engagement work begins."
        actions={
          <Link
            to="/sources/new"
            className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
          >
            New source
          </Link>
        }
      />

      <section className="rounded-[2rem] border border-black/5 bg-white/70 p-6 shadow-card backdrop-blur">
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <label className="block">
            <span className="mb-2 block text-sm text-ink/75">Search title or creator</span>
            <input
              value={filters.q ?? ''}
              onChange={(event) => setFilters((current) => ({ ...current, q: event.target.value }))}
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            />
          </label>

          <label className="block">
            <span className="mb-2 block text-sm text-ink/75">Medium</span>
            <select
              value={filters.medium ?? ''}
              onChange={(event) => setFilters((current) => ({ ...current, medium: event.target.value }))}
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            >
              <option value="">All media</option>
              {SOURCE_MEDIA.map((medium) => (
                <option key={medium} value={medium}>
                  {medium.toLowerCase().replace(/_/g, ' ')}
                </option>
              ))}
            </select>
          </label>

          <label className="block">
            <span className="mb-2 block text-sm text-ink/75">Original language</span>
            <input
              value={filters.original_language ?? ''}
              onChange={(event) => setFilters((current) => ({ ...current, original_language: event.target.value }))}
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            />
          </label>

          <label className="block">
            <span className="mb-2 block text-sm text-ink/75">Sort</span>
            <select
              value={filters.sort ?? 'recent'}
              onChange={(event) =>
                setFilters((current) => ({ ...current, sort: event.target.value as ListSourcesFilters['sort'] }))
              }
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            >
              <option value="recent">Recently updated</option>
              <option value="title">Title</option>
            </select>
          </label>
        </div>
      </section>

      {error ? (
        <section className="rounded-[2rem] border border-amber-200 bg-amber-50 px-6 py-5 text-amber-700 shadow-card">
          {error}
        </section>
      ) : null}

      {loading ? (
        <section className="rounded-[2rem] border border-black/5 bg-white/70 px-6 py-8 shadow-card backdrop-blur">
          Loading sources...
        </section>
      ) : sources.length === 0 ? (
        <EmptyState
          title="No sources yet"
          body="Create the first stable source record so later engagement capture stays clean and duplicate-free."
          action={
            <Link
              to="/sources/new"
              className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
            >
              Create first source
            </Link>
          }
        />
      ) : (
        <section className="grid gap-5 xl:grid-cols-2">
          {sources.map((source) => (
            <SourceCard key={source.id} source={source} />
          ))}
        </section>
      )}
    </div>
  )
}
