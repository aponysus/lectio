import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useState,
} from 'react'
import {
  Link,
  NavLink,
  matchPath,
  useLocation,
} from 'react-router-dom'
import { useSession } from '../../hooks/useSession'

const links = [
  { to: '/', label: 'Dashboard' },
  { to: '/search', label: 'Search' },
  { to: '/sources', label: 'Sources' },
  { to: '/engagements', label: 'Engagements' },
  { to: '/inquiries', label: 'Inquiries' },
  { to: '/syntheses', label: 'Syntheses' },
]

type WorkspaceSidebarLink = {
  label: string
  href: string
  detail?: string
}

type WorkspaceSidebarAction = {
  label: string
  href: string
  tone?: 'primary' | 'secondary'
}

type WorkspaceSidebarStat = {
  label: string
  value: string
}

export type WorkspaceSidebarContent = {
  eyebrow: string
  title: string
  body: string
  links?: WorkspaceSidebarLink[]
  actions?: WorkspaceSidebarAction[]
  stats?: WorkspaceSidebarStat[]
}

type WorkspaceSidebarEntry = {
  pathname: string
  content: WorkspaceSidebarContent | null
}

type WorkspaceSidebarContextValue = (pathname: string, content: WorkspaceSidebarContent | null) => void

const WorkspaceSidebarContext = createContext<WorkspaceSidebarContextValue | null>(null)

export function useWorkspaceSidebar() {
  const setSidebarContent = useContext(WorkspaceSidebarContext)
  const location = useLocation()

  if (!setSidebarContent) {
    throw new Error('useWorkspaceSidebar must be used inside AppShell')
  }

  return useCallback(
    (content: WorkspaceSidebarContent | null) => {
      setSidebarContent(location.pathname, content)
    },
    [location.pathname, setSidebarContent],
  )
}

export function AppShell({ children }: { children: ReactNode }) {
  const { session, logout } = useSession()
  const location = useLocation()
  const [workspaceSidebar, setWorkspaceSidebar] = useState<WorkspaceSidebarEntry | null>(null)
  const handleSetSidebarContent = useCallback((pathname: string, content: WorkspaceSidebarContent | null) => {
    setWorkspaceSidebar({ pathname, content })
  }, [])
  const activeSidebarContent =
    workspaceSidebar && workspaceSidebar.pathname === location.pathname
      ? workspaceSidebar.content
      : null
  const sidebarContent = activeSidebarContent ?? getDefaultSidebarContent(location.pathname)

  return (
    <WorkspaceSidebarContext.Provider value={handleSetSidebarContent}>
      <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,_rgba(139,94,52,0.18),_transparent_28%),linear-gradient(180deg,_#f4efe4_0%,_#efe5d2_100%)] text-ink">
        <div className="mx-auto flex min-h-screen w-full max-w-7xl flex-col gap-5 px-4 py-4 md:px-6 lg:flex-row lg:gap-6 lg:py-6">
          <aside className="w-full rounded-[1.5rem] border border-black/5 bg-white/72 p-4 shadow-card backdrop-blur lg:w-72">
            <div className="mb-6">
              <p className="text-xs uppercase tracking-[0.26em] text-accent/80">Lectio</p>
              <h1 className="mt-2 font-display text-[2rem] leading-tight text-ink">Workspace</h1>
              <p className="mt-2 text-sm leading-6 text-ink/70">Capture, connect, revisit, compress.</p>
            </div>

            <div className="mb-6 rounded-[1.25rem] border border-black/6 bg-canvas/85 p-4">
              <p className="text-xs uppercase tracking-[0.22em] text-accent/80">Quick actions</p>
              <div className="mt-3 grid gap-2">
                <Link
                  to="/engagements/new"
                  className="rounded-xl bg-pine px-3 py-2 text-sm font-medium text-white transition hover:bg-pine/90"
                >
                  Log engagement
                </Link>
                <Link
                  to="/inquiries/new"
                  className="rounded-xl border border-black/10 bg-white/90 px-3 py-2 text-sm text-ink transition hover:bg-white"
                >
                  New inquiry
                </Link>
                <Link
                  to="/sources/new"
                  className="rounded-xl border border-black/10 bg-white/90 px-3 py-2 text-sm text-ink transition hover:bg-white"
                >
                  New source
                </Link>
              </div>
            </div>

            <nav className="space-y-2">
              {links.map((link) => (
                <NavLink
                  key={link.to}
                  to={link.to}
                  className={({ isActive }) =>
                    `flex items-center rounded-xl border px-3 py-3 text-sm transition ${
                      isActive
                        ? 'border-pine/20 bg-pine/10 text-pine shadow-sm'
                        : 'border-transparent bg-stone-900/[0.03] text-ink hover:bg-stone-900/[0.06]'
                    }`
                  }
                >
                  {({ isActive }) => (
                    <span className="flex items-center gap-3">
                      <span className={`h-2.5 w-2.5 rounded-full ${isActive ? 'bg-pine' : 'bg-black/12'}`} />
                      <span>{link.label}</span>
                    </span>
                  )}
                </NavLink>
              ))}
            </nav>

            <div className="mt-6">
              <WorkspaceContextPanel content={sidebarContent} />
            </div>

            <div className="mt-6 rounded-[1.25rem] border border-black/5 bg-white/78 px-4 py-4">
              <p className="text-xs uppercase tracking-[0.22em] text-accent/80">Session</p>
              <p className="mt-2 text-sm text-ink/80">{session.user_id ?? 'Authenticated user'}</p>
              <Link
                to="/settings/export"
                className="mt-4 block rounded-xl border border-black/10 bg-white/90 px-3 py-2 text-center text-sm text-ink transition hover:bg-white"
              >
                Export data
              </Link>
              <button
                type="button"
                onClick={() => void logout()}
                className="mt-3 w-full rounded-xl bg-ink px-3 py-2 text-sm text-white transition hover:bg-ink/90"
              >
                Sign out
              </button>
            </div>
          </aside>

          <main className="flex-1">{children}</main>
        </div>
      </div>
    </WorkspaceSidebarContext.Provider>
  )
}

