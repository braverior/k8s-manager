import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { Toaster } from '@/components/ui/toaster';
import { useCluster } from '@/hooks/use-cluster';
import { useAuth } from '@/hooks/use-auth';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { ShieldX, LogOut } from 'lucide-react';

function NoAccessPage() {
  const { user, logout } = useAuth();

  const handleLogout = async () => {
    await logout();
    window.location.href = '/login';
  };

  return (
    <div className="flex-1 flex items-center justify-center p-6">
      <Card className="max-w-md w-full">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <div className="w-16 h-16 rounded-full bg-destructive/10 flex items-center justify-center">
              <ShieldX className="w-8 h-8 text-destructive" />
            </div>
          </div>
          <CardTitle>No Access</CardTitle>
          <CardDescription>
            You don't have permission to access any clusters
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="p-4 rounded-lg bg-muted text-sm">
            <p className="font-medium mb-2">Current user:</p>
            <div className="space-y-1 text-muted-foreground">
              <p>Name: {user?.name}</p>
              <p>Email: {user?.email}</p>
              {user?.department?.name && (
                <p>Department: {user.department.name}</p>
              )}
            </div>
          </div>
          <p className="text-sm text-muted-foreground text-center">
            Please contact your administrator to request access to clusters and namespaces.
          </p>
          <Button variant="outline" className="w-full" onClick={handleLogout}>
            <LogOut className="w-4 h-4 mr-2" />
            Sign out and try another account
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

export function Layout() {
  const { clusters, loading } = useCluster();
  const { isAdmin } = useAuth();

  // Show no access page if user has no cluster permissions (and is not admin)
  const hasNoAccess = !loading && clusters.length === 0 && !isAdmin;

  return (
    <div className="flex min-h-screen bg-background">
      <Sidebar />
      <main className="flex-1 overflow-auto">
        {hasNoAccess ? <NoAccessPage /> : <Outlet />}
      </main>
      <Toaster />
    </div>
  );
}
