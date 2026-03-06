import type { ReactNode } from 'react'
import { NavLink } from 'react-router-dom'
import { useSession } from '../../hooks/useSession'

const links = [
  { to: '/', label: 'Dashboard' },
  { to: '/sources', label: 'Sources' },
  { to: '/engagements/new', label: 'New Engagement' },
  { to: '/inquiries', label: 'Inquiries' },
  { to: '/syntheses', label: 'Syntheses' },
]

export function AppShell({ children }: { children: ReactNode }) {
  const { session, logout } = useSession()

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,_rgba(139,94,52,0.18),_transparent_28%),linear-gradient(180deg,_#f4efe4_0%,_#efe5d2_100%)] text-ink">
      <div className="mx-auto flex min-h-screen w-full max-w-7xl flex-col gap-6 px-4 py-4 md:px-6 lg:flex-row lg:gap-8 lg:py-8">
        <aside className="w-full rounded-[2rem] border border-black/5 bg-white/70 p-5 shadow-card backdrop-blur lg:w-80">
          <div className="mb-8">
            <p className="text-xs uppercase tracking-[0.3em] text-accent/80">Lectio</p>
            <h1 className="mt-2 font-display text-3xl text-ink">Built for serious source work</h1>
            <p className="mt-3 text-sm leading-6 text-ink/70">
              The MVP loop now includes live source records, engagement capture, inquiry workspaces, claims, and synthesis.
            </p>
          </div>

          <nav className="space-y-2">
            {links.map((link) => (
              <NavLink
                key={link.to}
                to={link.to}
                className={({ isActive }) =>
                  `flex items-center justify-between rounded-2xl px-4 py-3 text-sm transition ${
                    isActive ? 'bg-pine text-white shadow-sm' : 'bg-stone-900/[0.03] text-ink hover:bg-stone-900/[0.06]'
                  }`
                }
              >
                <span>{link.label}</span>
                <span className="text-xs opacity-70">
                  {link.to === '/' || link.to === '/sources' || link.to === '/engagements/new' || link.to === '/inquiries' || link.to === '/syntheses'
                    ? 'Live'
                    : 'Later'}
                </span>
              </NavLink>
            ))}
          </nav>

          <div className="mt-8 rounded-2xl border border-black/5 bg-canvas/80 px-4 py-4">
            <p className="text-xs uppercase tracking-[0.25em] text-accent/80">Session</p>
            <p className="mt-2 text-sm text-ink/80">{session.user_id ?? 'Authenticated user'}</p>
            <button
              type="button"
              onClick={() => void logout()}
              className="mt-4 w-full rounded-xl bg-ink px-3 py-2 text-sm text-white transition hover:bg-ink/90"
            >
              Sign out
            </button>
          </div>
        </aside>

        <main className="flex-1">{children}</main>
      </div>
    </div>
  )
}
