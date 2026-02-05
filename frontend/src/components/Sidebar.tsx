import { Link, useLocation } from 'react-router-dom';
import { cn } from '@/lib/utils';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Button } from '@/components/ui/button';
import { useCluster } from '@/hooks/use-cluster';
import { useAuth } from '@/hooks/use-auth';
import {
  Box,
  Layers,
  Settings,
  Server,
  History,
  LayoutDashboard,
  Container,
  Users,
  LogOut,
  ShieldCheck,
  Gauge,
  Mail,
  Building2,
  Hash,
  Phone,
} from 'lucide-react';

const navItems = [
  { path: '/', label: 'Dashboard', icon: LayoutDashboard },
  { path: '/nodes', label: 'Nodes', icon: Server },
  { path: '/configmaps', label: 'ConfigMaps', icon: Settings },
  { path: '/deployments', label: 'Deployments', icon: Box },
  { path: '/pods', label: 'Pods', icon: Container },
  { path: '/hpas', label: 'HPAs', icon: Gauge },
  { path: '/history', label: 'History', icon: History },
];

const adminNavItems = [
  { path: '/users', label: 'Users', icon: Users },
];

export function Sidebar() {
  const location = useLocation();
  const { user, isAdmin, logout } = useAuth();
  const {
    clusters,
    namespaces,
    selectedCluster,
    selectedNamespace,
    setSelectedCluster,
    setSelectedNamespace,
  } = useCluster();

  const handleLogout = async () => {
    await logout();
    window.location.href = '/login';
  };

  return (
    <aside className="w-64 min-h-screen bg-card border-r border-border flex flex-col">
      {/* Logo */}
      <div className="p-4 flex items-center gap-3">
        <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
          <Layers className="w-6 h-6 text-primary" />
        </div>
        <div>
          <h1 className="font-semibold text-foreground">K8S Manager</h1>
          <p className="text-xs text-muted-foreground">Kubernetes Console</p>
        </div>
      </div>

      <Separator />

      {/* Cluster & Namespace Selectors */}
      <div className="p-4 space-y-3">
        {clusters.length === 0 ? (
          <div className="p-3 rounded-md bg-muted/50 text-center">
            <p className="text-xs text-muted-foreground">No clusters available</p>
          </div>
        ) : (
          <>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-muted-foreground flex items-center gap-1.5">
                <Server className="w-3.5 h-3.5" />
                Cluster
              </label>
              <Select value={selectedCluster} onValueChange={setSelectedCluster}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Select cluster" />
                </SelectTrigger>
                <SelectContent>
                  {clusters.map((cluster) => (
                    <SelectItem key={cluster.name} value={cluster.name}>
                      <span className="truncate">{cluster.name}</span>
                      <Badge
                        variant={cluster.status === 'connected' ? 'success' : 'destructive'}
                        className="text-[10px] px-1.5 py-0 shrink-0"
                      >
                        {cluster.status}
                      </Badge>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-medium text-muted-foreground flex items-center gap-1.5">
                <Box className="w-3.5 h-3.5" />
                Namespace
              </label>
              <Select
                value={selectedNamespace}
                onValueChange={setSelectedNamespace}
                disabled={namespaces.length === 0}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder={namespaces.length === 0 ? "No namespaces" : "Select namespace"} />
                </SelectTrigger>
                <SelectContent>
                  {namespaces.map((ns) => (
                    <SelectItem key={ns.name} value={ns.name}>
                      {ns.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </>
        )}
      </div>

      <Separator />

      {/* Navigation */}
      <nav className="flex-1 p-3">
        <ul className="space-y-1">
          {navItems.map((item) => {
            const isActive = location.pathname === item.path;
            const Icon = item.icon;
            return (
              <li key={item.path}>
                <Link
                  to={item.path}
                  className={cn(
                    'flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition-colors cursor-pointer',
                    isActive
                      ? 'bg-primary/10 text-primary'
                      : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                  )}
                >
                  <Icon className="w-4 h-4" />
                  {item.label}
                </Link>
              </li>
            );
          })}
        </ul>

        {/* Admin Section */}
        {isAdmin && (
          <>
            <Separator className="my-3" />
            <p className="px-3 mb-2 text-xs font-medium text-muted-foreground flex items-center gap-1.5">
              <ShieldCheck className="w-3.5 h-3.5" />
              Admin
            </p>
            <ul className="space-y-1">
              {adminNavItems.map((item) => {
                const isActive = location.pathname === item.path;
                const Icon = item.icon;
                return (
                  <li key={item.path}>
                    <Link
                      to={item.path}
                      className={cn(
                        'flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition-colors cursor-pointer',
                        isActive
                          ? 'bg-primary/10 text-primary'
                          : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                      )}
                    >
                      <Icon className="w-4 h-4" />
                      {item.label}
                    </Link>
                  </li>
                );
              })}
            </ul>
          </>
        )}
      </nav>

      {/* User Profile & Footer */}
      <div className="p-3 border-t border-border">
        {user && (
          <Popover>
            <div className="flex items-center gap-2">
              <PopoverTrigger asChild>
                <div className="flex items-center gap-3 flex-1 min-w-0 cursor-pointer hover:bg-muted/50 rounded-md p-1.5 -m-1.5 transition-colors">
                  {user.avatar_url ? (
                    <img
                      src={user.avatar_url}
                      alt={user.name}
                      className="w-8 h-8 rounded-full ring-2 ring-border"
                    />
                  ) : (
                    <div className="w-8 h-8 rounded-full bg-gradient-to-br from-primary/80 to-primary flex items-center justify-center ring-2 ring-border">
                      <span className="text-xs font-medium text-primary-foreground">{user.name[0]}</span>
                    </div>
                  )}
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium truncate">{user.name}</p>
                    <div className="flex items-center gap-2">
                      <span className="text-[10px] text-muted-foreground">
                        {user.is_admin ? 'Admin' : 'User'}
                      </span>
                      <span className="text-[10px] text-muted-foreground/50">·</span>
                      <span className="text-[10px] text-muted-foreground/50">v1.0.0</span>
                    </div>
                  </div>
                </div>
              </PopoverTrigger>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 shrink-0 text-muted-foreground hover:text-destructive"
                onClick={handleLogout}
                title="Logout"
              >
                <LogOut className="w-3.5 h-3.5" />
              </Button>
            </div>
            <PopoverContent className="w-72 p-0" side="top" align="start">
              {/* User Header */}
              <div className="p-4 bg-gradient-to-br from-primary/10 to-primary/5">
                <div className="flex items-center gap-3">
                  {user.avatar_url ? (
                    <img
                      src={user.avatar_url}
                      alt={user.name}
                      className="w-12 h-12 rounded-full ring-2 ring-background shadow-md"
                    />
                  ) : (
                    <div className="w-12 h-12 rounded-full bg-gradient-to-br from-primary to-primary/80 flex items-center justify-center ring-2 ring-background shadow-md">
                      <span className="text-lg font-semibold text-primary-foreground">{user.name[0]}</span>
                    </div>
                  )}
                  <div className="flex-1 min-w-0">
                    <p className="font-semibold truncate">{user.name}</p>
                    <Badge variant={user.is_admin ? 'default' : 'secondary'} className="mt-1 text-[10px]">
                      {user.is_admin ? (
                        <><ShieldCheck className="w-3 h-3 mr-1" />Admin</>
                      ) : (
                        'User'
                      )}
                    </Badge>
                  </div>
                </div>
              </div>

              {/* User Details */}
              <div className="p-3 space-y-2.5">
                {user.employee_id && (
                  <div className="flex items-center gap-2.5 text-sm">
                    <Hash className="w-4 h-4 text-muted-foreground shrink-0" />
                    <span className="text-muted-foreground">{user.employee_id}</span>
                  </div>
                )}
                {user.department?.name && (
                  <div className="flex items-center gap-2.5 text-sm">
                    <Building2 className="w-4 h-4 text-muted-foreground shrink-0" />
                    <span className="text-muted-foreground truncate">{user.department.name}</span>
                  </div>
                )}
                {user.email && (
                  <div className="flex items-center gap-2.5 text-sm">
                    <Mail className="w-4 h-4 text-muted-foreground shrink-0" />
                    <span className="text-muted-foreground truncate">{user.email}</span>
                  </div>
                )}
                {user.mobile && (
                  <div className="flex items-center gap-2.5 text-sm">
                    <Phone className="w-4 h-4 text-muted-foreground shrink-0" />
                    <span className="text-muted-foreground">{user.mobile}</span>
                  </div>
                )}
              </div>

              <Separator />

              {/* Actions */}
              <div className="p-2">
                <Button
                  variant="ghost"
                  className="w-full justify-start text-destructive hover:text-destructive hover:bg-destructive/10"
                  onClick={handleLogout}
                >
                  <LogOut className="w-4 h-4 mr-2" />
                  Logout
                </Button>
              </div>

              {/* Version */}
              <div className="px-3 pb-2 text-center">
                <span className="text-[10px] text-muted-foreground/50">K8S Manager v1.0.0</span>
              </div>
            </PopoverContent>
          </Popover>
        )}
      </div>
    </aside>
  );
}
