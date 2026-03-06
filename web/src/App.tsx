import { Navigate, Outlet, Route, Routes } from 'react-router-dom'
import { AppShell } from './components/shared/AppShell'
import { SessionProvider, useSession } from './hooks/useSession'
import { EngagementDetailPage } from './pages/engagements/EngagementDetailPage'
import { EngagementFormPage } from './pages/engagements/EngagementFormPage'
import { DashboardPage } from './pages/DashboardPage'
import { InquiryDetailPage } from './pages/inquiries/InquiryDetailPage'
import { InquiryFormPage } from './pages/inquiries/InquiryFormPage'
import { InquiriesPage } from './pages/inquiries/InquiriesPage'
import { LoginPage } from './pages/LoginPage'
import { PlaceholderPage } from './pages/PlaceholderPage'
import { SourceDetailPage } from './pages/sources/SourceDetailPage'
import { SourceFormPage } from './pages/sources/SourceFormPage'
import { SourcesPage } from './pages/sources/SourcesPage'

function ProtectedLayout() {
  const { loading, session } = useSession()

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-canvas px-6 text-ink">
        <div className="rounded-3xl border border-black/5 bg-white/80 px-6 py-5 shadow-card">
          Loading Lectio...
        </div>
      </div>
    )
  }

  if (!session.authenticated) {
    return <Navigate to="/login" replace />
  }

  return (
    <AppShell>
      <Outlet />
    </AppShell>
  )
}

function AppRoutes() {
  const { session } = useSession()

  return (
    <Routes>
      <Route path="/login" element={session.authenticated ? <Navigate to="/" replace /> : <LoginPage />} />
      <Route path="/" element={<ProtectedLayout />}>
        <Route index element={<DashboardPage />} />
        <Route path="sources" element={<SourcesPage />} />
        <Route path="sources/new" element={<SourceFormPage mode="create" />} />
        <Route path="sources/:sourceId" element={<SourceDetailPage />} />
        <Route path="sources/:sourceId/edit" element={<SourceFormPage mode="edit" />} />
        <Route path="engagements/new" element={<EngagementFormPage mode="create" />} />
        <Route path="engagements/:engagementId" element={<EngagementDetailPage />} />
        <Route path="engagements/:engagementId/edit" element={<EngagementFormPage mode="edit" />} />
        <Route path="inquiries" element={<InquiriesPage />} />
        <Route path="inquiries/new" element={<InquiryFormPage mode="create" />} />
        <Route path="inquiries/:inquiryId" element={<InquiryDetailPage />} />
        <Route path="inquiries/:inquiryId/edit" element={<InquiryFormPage mode="edit" />} />
        <Route
          path="syntheses"
          element={<PlaceholderPage title="Syntheses" body="Synthesis remains in MVP, but not before the core capture loop exists." />}
        />
      </Route>
      <Route path="*" element={<Navigate to={session.authenticated ? '/' : '/login'} replace />} />
    </Routes>
  )
}

export default function App() {
  return (
    <SessionProvider>
      <AppRoutes />
    </SessionProvider>
  )
}
