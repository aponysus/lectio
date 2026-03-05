import { useQuery } from '@tanstack/react-query'
import ReactMarkdown from 'react-markdown'
import rehypeSanitize from 'rehype-sanitize'
import remarkGfm from 'remark-gfm'
import { Link } from 'react-router-dom'
import { apiFetch, isApiError } from '../api/client'

type Entry = {
  id: number
  source_id: number
  passage: string
  reflection: string
  mood: string
  energy: number | null
  tags: Array<{
    id: number
    slug: string
    label: string
  }>
  created_at: string
  updated_at: string
}

type EntryListResponse = {
  data: Entry[]
  meta: {
    page: number
    page_size: number
    total: number
    has_next: boolean
  }
}

const dashboardPrelude = `## Practice

Lectio works best when entry capture feels lighter than postponement.

- Save the raw thought while the passage is still alive.
- Add the thematic tags only when they actually clarify the note.
- Let recurrence, not forced structure, tell you what matters.
`

export function DashboardPage() {
  const entriesQuery = useQuery({
    queryKey: ['entries', 'dashboard'],
    queryFn: () => apiFetch<EntryListResponse>('/entries?page_size=6'),
    retry: false,
  })

  const entries = entriesQuery.data?.data ?? []
  const themeCounts = new Map<string, number>()
  for (const entry of entries) {
    for (const tag of entry.tags) {
      themeCounts.set(tag.label, (themeCounts.get(tag.label) ?? 0) + 1)
    }
  }
  const currentThemes = [...themeCounts.entries()]
    .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
    .slice(0, 5)

  const isUnauthorized = entriesQuery.isError && isApiError(entriesQuery.error) && entriesQuery.error.status === 401

  return (
    <section className="page">
      <section className="hero-grid">
        <article className="hero-card hero-card-feature">
          <p className="eyebrow">Today&apos;s desk</p>
          <h2>Write while the thought is still warm.</h2>
          <p className="hero-copy">
            The first screen should lower resistance: capture the source, passage, and reflection now. Structure can follow.
          </p>
          <div className="hero-actions">
            <Link className="button" to="/entries/new">
              Start a new entry
            </Link>
            <Link className="button button-secondary" to="/login">
              {isUnauthorized ? 'Sign in to continue' : 'Review session'}
            </Link>
          </div>
        </article>

        <article className="hero-card hero-card-note">
          <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeSanitize]}>
            {dashboardPrelude}
          </ReactMarkdown>
        </article>
      </section>

      <section className="dashboard-grid">
        <article className="panel panel-tall">
          <div className="panel-header">
            <div>
              <p className="eyebrow">Recent entries</p>
              <h3>Latest reflections</h3>
            </div>
            <span className="meta-pill">{entries.length} loaded</span>
          </div>

          {entriesQuery.isPending ? <p className="muted-copy">Loading your recent notes...</p> : null}
          {isUnauthorized ? (
            <p className="muted-copy">
              Sign in to see your reading history and start building resonances between entries.
            </p>
          ) : null}
          {entriesQuery.isError && !isUnauthorized ? (
            <p className="error">Could not load recent entries.</p>
          ) : null}
          {!entriesQuery.isPending && !entries.length && !isUnauthorized ? (
            <p className="muted-copy">Your first entry will show up here as soon as you save it.</p>
          ) : null}

          <div className="entry-list">
            {entries.map((entry) => (
              <article key={entry.id} className="entry-card">
                <div className="entry-card-header">
                  <span className="meta-pill">Source #{entry.source_id}</span>
                  <time>{formatShortDate(entry.created_at)}</time>
                </div>
                <p className="entry-excerpt">{excerpt(entry.reflection, 180)}</p>
                <div className="tag-row">
                  {entry.tags.map((tag) => (
                    <span key={tag.id} className="tag-pill">
                      {tag.label}
                    </span>
                  ))}
                </div>
              </article>
            ))}
          </div>
        </article>

        <div className="dashboard-stack">
          <article className="panel">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Resonance</p>
                <h3>Daily revisit</h3>
              </div>
            </div>
            <p className="muted-copy">
              The daily resonance surface will become useful after a small body of entries exists. For now, focus on capture quality.
            </p>
          </article>

          <article className="panel">
            <div className="panel-header">
              <div>
                <p className="eyebrow">Themes</p>
                <h3>What is recurring</h3>
              </div>
            </div>
            {currentThemes.length ? (
              <div className="theme-list">
                {currentThemes.map(([label, count]) => (
                  <div key={label} className="theme-row">
                    <span>{label}</span>
                    <strong>{count}</strong>
                  </div>
                ))}
              </div>
            ) : (
              <p className="muted-copy">Theme density appears once you begin tagging entries consistently.</p>
            )}
          </article>
        </div>
      </section>
    </section>
  )
}

function excerpt(value: string, limit: number) {
  if (value.length <= limit) {
    return value
  }
  return `${value.slice(0, limit).trimEnd()}...`
}

function formatShortDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}
