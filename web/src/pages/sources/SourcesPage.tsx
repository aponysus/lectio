import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { type ListSourcesFilters, listSources, SOURCE_MEDIA, type Source } from '../../api/client'
import { SourceCard } from '../../components/sources/SourceCard'
import { SourceListRow } from '../../components/sources/SourceListRow'
import { EmptyState } from '../../components/shared/EmptyState'
import { formFieldClassName } from '../../components/shared/formStyles'
import { LoadingPanel } from '../../components/shared/LoadingPanel'
import { PageHeader } from '../../components/shared/PageHeader'
import { type BrowseViewMode, ViewModeToggle } from '../../components/shared/ViewModeToggle'

export function SourcesPage() {
  const [searchParams] = useSearchParams()
  const [viewMode, setViewMode] = useState<BrowseViewMode>('list')
  const [filters, setFilters] = useState<ListSourcesFilters>(() => ({
    q: searchParams.get('q') ?? '',
    medium: '',
    original_language: '',
    sort: 'recent',
  }))
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

      <section className="rounded-[1.5rem] border border-black/5 bg-white/70 p-5 shadow-card backdrop-blur">
        <div className="mb-4 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
          <p className="text-xs uppercase tracking-[0.22em] text-accent/80">Filters</p>
          <div className="flex flex-wrap items-center gap-3">
            <p className="text-sm text-ink/68">
              {sources.length} {sources.length === 1 ? 'source' : 'sources'}
              {loading ? ' loading' : ' visible'}
            </p>
            <ViewModeToggle value={viewMode} onChange={setViewMode} />
          </div>
        </div>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <label className="block">
            <span className="mb-2 block text-sm text-ink/75">Search title or creator</span>
            <input
              value={filters.q ?? ''}
              onChange={(event) => setFilters((current) => ({ ...current, q: event.target.value }))}
              className={formFieldClassName}
            />
          </label>

          <label className="block">
            <span className="mb-2 block text-sm text-ink/75">Medium</span>
            <select
              value={filters.medium ?? ''}
              onChange={(event) => setFilters((current) => ({ ...current, medium: event.target.value }))}
              className={formFieldClassName}
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
              className={formFieldClassName}
            />
          </label>

          <label className="block">
            <span className="mb-2 block text-sm text-ink/75">Sort</span>
            <select
              value={filters.sort ?? 'recent'}
              onChange={(event) =>
                setFilters((current) => ({ ...current, sort: event.target.value as ListSourcesFilters['sort'] }))
              }
              className={formFieldClassName}
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
        <LoadingPanel label="Loading sources" variant="list" />
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
      ) : viewMode === 'list' ? (
        <section className="space-y-3">
          {sources.map((source) => (
            <SourceListRow key={source.id} source={source} />
          ))}
        </section>
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
