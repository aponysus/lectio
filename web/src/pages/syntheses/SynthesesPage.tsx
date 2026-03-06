import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listSyntheses, type Synthesis } from '../../api/client'
import { SynthesisCard } from '../../components/syntheses/SynthesisCard'
import { SynthesisListRow } from '../../components/syntheses/SynthesisListRow'
import { EmptyState } from '../../components/shared/EmptyState'
import { LoadingPanel } from '../../components/shared/LoadingPanel'
import { PageHeader } from '../../components/shared/PageHeader'
import { type BrowseViewMode, ViewModeToggle } from '../../components/shared/ViewModeToggle'

export function SynthesesPage() {
  const [viewMode, setViewMode] = useState<BrowseViewMode>('list')
  const [syntheses, setSyntheses] = useState<Synthesis[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        const nextSyntheses = await listSyntheses(50)
        if (!cancelled) {
          setSyntheses(nextSyntheses)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load syntheses')
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
  }, [])

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Syntheses"
        title="Compressed inquiry checkpoints"
        description="This is the payoff layer: stored attempts to compress linked engagements and claims into a clearer current position."
        actions={
          <>
            <ViewModeToggle value={viewMode} onChange={setViewMode} />
            <Link
              to="/inquiries"
              className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
            >
              Choose inquiry
            </Link>
          </>
        }
      />

      <section className="rounded-[1.5rem] border border-black/5 bg-white/70 px-5 py-4 shadow-card backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <p className="text-xs uppercase tracking-[0.22em] text-accent/80">Browse</p>
          <p className="text-sm text-ink/68">
            {syntheses.length} {syntheses.length === 1 ? 'synthesis' : 'syntheses'}
            {loading ? ' loading' : ' visible'}
          </p>
        </div>
      </section>

      {error ? (
        <section className="rounded-[2rem] border border-amber-200 bg-amber-50 px-6 py-5 text-amber-700 shadow-card">
          {error}
        </section>
      ) : null}

      {loading ? (
        <LoadingPanel label="Loading syntheses" variant="list" />
      ) : syntheses.length === 0 ? (
        <EmptyState
          title="No syntheses yet"
          body="The core loop now supports synthesis, but every synthesis should start from an inquiry. Open an inquiry or a dashboard prompt to write the first one."
          action={
            <Link
              to="/inquiries"
              className="rounded-2xl bg-accent px-4 py-3 text-sm font-medium text-white transition hover:bg-accent/90"
            >
              Open inquiries
            </Link>
          }
        />
      ) : viewMode === 'list' ? (
        <section className="space-y-3">
          {syntheses.map((synthesis) => (
            <SynthesisListRow key={synthesis.id} synthesis={synthesis} />
          ))}
        </section>
      ) : (
        <section className="grid gap-5 xl:grid-cols-2">
          {syntheses.map((synthesis) => (
            <SynthesisCard key={synthesis.id} synthesis={synthesis} />
          ))}
        </section>
      )}
    </div>
  )
}
