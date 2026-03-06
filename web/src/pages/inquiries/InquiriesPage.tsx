import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { INQUIRY_STATUSES, type Inquiry, type ListInquiriesFilters, listInquiries } from '../../api/client'
import { InquiryCard } from '../../components/inquiries/InquiryCard'
import { EmptyState } from '../../components/shared/EmptyState'
import { PageHeader } from '../../components/shared/PageHeader'

export function InquiriesPage() {
  const [filters, setFilters] = useState<ListInquiriesFilters>({
    q: '',
    status: '',
  })
  const [inquiries, setInquiries] = useState<Inquiry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        const nextInquiries = await listInquiries(filters)
        if (!cancelled) {
          setInquiries(nextInquiries)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load inquiries')
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
        eyebrow="Inquiries"
        title="Organize the work around live questions"
        description="Inquiry pages turn the app from archive storage into a thinking tool. Keep the active ones sharp and visible."
        actions={
          <Link
            to="/inquiries/new"
            className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
          >
            New inquiry
          </Link>
        }
      />

      <section className="rounded-[2rem] border border-black/5 bg-white/70 p-6 shadow-card backdrop-blur">
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <label className="block xl:col-span-2">
            <span className="mb-2 block text-sm text-ink/75">Search title or question</span>
            <input
              value={filters.q ?? ''}
              onChange={(event) => setFilters((current) => ({ ...current, q: event.target.value }))}
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            />
          </label>

          <label className="block">
            <span className="mb-2 block text-sm text-ink/75">Status</span>
            <select
              value={filters.status ?? ''}
              onChange={(event) => setFilters((current) => ({ ...current, status: event.target.value }))}
              className="w-full rounded-2xl border border-black/10 bg-canvas/80 px-4 py-3 outline-none transition focus:border-accent"
            >
              <option value="">All statuses</option>
              {INQUIRY_STATUSES.map((status) => (
                <option key={status} value={status}>
                  {status.toLowerCase().replace(/_/g, ' ')}
                </option>
              ))}
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
          Loading inquiries...
        </section>
      ) : inquiries.length === 0 ? (
        <EmptyState
          title="No inquiries yet"
          body="Create the first live question so engagements can start accumulating into a real workspace instead of a loose pile."
          action={
            <Link
              to="/inquiries/new"
              className="rounded-2xl bg-pine px-4 py-3 text-sm font-medium text-white transition hover:bg-pine/90"
            >
              Create first inquiry
            </Link>
          }
        />
      ) : (
        <section className="grid gap-5 xl:grid-cols-2">
          {inquiries.map((inquiry) => (
            <InquiryCard key={inquiry.id} inquiry={inquiry} />
          ))}
        </section>
      )}
    </div>
  )
}
