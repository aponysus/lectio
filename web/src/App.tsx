import { NavLink, Route, Routes } from 'react-router-dom'
import { DashboardPage } from './pages/DashboardPage'
import { LoginPage } from './pages/LoginPage'

export function App() {
  return (
    <div className="app-shell">
      <header className="topbar">
        <h1>Lectio</h1>
        <nav>
          <NavLink to="/">Dashboard</NavLink>
          <NavLink to="/entries/new">New Entry</NavLink>
          <NavLink to="/sources">Sources</NavLink>
          <NavLink to="/tags">Tags</NavLink>
          <NavLink to="/timeline">Timeline</NavLink>
          <NavLink to="/login">Login</NavLink>
        </nav>
      </header>

      <main>
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/entries/new" element={<Placeholder title="Entry form scaffold" />} />
          <Route path="/sources" element={<Placeholder title="Source library scaffold" />} />
          <Route path="/tags" element={<Placeholder title="Tag browsing scaffold" />} />
          <Route path="/timeline" element={<Placeholder title="Timeline scaffold" />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="*" element={<DashboardPage />} />
        </Routes>
      </main>
    </div>
  )
}

function Placeholder({ title }: { title: string }) {
  return (
    <section className="panel">
      <h2>{title}</h2>
      <p>Page route is wired. API integration will be added during feature implementation.</p>
    </section>
  )
}
