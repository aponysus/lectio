import { Suspense, lazy } from 'react'
import { Navigate, Outlet, Route, Routes } from 'react-router-dom'
import { ConfirmProvider } from './components/feedback/ConfirmProvider'
import { ToastProvider } from './components/feedback/ToastProvider'
import { AppShell } from './components/shared/AppShell'
import { LoadingPanel } from './components/shared/LoadingPanel'
import { SessionProvider, useSession } from './hooks/useSession'

const DashboardPage = lazy(async () => {
  const module = await import('./pages/DashboardPage')
  return { default: module.DashboardPage }
})

const LoginPage = lazy(async () => {
  const module = await import('./pages/LoginPage')
  return { default: module.LoginPage }
})

const SearchPage = lazy(async () => {
  const module = await import('./pages/SearchPage')
  return { default: module.SearchPage }
})

const ExportPage = lazy(async () => {
  const module = await import('./pages/settings/ExportPage')
  return { default: module.ExportPage }
})

const SourcesPage = lazy(async () => {
  const module = await import('./pages/sources/SourcesPage')
  return { default: module.SourcesPage }
})

const SourceFormPage = lazy(async () => {
  const module = await import('./pages/sources/SourceFormPage')
  return { default: module.SourceFormPage }
})

const SourceDetailPage = lazy(async () => {
  const module = await import('./pages/sources/SourceDetailPage')
  return { default: module.SourceDetailPage }
})

const EngagementsPage = lazy(async () => {
  const module = await import('./pages/engagements/EngagementsPage')
  return { default: module.EngagementsPage }
})

const EngagementFormPage = lazy(async () => {
  const module = await import('./pages/engagements/EngagementFormPage')
  return { default: module.EngagementFormPage }
})

const EngagementDetailPage = lazy(async () => {
  const module = await import('./pages/engagements/EngagementDetailPage')
  return { default: module.EngagementDetailPage }
})

const InquiriesPage = lazy(async () => {
  const module = await import('./pages/inquiries/InquiriesPage')
  return { default: module.InquiriesPage }
})

const InquiryFormPage = lazy(async () => {
  const module = await import('./pages/inquiries/InquiryFormPage')
  return { default: module.InquiryFormPage }
})

const InquiryDetailPage = lazy(async () => {
  const module = await import('./pages/inquiries/InquiryDetailPage')
  return { default: module.InquiryDetailPage }
})

const SynthesesPage = lazy(async () => {
  const module = await import('./pages/syntheses/SynthesesPage')
  return { default: module.SynthesesPage }
})

const SynthesisFormPage = lazy(async () => {
  const module = await import('./pages/syntheses/SynthesisFormPage')
  return { default: module.SynthesisFormPage }
})

const SynthesisDetailPage = lazy(async () => {
  const module = await import('./pages/syntheses/SynthesisDetailPage')
  return { default: module.SynthesisDetailPage }
})

function FullScreenLoading() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-canvas px-6 text-ink">
      <LoadingPanel label="Loading Lectio" variant="shell" />
    </div>
  )
}

function ProtectedLayout() {
  const { loading, session } = useSession()

  if (loading) {
    return <FullScreenLoading />
  }

  if (!session.authenticated) {
    return <Navigate to="/login" replace />
  }

  return (
    <AppShell>
      <Suspense fallback={<LoadingPanel label="Loading page" />}>
        <Outlet />
      </Suspense>
    </AppShell>
  )
}

function LoginRoute() {
  return (
    <Suspense fallback={<FullScreenLoading />}>
      <LoginPage />
    </Suspense>
  )
}

function AppRoutes() {
  const { session } = useSession()

  return (
    <Routes>
      <Route path="/login" element={session.authenticated ? <Navigate to="/" replace /> : <LoginRoute />} />
      <Route path="/" element={<ProtectedLayout />}>
        <Route index element={<DashboardPage />} />
        <Route path="search" element={<SearchPage />} />
        <Route path="settings/export" element={<ExportPage />} />
        <Route path="sources" element={<SourcesPage />} />
        <Route path="sources/new" element={<SourceFormPage mode="create" />} />
        <Route path="sources/:sourceId" element={<SourceDetailPage />} />
        <Route path="sources/:sourceId/edit" element={<SourceFormPage mode="edit" />} />
        <Route path="engagements" element={<EngagementsPage />} />
        <Route path="engagements/new" element={<EngagementFormPage mode="create" />} />
        <Route path="engagements/:engagementId" element={<EngagementDetailPage />} />
        <Route path="engagements/:engagementId/edit" element={<EngagementFormPage mode="edit" />} />
        <Route path="inquiries" element={<InquiriesPage />} />
        <Route path="inquiries/new" element={<InquiryFormPage mode="create" />} />
        <Route path="inquiries/:inquiryId" element={<InquiryDetailPage />} />
        <Route path="inquiries/:inquiryId/edit" element={<InquiryFormPage mode="edit" />} />
        <Route path="syntheses" element={<SynthesesPage />} />
        <Route path="syntheses/new" element={<SynthesisFormPage mode="create" />} />
        <Route path="syntheses/:synthesisId" element={<SynthesisDetailPage />} />
        <Route path="syntheses/:synthesisId/edit" element={<SynthesisFormPage mode="edit" />} />
      </Route>
      <Route path="*" element={<Navigate to={session.authenticated ? '/' : '/login'} replace />} />
    </Routes>
  )
}

export default function App() {
  return (
    <SessionProvider>
      <ToastProvider>
        <ConfirmProvider>
          <AppRoutes />
        </ConfirmProvider>
      </ToastProvider>
    </SessionProvider>
  )
}
