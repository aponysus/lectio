import { type ReactNode, useDeferredValue, useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import {
  listClaims,
  listEngagements,
  listInquiries,
  listSources,
  type Claim,
  type Engagement,
  type Inquiry,
  type Source,
} from '../api/client'
import { ClaimListRow } from '../components/claims/ClaimListRow'
import { EngagementListRow } from '../components/engagements/EngagementListRow'
import { InquiryListRow } from '../components/inquiries/InquiryListRow'
import { EmptyState } from '../components/shared/EmptyState'
import { formFieldClassName } from '../components/shared/formStyles'
import { LoadingPanel } from '../components/shared/LoadingPanel'
import { PageHeader } from '../components/shared/PageHeader'
import { SourceListRow } from '../components/sources/SourceListRow'

const sectionResultLimit = 6

type SearchResults = {
  sources: Source[]
  inquiries: Inquiry[]
  engagements: Engagement[]
  claims: Claim[]
}

const emptyResults: SearchResults = {
  sources: [],
  inquiries: [],
  engagements: [],
  claims: [],
}

export function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const query = searchParams.get('q') ?? ''
  const deferredQuery = useDeferredValue(query)
  const [inputValue, setInputValue] = useState(query)
  const [results, setResults] = useState<SearchResults>(emptyResults)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setInputValue(query)
  }, [query])

  useEffect(() => {
    const nextQuery = inputValue.trim()
    if (nextQuery === query) {
      return
    }

    const timeoutID = window.setTimeout(() => {
      setSearchParams(nextQuery ? { q: nextQuery } : {}, { replace: true })
    }, 180)

    return () => {
      window.clearTimeout(timeoutID)
    }
  }, [inputValue, query, setSearchParams])

  useEffect(() => {
    if (!deferredQuery) {
      setResults(emptyResults)
      setLoading(false)
      setError(null)
      return
    }

    let cancelled = false
    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        const [sources, inquiries, engagements, claims] = await Promise.all([
          listSources({ q: deferredQuery, limit: sectionResultLimit }),
          listInquiries({ q: deferredQuery, limit: sectionResultLimit }),
          listEngagements({ q: deferredQuery, limit: sectionResultLimit }),
          listClaims({ q: deferredQuery, limit: sectionResultLimit }),
        ])

        if (!cancelled) {
          setResults({
            sources,
            inquiries,
            engagements,
            claims,
          })
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to search corpus')
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
  }, [deferredQuery])

  const totalResults =
    results.sources.length +
    results.inquiries.length +
    results.engagements.length +
    results.claims.length
  const sectionCount = [
    results.sources.length,
    results.inquiries.length,
    results.engagements.length,
    results.claims.length,
  ].filter((count) => count > 0).length

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Search"
        title="Search the working corpus"
        description="Find a remembered fragment across sources, live questions, engagements, and extracted claims without leaving the app’s main loop."
      />

      <section className="rounded-[1.5rem] border border-black/5 bg-white/70 p-5 shadow-card backdrop-blur">
        <div className="grid gap-4 xl:grid-cols-[minmax(0,1.2fr)_auto] xl:items-end">
          <label className="block">
            <span className="mb-2 block text-sm text-ink/75">Search title, creator, question, reflection, or claim text</span>
            <input
              value={inputValue}
              onChange={(event) => setInputValue(event.target.value)}
              placeholder="Try a title fragment, author name, question, or remembered phrase"
              className={formFieldClassName}
            />
          </label>

          <div className="rounded-[1.25rem] bg-black/[0.03] px-4 py-3 text-sm leading-6 text-ink/74">
            {deferredQuery
              ? `${totalResults} visible across ${sectionCount} ${sectionCount === 1 ? 'section' : 'sections'}`
              : 'Showing up to 6 results per section once you start typing.'}
          </div>
        </div>
      </section>

      {error ? (
        <section className="rounded-[2rem] border border-amber-200 bg-amber-50 px-6 py-5 text-amber-700 shadow-card">
          {error}
        </section>
      ) : null}

      {!deferredQuery ? (
        <EmptyState
          title="Search is ready"
          body="Start with whatever you remember: a source title, a creator, a question fragment, a reflection phrase, or claim wording."
        />
      ) : loading ? (
        <LoadingPanel label="Searching corpus" variant="list" />
      ) : totalResults === 0 ? (
        <EmptyState
          title={`No matches for "${deferredQuery}"`}
          body="Try a shorter phrase, a creator surname, or a more distinctive fragment from the original note."
          action={
            <button
              type="button"
              onClick={() => setInputValue('')}
              className="rounded-2xl bg-accent px-4 py-3 text-sm font-medium text-white transition hover:bg-accent/90"
            >
              Clear search
            </button>
          }
        />
      ) : (
        <div className="space-y-5">
          {results.sources.length > 0 ? (
            <SearchSection
              eyebrow="Sources"
              title="Inputs"
              count={results.sources.length}
              browseHref={`/sources?q=${encodeURIComponent(deferredQuery)}`}
            >
              {results.sources.map((source) => (
                <SourceListRow key={source.id} source={source} />
              ))}
            </SearchSection>
          ) : null}

          {results.inquiries.length > 0 ? (
            <SearchSection
              eyebrow="Inquiries"
              title="Live questions"
              count={results.inquiries.length}
              browseHref={`/inquiries?q=${encodeURIComponent(deferredQuery)}`}
            >
              {results.inquiries.map((inquiry) => (
                <InquiryListRow key={inquiry.id} inquiry={inquiry} />
              ))}
            </SearchSection>
          ) : null}

          {results.engagements.length > 0 ? (
            <SearchSection
              eyebrow="Engagements"
              title="Captured encounters"
              count={results.engagements.length}
              browseHref={`/engagements?q=${encodeURIComponent(deferredQuery)}`}
            >
              {results.engagements.map((engagement) => (
                <EngagementListRow key={engagement.id} engagement={engagement} />
              ))}
            </SearchSection>
          ) : null}

          {results.claims.length > 0 ? (
            <SearchSection eyebrow="Claims" title="Sharpened statements" count={results.claims.length}>
              {results.claims.map((claim) => (
                <ClaimListRow key={claim.id} claim={claim} />
              ))}
            </SearchSection>
          ) : null}
        </div>
      )}
    </div>
  )
}

function SearchSection({
  eyebrow,
  title,
  count,
  browseHref,
  children,
}: {
  eyebrow: string
  title: string
  count: number
  browseHref?: string
  children: ReactNode
}) {
  return (
    <section className="rounded-[1.5rem] border border-black/5 bg-white/70 p-5 shadow-card backdrop-blur">
      <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.22em] text-accent/80">{eyebrow}</p>
          <h3 className="mt-2 font-display text-[1.8rem] leading-tight text-ink">{title}</h3>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <p className="text-sm text-ink/68">
            {count} {count === 1 ? 'result' : 'results'}
          </p>
          {browseHref ? (
            <Link
              to={browseHref}
              className="rounded-xl border border-black/10 bg-white/90 px-3 py-2 text-sm text-ink transition hover:bg-white"
            >
              Open full list
            </Link>
          ) : null}
        </div>
      </div>

      <div className="mt-5 space-y-3">{children}</div>
    </section>
  )
}
