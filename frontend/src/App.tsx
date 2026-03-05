import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from '@/hooks/use-auth';
import { ClusterProvider } from '@/hooks/use-cluster';
import { Layout } from '@/components/Layout';
import { LoginPage } from '@/pages/LoginPage';
import { DashboardPage } from '@/pages/DashboardPage';
import { NodesPage } from '@/pages/NodesPage';
import { ConfigMapsPage } from '@/pages/ConfigMapsPage';
import { DeploymentsPage } from '@/pages/DeploymentsPage';
import { PodsPage } from '@/pages/PodsPage';
import { HPAPage } from '@/pages/HPAPage';
import { HistoryPage } from '@/pages/HistoryPage';
import { UsersPage } from '@/pages/UsersPage';
import { ClustersPage } from '@/pages/ClustersPage';
import { Loading } from '@/components/ui/spinner';

// Protected route wrapper
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loading text="Loading..." />
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          {/* Public routes */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/auth/feishu/callback" element={<LoginPage />} />

          {/* Protected routes */}
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <ClusterProvider>
                  <Layout />
                </ClusterProvider>
              </ProtectedRoute>
            }
          >
            <Route index element={<DashboardPage />} />
            <Route path="nodes" element={<NodesPage />} />
            <Route path="configmaps" element={<ConfigMapsPage />} />
            <Route path="deployments" element={<DeploymentsPage />} />
            <Route path="pods" element={<PodsPage />} />
            <Route path="hpas" element={<HPAPage />} />
            <Route path="history" element={<HistoryPage />} />
            <Route path="users" element={<UsersPage />} />
            <Route path="clusters" element={<ClustersPage />} />
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
