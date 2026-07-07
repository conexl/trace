import * as React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { TooltipProvider } from '@/components/ui/Tooltip';
import { AuthProvider, useAuth } from '@/lib/auth';
import { ToastProvider } from '@/components/ToastProvider';
import { Layout } from '@/components/Layout';
import { LandingPage } from '@/pages/LandingPage';
import { LoginPage } from '@/pages/LoginPage';
import { RegisterPage } from '@/pages/RegisterPage';
import { NotFoundPage } from '@/pages/NotFoundPage';

const ServersPage = React.lazy(() => import('@/pages/ServersPage').then((m) => ({ default: m.ServersPage })));
const ServerDetailPage = React.lazy(() => import('@/pages/ServerDetailPage').then((m) => ({ default: m.ServerDetailPage })));
const AlertsPage = React.lazy(() => import('@/pages/AlertsPage').then((m) => ({ default: m.AlertsPage })));
const TasksPage = React.lazy(() => import('@/pages/TasksPage').then((m) => ({ default: m.TasksPage })));
const IncidentsPage = React.lazy(() => import('@/pages/IncidentsPage').then((m) => ({ default: m.IncidentsPage })));

function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  );
}

function AppContent() {
  const { loading } = useAuth();

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-canvas text-sm text-muted">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-border border-t-accent" />
      </div>
    );
  }

  return (
    <TooltipProvider delayDuration={200}>
      <ToastProvider>
        <BrowserRouter>
          <React.Suspense
            fallback={
              <div className="flex h-screen items-center justify-center bg-canvas text-sm text-muted">
                <div className="h-6 w-6 animate-spin rounded-full border-2 border-border border-t-accent" />
              </div>
            }
          >
            <Routes>
              <Route element={<Layout />}>
                <Route path="/" element={<LandingPage />} />
                <Route path="/servers" element={<ServersPage />} />
                <Route path="/servers/:id" element={<ServerDetailPage />} />
                <Route path="/alerts" element={<AlertsPage />} />
                <Route path="/incidents" element={<IncidentsPage />} />
                <Route path="/tasks" element={<TasksPage />} />
              </Route>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
              <Route path="*" element={<NotFoundPage />} />
            </Routes>
          </React.Suspense>
        </BrowserRouter>
      </ToastProvider>
    </TooltipProvider>
  );
}

export default App;
