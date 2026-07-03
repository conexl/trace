import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { TooltipProvider } from '@/components/ui/Tooltip';
import { AuthProvider } from '@/lib/auth';
import { ToastProvider } from '@/components/ToastProvider';
import { Layout } from '@/components/Layout';
import { LandingPage } from '@/pages/LandingPage';
import { ServersPage } from '@/pages/ServersPage';
import { ServerDetailPage } from '@/pages/ServerDetailPage';
import { LoginPage } from '@/pages/LoginPage';
import { RegisterPage } from '@/pages/RegisterPage';
import { NotFoundPage } from '@/pages/NotFoundPage';

function App() {
  return (
    <AuthProvider>
      <TooltipProvider delayDuration={200}>
        <ToastProvider>
          <BrowserRouter>
            <Routes>
              <Route element={<Layout />}>
                <Route path="/" element={<LandingPage />} />
                <Route path="/servers" element={<ServersPage />} />
                <Route path="/servers/:id" element={<ServerDetailPage />} />
              </Route>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
              <Route path="*" element={<NotFoundPage />} />
            </Routes>
          </BrowserRouter>
        </ToastProvider>
      </TooltipProvider>
    </AuthProvider>
  );
}

export default App;