function WorkspaceContextPanel({ content }: { content: WorkspaceSidebarContent }) {
  return (
    <section className="rounded-[1.25rem] border border-black/6 bg-white/82 px-4 py-4">
      <p className="text-xs uppercase tracking-[0.22em] text-accent/80">{content.eyebrow}</p>
      <h2 className="mt-2 font-display text-[1.55rem] leading-tight text-ink">{content.title}</h2>
      <p className="mt-3 text-sm leading-6 text-ink/74">{content.body}</p>

      {content.stats && content.stats.length > 0 ? (
        <dl className="mt-4 grid grid-cols-2 gap-2">
          {content.stats.map((stat) => (
            <div key={stat.label} className="rounded-xl bg-black/[0.03] px-3 py-3">
              <dt className="text-[0.68rem] uppercase tracking-[0.18em] text-accent/75">{stat.label}</dt>
              <dd className="mt-1 text-sm text-ink/82">{stat.value}</dd>
            </div>
          ))}
        </dl>
      ) : null}

      {content.links && content.links.length > 0 ? (
        <div className="mt-4 space-y-2">
          {content.links.map((link) => (
            <SidebarLinkRow key={`${link.href}-${link.label}`} link={link} />
          ))}
        </div>
      ) : null}

      {content.actions && content.actions.length > 0 ? (
        <div className="mt-4 grid gap-2">
          {content.actions.map((action) =>
            action.href.startsWith('#') || action.href.startsWith('/api/') ? (
              <a
                key={`${action.href}-${action.label}`}
                href={action.href}
                className={action.tone === 'primary' ? primaryActionClassName : secondaryActionClassName}
              >
                {action.label}
              </a>
            ) : (
              <Link
                key={`${action.href}-${action.label}`}
                to={action.href}
                className={action.tone === 'primary' ? primaryActionClassName : secondaryActionClassName}
              >
                {action.label}
              </Link>
            ),
          )}
        </div>
      ) : null}
    </section>
  )
}

