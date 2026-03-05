import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { NavLink, Route, Routes } from 'react-router-dom'
import { apiFetch, isApiError } from './api/client'
import { DashboardPage } from './pages/DashboardPage'
import { LoginPage } from './pages/LoginPage'
import { NewEntryPage } from './pages/NewEntryPage'

type EntryListResponse = {
  data: Array<{ id: number }>
  meta: {
    page: number
    page_size: number
    total: number
    has_next: boolean
  }
}

export function App() {
  const queryClient = useQueryClient()
  const sessionQuery = useQuery({
    queryKey: ['session'],
    queryFn: () => apiFetch<EntryListResponse>('/entries?page_size=1'),
    retry: false,
    staleTime: 30_000,
  })

  const logoutMutation = useMutation({
    mutationFn: () => apiFetch<void>('/auth/logout', { method: 'POST' }),
    onSuccess: async () => {
      await queryClient.invalidateQueries()
    },
  })

  const isAuthenticated = sessionQuery.isSuccess
  const isUnauthorized = sessionQuery.isError && isApiError(sessionQuery.error) && sessionQuery.error.status === 401

  return (
    <div className="app-shell">
      <div className="ambient ambient-left" />
      <div className="ambient ambient-right" />

      <header className="shell-header">
        <div className="brand-block">
          <p className="eyebrow">Contemplative reading companion</p>
          <div className="brand-row">
            <div>
              <h1>Lectio</h1>
              <p className="brand-copy">
                Capture live reading thoughts, revisit older notes, and let patterns emerge over time.
              </p>
            </div>
            <div className="session-chip">
              <span className={`session-dot ${isAuthenticated ? 'session-dot-live' : ''}`} />
              {isAuthenticated ? 'Signed in' : isUnauthorized ? 'Guest mode' : 'Checking session'}
            </div>
          </div>
        </div>

        <div className="shell-toolbar">
          <nav className="topnav" aria-label="Primary">
            <NavLink to="/">Dashboard</NavLink>
            <NavLink to="/entries/new">New Entry</NavLink>
            <NavLink to="/sources">Sources</NavLink>
            <NavLink to="/tags">Tags</NavLink>
            <NavLink to="/timeline">Timeline</NavLink>
          </nav>

          <div className="toolbar-actions">
            {isAuthenticated ? (
              <button
                type="button"
                className="button button-secondary"
                onClick={() => logoutMutation.mutate()}
                disabled={logoutMutation.isPending}
              >
                {logoutMutation.isPending ? 'Signing out...' : 'Sign out'}
              </button>
            ) : (
              <NavLink className="button button-secondary nav-button" to="/login">
                Sign in
              </NavLink>
            )}
          </div>
        </div>
      </header>

      <main className="shell-main">
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/entries/new" element={<NewEntryPage />} />
          <Route path="/sources" element={<Placeholder title="Source library" body="Source browse and autocomplete are next. The backend entry flow already supports source creation on submit." />} />
          <Route path="/tags" element={<Placeholder title="Tag atlas" body="Weighted tags and thematic browse will build on the tags now created during entry capture." />} />
          <Route path="/timeline" element={<Placeholder title="Intellectual timeline" body="Chronological browse is ready for a stronger filter surface once source and tag views are in place." />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="*" element={<DashboardPage />} />
        </Routes>
      </main>
    </div>
  )
}

function Placeholder({ title, body }: { title: string; body: string }) {
  return (
    <section className="page page-single">
      <div className="hero-card">
        <p className="eyebrow">In progress</p>
        <h2>{title}</h2>
        <p className="hero-copy">{body}</p>
      </div>
    </section>
  )
}
