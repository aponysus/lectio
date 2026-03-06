import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { ACCESS_MODES, listEngagements, type AccessMode, type Engagement, type ListEngagementsFilters } from '../../api/client'
import { EngagementCard } from '../../components/engagements/EngagementCard'
import { EngagementListRow } from '../../components/engagements/EngagementListRow'
import { EmptyState } from '../../components/shared/EmptyState'
import { formFieldClassName } from '../../components/shared/formStyles'
import { LoadingPanel } from '../../components/shared/LoadingPanel'
import { PageHeader } from '../../components/shared/PageHeader'
import { type BrowseViewMode, ViewModeToggle } from '../../components/shared/ViewModeToggle'

export function EngagementsPage() {
  const [searchParams] = useSearchParams()
  const [viewMode, setViewMode] = useState<BrowseViewMode>('list')
  const [filters, setFilters] = useState<ListEngagementsFilters>(() => ({
    q: searchParams.get('q') ?? '',
    access_mode: '',
    has_language_notes: false,
    limit: 30,
  }))
  const [engagements, setEngagements] = useState<Engagement[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        const nextEngagements = await listEngagements(filters)
        if (!cancelled) {
          setEngagements(nextEngagements)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load engagements')
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
        eyebrow="Engagements"
        title="Browse captured encounters"
        description="Filter the engagement corpus by access mode or by whether language notes exist, then open the records that need deeper work."
        actions={
          <Link
            to="/engagements/new"
            className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
          >
            New engagement
          </Link>
        }
      />

      <section className="rounded-[1.5rem] border border-black/5 bg-white/70 p-5 shadow-card backdrop-blur">
        <div className="mb-4 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
          <p className="text-xs uppercase tracking-[0.22em] text-accent/80">Filters</p>
          <div className="flex flex-wrap items-center gap-3">
            <p className="text-sm text-ink/68">
              {engagements.length} {engagements.length === 1 ? 'engagement' : 'engagements'}
              {loading ? ' loading' : ' visible'}
            </p>
            <ViewModeToggle value={viewMode} onChange={setViewMode} />
          </div>
        </div>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-[minmax(0,1.2fr)_minmax(0,0.8fr)_auto]">
          <label className="block">
            <span className="mb-2 block text-sm text-ink/75">Search reflection text</span>
            <input
              value={filters.q ?? ''}
              onChange={(event) => setFilters((current) => ({ ...current, q: event.target.value }))}
              className={formFieldClassName}
            />
          </label>

          <label className="block">
            <span className="mb-2 block text-sm text-ink/75">Access mode</span>
            <select
              value={filters.access_mode ?? ''}
              onChange={(event) =>
                setFilters((current) => ({ ...current, access_mode: event.target.value as AccessMode | '' }))
              }
              className={formFieldClassName}
            >
              <option value="">All access modes</option>
              {ACCESS_MODES.map((mode) => (
                <option key={mode} value={mode}>
                  {mode.toLowerCase().replace(/_/g, ' ')}
                </option>
              ))}
            </select>
          </label>

          <label className="flex items-end">
            <span className="flex items-center gap-3 rounded-2xl bg-black/[0.03] px-4 py-3 text-sm text-ink/80">
              <input
                type="checkbox"
                checked={Boolean(filters.has_language_notes)}
                onChange={(event) =>
                  setFilters((current) => ({ ...current, has_language_notes: event.target.checked }))
                }
                className="h-4 w-4 rounded border-black/20"
              />
              With language notes only
            </span>
          </label>
        </div>
      </section>

      {error ? (
        <section className="rounded-[2rem] border border-amber-200 bg-amber-50 px-6 py-5 text-amber-700 shadow-card">
          {error}
        </section>
      ) : null}

      {loading ? (
        <LoadingPanel label="Loading engagements" variant="list" />
      ) : engagements.length === 0 ? (
        <EmptyState
          title="No engagements match these filters"
          body="Try broadening the text query, adjusting access mode, or turning off the language-note filter."
          action={
            <button
              type="button"
              onClick={() => setFilters({ q: '', access_mode: '', has_language_notes: false, limit: 30 })}
              className="rounded-2xl bg-accent px-4 py-3 text-sm font-medium text-white transition hover:bg-accent/90"
            >
              Clear filters
            </button>
          }
        />
      ) : viewMode === 'list' ? (
        <section className="space-y-3">
          {engagements.map((engagement) => (
            <EngagementListRow key={engagement.id} engagement={engagement} />
          ))}
        </section>
      ) : (
        <section className="grid gap-5 xl:grid-cols-2">
          {engagements.map((engagement) => (
            <EngagementCard key={engagement.id} engagement={engagement} />
          ))}
        </section>
      )}
    </div>
  )
}