function SidebarLinkRow({ link }: { link: WorkspaceSidebarLink }) {
  const className =
    'flex items-start justify-between gap-3 rounded-xl bg-black/[0.03] px-3 py-3 text-sm text-ink transition hover:bg-black/[0.05]'

  const content = (
    <>
      <span>{link.label}</span>
      {link.detail ? <span className="text-xs uppercase tracking-[0.18em] text-accent/70">{link.detail}</span> : null}
    </>
  )

  if (link.href.startsWith('#')) {
    return (
      <a href={link.href} className={className}>
        {content}
      </a>
    )
  }

  return (
    <Link to={link.href} className={className}>
      {content}
    </Link>
  )
}

function getDefaultSidebarContent(pathname: string): WorkspaceSidebarContent {
  if (pathname === '/') {
    return {
      eyebrow: 'Today',
      title: 'Desk',
      body: 'Use the dashboard to decide the next move: revisit a prompt, continue an inquiry, or log a fresh engagement before the thread cools off.',
      links: [
        { label: 'Review rediscovery prompts', href: '/', detail: 'daily' },
        { label: 'Browse active inquiries', href: '/inquiries', detail: 'questions' },
      ],
      actions: [
        { label: 'Log engagement', href: '/engagements/new', tone: 'primary' },
        { label: 'New inquiry', href: '/inquiries/new' },
      ],
    }
  }

  if (pathname === '/search') {
    return {
      eyebrow: 'Corpus search',
      title: 'Find the thread',
      body: 'Search across sources, inquiries, engagements, and claims when you remember a fragment but not where it lives.',
      links: [
        { label: 'Browse sources', href: '/sources', detail: 'inputs' },
        { label: 'Browse inquiries', href: '/inquiries', detail: 'questions' },
        { label: 'Browse engagements', href: '/engagements', detail: 'captures' },
      ],
      actions: [
        { label: 'Log engagement', href: '/engagements/new', tone: 'primary' },
        { label: 'New inquiry', href: '/inquiries/new' },
      ],
    }
  }

  if (pathname === '/settings/export') {
    return {
      eyebrow: 'Backup',
      title: 'Export your work',
      body: 'Download a JSON snapshot of the core Lectio tables so you can keep an offline backup or use it later for migration work.',
      links: [
        { label: 'Return to dashboard', href: '/', detail: 'workspace' },
        { label: 'Open search', href: '/search', detail: 'lookup' },
      ],
      actions: [
        { label: 'Download export', href: '/api/export', tone: 'primary' },
        { label: 'Back to dashboard', href: '/' },
      ],
    }
  }

  if (pathname === '/sources/new' || matchPath('/sources/:sourceId/edit', pathname)) {
    return {
      eyebrow: 'Source form',
      title: 'Keep it minimal',
      body: 'Only capture what helps you recognize the source later and connect it to real reading, listening, or viewing work.',
      links: [
        { label: 'Back to sources', href: '/sources', detail: 'library' },
        { label: 'Why sources matter', href: '/sources', detail: 'capture loop' },
      ],
      actions: [{ label: 'Open source list', href: '/sources' }],
    }
  }

  if (pathname === '/sources' || matchPath('/sources/:sourceId', pathname)) {
    return {
      eyebrow: 'Source library',
      title: 'Inputs first',
      body: 'Keep sources lean and usable. The record only needs enough structure to support later engagement capture and rediscovery.',
      links: [
        { label: 'Browse all sources', href: '/sources', detail: 'library' },
        { label: 'Log engagement from a source', href: '/engagements/new', detail: 'capture' },
      ],
      actions: [
        { label: 'New source', href: '/sources/new', tone: 'primary' },
        { label: 'Open engagements', href: '/engagements' },
      ],
    }
  }

  if (pathname === '/engagements/new' || matchPath('/engagements/:engagementId/edit', pathname)) {
    return {
      eyebrow: 'Capture form',
      title: 'Shortest honest pass',
      body: 'Get the encounter down first. Claims, inquiry links, and language notes stay available, but the core move is recording what changed in this session.',
      links: [
        { label: 'Browse engagements', href: '/engagements', detail: 'recent' },
        { label: 'Open inquiries', href: '/inquiries', detail: 'attach' },
      ],
      actions: [{ label: 'Back to engagement list', href: '/engagements' }],
    }
  }

  if (pathname === '/engagements' || matchPath('/engagements/:engagementId', pathname)) {
    return {
      eyebrow: 'Engagement desk',
      title: 'Capture while it is live',
      body: 'The goal is to preserve what happened in the encounter while the texture is still available, then connect it to an inquiry if one is already present.',
      links: [
        { label: 'Browse recent engagements', href: '/engagements', detail: 'stream' },
        { label: 'Review linked inquiries', href: '/inquiries', detail: 'context' },
      ],
      actions: [
        { label: 'Log engagement', href: '/engagements/new', tone: 'primary' },
        { label: 'New source', href: '/sources/new' },
      ],
    }
  }

  if (pathname === '/inquiries/new' || matchPath('/inquiries/:inquiryId/edit', pathname)) {
    return {
      eyebrow: 'Inquiry form',
      title: 'Question design',
      body: 'The inquiry only needs enough shape to orient later work: the live question, why it matters, and the tensions you expect to revisit.',
      links: [
        { label: 'Browse inquiries', href: '/inquiries', detail: 'workspace' },
        { label: 'Open syntheses', href: '/syntheses', detail: 'later' },
      ],
      actions: [{ label: 'Back to inquiries', href: '/inquiries' }],
    }
  }

  if (pathname === '/inquiries' || matchPath('/inquiries/:inquiryId', pathname)) {
    return {
      eyebrow: 'Inquiry workspace',
      title: 'Live questions',
      body: 'Use inquiries to gather evidence around a durable question, track the current view, and decide when the material is ready for synthesis.',
      links: [
        { label: 'Browse inquiries', href: '/inquiries', detail: 'questions' },
        { label: 'Review syntheses', href: '/syntheses', detail: 'compression' },
      ],
      actions: [
        { label: 'New inquiry', href: '/inquiries/new', tone: 'primary' },
        { label: 'Log engagement', href: '/engagements/new' },
      ],
    }
  }

  if (pathname === '/syntheses/new' || matchPath('/syntheses/:synthesisId/edit', pathname)) {
    return {
      eyebrow: 'Synthesis form',
      title: 'State the current compression',
      body: 'A good pass should say what the inquiry now looks like, what remains unresolved, and why this synthesis should exist at all.',
      links: [
        { label: 'Browse syntheses', href: '/syntheses', detail: 'history' },
        { label: 'Browse inquiries', href: '/inquiries', detail: 'inputs' },
      ],
      actions: [{ label: 'Back to syntheses', href: '/syntheses' }],
    }
  }

  if (pathname === '/syntheses' || matchPath('/syntheses/:synthesisId', pathname)) {
    return {
      eyebrow: 'Synthesis desk',
      title: 'Compression',
      body: 'A synthesis is where the inquiry cashes out for now: current view, unresolved tension, and what the evidence presently supports.',
      links: [
        { label: 'Browse syntheses', href: '/syntheses', detail: 'history' },
        { label: 'Review inquiries', href: '/inquiries', detail: 'source' },
      ],
      actions: [
        { label: 'New synthesis', href: '/syntheses/new', tone: 'primary' },
        { label: 'Open inquiries', href: '/inquiries' },
      ],
    }
  }

  return {
    eyebrow: 'Workspace',
    title: 'Stay in the loop',
    body: 'Capture new material, connect it to a live question, and compress it when the evidence is dense enough.',
    actions: [{ label: 'Dashboard', href: '/', tone: 'primary' }],
  }
}

const primaryActionClassName =
  'rounded-xl bg-pine px-3 py-2 text-center text-sm font-medium text-white transition hover:bg-pine/90'

const secondaryActionClassName =
  'rounded-xl border border-black/10 bg-white/90 px-3 py-2 text-center text-sm text-ink transition hover:bg-white'
