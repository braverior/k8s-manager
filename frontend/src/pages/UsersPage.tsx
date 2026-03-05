import { useState, useEffect, useCallback } from 'react';
import { adminApi, clusterApi, fetchApi } from '@/api';
import { useAuth } from '@/hooks/use-auth';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Loading } from '@/components/ui/spinner';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import type { User, Cluster, Namespace, ClusterPermission } from '@/types';
import {
  Search,
  RefreshCw,
  Users,
  Shield,
  ShieldCheck,
  UserX,
  UserCheck,
  Settings,
} from 'lucide-react';

export function UsersPage() {
  const { isAdmin } = useAuth();
  const { toast } = useToast();

  const [users, setUsers] = useState<User[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState<string>('all');
  const [page, setPage] = useState(1);
  const pageSize = 20;

  // Permission dialog
  const [permDialogOpen, setPermDialogOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [clusters, setClusters] = useState<Cluster[]>([]);
  const [clusterNamespaces, setClusterNamespaces] = useState<Record<string, Namespace[]>>({});
  const [editingPermissions, setEditingPermissions] = useState<ClusterPermission[]>([]);
  const [savingPermissions, setSavingPermissions] = useState(false);

  const fetchUsers = useCallback(async () => {
    try {
      setLoading(true);
      const params: { keyword?: string; role?: string; page: number; page_size: number } = {
        page,
        page_size: pageSize,
      };
      if (searchTerm) params.keyword = searchTerm;
      if (roleFilter !== 'all') params.role = roleFilter;

      const data = await adminApi.listUsers('', params);
      setUsers(data.items || []);
      setTotal(data.total);
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch users',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [page, searchTerm, roleFilter, toast]);

  useEffect(() => {
    // Fetch clusters from API for permissions dialog
    const loadClusters = async () => {
      try {
        const data = await fetchApi<Cluster[]>('', '/clusters');
        setClusters(data || []);
      } catch {
        setClusters([]);
      }
    };
    loadClusters();
  }, []);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  const handleToggleRole = async (user: User) => {
    const newRole = user.is_admin ? 'user' : 'admin';
    try {
      await adminApi.updateUserRole('', user.id, newRole);
      toast({
        title: 'Success',
        description: `User role updated to ${newRole}`,
      });
      fetchUsers();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to update role',
        variant: 'destructive',
      });
    }
  };

  const handleToggleStatus = async (user: User) => {
    const newStatus = user.status === 'active' ? 'disabled' : 'active';
    try {
      await adminApi.updateUserStatus('', user.id, newStatus);
      toast({
        title: 'Success',
        description: `User ${newStatus === 'active' ? 'enabled' : 'disabled'}`,
      });
      fetchUsers();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to update status',
        variant: 'destructive',
      });
    }
  };

  const openPermissionDialog = async (user: User) => {
    setSelectedUser(user);
    setEditingPermissions(user.permissions || []);
    setPermDialogOpen(true);

    // Fetch namespaces for each cluster
    const nsMap: Record<string, Namespace[]> = {};
    for (const cluster of clusters) {
      try {
        const namespaces = await clusterApi.getNamespaces(cluster.name);
        nsMap[cluster.name] = namespaces || [];
      } catch {
        nsMap[cluster.name] = [];
      }
    }
    setClusterNamespaces(nsMap);
  };

  const handleClusterToggle = (clusterName: string, checked: boolean) => {
    if (checked) {
      // Add cluster with all namespaces
      setEditingPermissions([
        ...editingPermissions,
        { cluster: clusterName, namespaces: ['*'] },
      ]);
    } else {
      // Remove cluster
      setEditingPermissions(editingPermissions.filter((p) => p.cluster !== clusterName));
    }
  };

  const handleNamespaceToggle = (clusterName: string, namespace: string, checked: boolean) => {
    const clusterPerm = editingPermissions.find((p) => p.cluster === clusterName);

    if (!clusterPerm) {
      if (checked) {
        setEditingPermissions([
          ...editingPermissions,
          { cluster: clusterName, namespaces: [namespace] },
        ]);
      }
      return;
    }

    let newNamespaces: string[];
    if (namespace === '*') {
      newNamespaces = checked ? ['*'] : [];
    } else {
      if (clusterPerm.namespaces.includes('*')) {
        // Switching from all to specific
        const allNs = clusterNamespaces[clusterName]?.map((n) => n.name) || [];
        newNamespaces = checked
          ? allNs
          : allNs.filter((n) => n !== namespace);
      } else {
        newNamespaces = checked
          ? [...clusterPerm.namespaces.filter((n) => n !== '*'), namespace]
          : clusterPerm.namespaces.filter((n) => n !== namespace);
      }
    }

    if (newNamespaces.length === 0) {
      setEditingPermissions(editingPermissions.filter((p) => p.cluster !== clusterName));
    } else {
      setEditingPermissions(
        editingPermissions.map((p) =>
          p.cluster === clusterName ? { ...p, namespaces: newNamespaces } : p
        )
      );
    }
  };

  const savePermissions = async () => {
    if (!selectedUser) return;
    try {
      setSavingPermissions(true);
      await adminApi.setUserPermissions('', selectedUser.id, editingPermissions);
      toast({ title: 'Success', description: 'Permissions updated' });
      setPermDialogOpen(false);
      fetchUsers();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to save permissions',
        variant: 'destructive',
      });
    } finally {
      setSavingPermissions(false);
    }
  };

  if (!isAdmin) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-muted-foreground">Access denied. Admin privileges required.</p>
      </div>
    );
  }

  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold flex items-center gap-2">
            <Users className="w-7 h-7 text-primary" />
            User Management
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage users and permissions
          </p>
        </div>
        <Badge variant="outline" className="text-sm">
          {total} users
        </Badge>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search users..."
            value={searchTerm}
            onChange={(e) => {
              setSearchTerm(e.target.value);
              setPage(1);
            }}
            className="pl-9"
          />
        </div>
        <Select value={roleFilter} onValueChange={(v) => { setRoleFilter(v); setPage(1); }}>
          <SelectTrigger className="w-[150px]">
            <SelectValue placeholder="Filter by role" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Roles</SelectItem>
            <SelectItem value="admin">Admin</SelectItem>
            <SelectItem value="user">User</SelectItem>
          </SelectContent>
        </Select>
        <Button variant="outline" size="icon" onClick={fetchUsers}>
          <RefreshCw className="w-4 h-4" />
        </Button>
      </div>

      {/* User List */}
      {loading ? (
        <Loading text="Loading users..." />
      ) : users.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Users className="w-12 h-12 text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">No users found</p>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left p-4 font-medium text-muted-foreground">User</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Department</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Role</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Status</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Last Login</th>
                  <th className="text-right p-4 font-medium text-muted-foreground">Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user) => (
                  <tr key={user.id} className="border-b border-border hover:bg-muted/50">
                    <td className="p-4">
                      <div className="flex items-center gap-3">
                        {user.avatar_url ? (
                          <img
                            src={user.avatar_url}
                            alt={user.name}
                            className="w-10 h-10 rounded-full"
                          />
                        ) : (
                          <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                            <span className="text-sm font-medium">{user.name[0]}</span>
                          </div>
                        )}
                        <div>
                          <p className="font-medium">{user.name}</p>
                          <p className="text-sm text-muted-foreground">{user.email}</p>
                        </div>
                      </div>
                    </td>
                    <td className="p-4">
                      <span className="text-sm">{user.department?.name || '-'}</span>
                    </td>
                    <td className="p-4">
                      <Badge variant={user.is_admin ? 'default' : 'secondary'}>
                        {user.is_admin ? (
                          <><ShieldCheck className="w-3 h-3 mr-1" /> Admin</>
                        ) : (
                          <><Shield className="w-3 h-3 mr-1" /> User</>
                        )}
                      </Badge>
                    </td>
                    <td className="p-4">
                      <Badge variant={user.status === 'active' ? 'success' : 'destructive'}>
                        {user.status === 'active' ? 'Active' : 'Disabled'}
                      </Badge>
                    </td>
                    <td className="p-4 text-sm text-muted-foreground">
                      {user.last_login_at
                        ? new Date(user.last_login_at).toLocaleString()
                        : '-'}
                    </td>
                    <td className="p-4">
                      <div className="flex items-center justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openPermissionDialog(user)}
                          disabled={user.is_admin}
                          title={user.is_admin ? 'Admins have all permissions' : 'Manage permissions'}
                        >
                          <Settings className="w-4 h-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleToggleRole(user)}
                          title={user.is_admin ? 'Remove admin' : 'Make admin'}
                        >
                          {user.is_admin ? (
                            <Shield className="w-4 h-4" />
                          ) : (
                            <ShieldCheck className="w-4 h-4" />
                          )}
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleToggleStatus(user)}
                          className={user.status === 'active' ? 'text-destructive hover:text-destructive' : 'text-green-500 hover:text-green-500'}
                          title={user.status === 'active' ? 'Disable user' : 'Enable user'}
                        >
                          {user.status === 'active' ? (
                            <UserX className="w-4 h-4" />
                          ) : (
                            <UserCheck className="w-4 h-4" />
                          )}
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between p-4 border-t border-border">
              <p className="text-sm text-muted-foreground">
                Page {page} of {totalPages}
              </p>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage(page - 1)}
                  disabled={page === 1}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage(page + 1)}
                  disabled={page === totalPages}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </Card>
      )}

      {/* Permission Dialog */}
      <Dialog open={permDialogOpen} onOpenChange={setPermDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Manage Permissions</DialogTitle>
            <DialogDescription>
              Configure cluster and namespace access for {selectedUser?.name}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {clusters.map((cluster) => {
              const clusterPerm = editingPermissions.find((p) => p.cluster === cluster.name);
              const hasCluster = !!clusterPerm;
              const hasAllNs = clusterPerm?.namespaces.includes('*');
              const namespaces = clusterNamespaces[cluster.name] || [];

              return (
                <Card key={cluster.name}>
                  <CardHeader className="pb-2">
                    <div className="flex items-center justify-between">
                      <CardTitle className="text-sm flex items-center gap-2">
                        <Checkbox
                          checked={hasCluster}
                          onCheckedChange={(checked) =>
                            handleClusterToggle(cluster.name, checked as boolean)
                          }
                        />
                        {cluster.name}
                        <Badge variant="outline" className="text-xs">
                          {cluster.status}
                        </Badge>
                      </CardTitle>
                      {hasCluster && (
                        <Badge variant="secondary" className="text-xs">
                          {hasAllNs
                            ? 'All Namespaces'
                            : `${clusterPerm?.namespaces.length} namespace(s)`}
                        </Badge>
                      )}
                    </div>
                  </CardHeader>
                  {hasCluster && (
                    <CardContent className="pt-0">
                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <Checkbox
                            checked={hasAllNs}
                            onCheckedChange={(checked) =>
                              handleNamespaceToggle(cluster.name, '*', checked as boolean)
                            }
                          />
                          <span className="text-sm font-medium">All Namespaces (*)</span>
                        </div>
                        {!hasAllNs && (
                          <div className="grid grid-cols-3 gap-2 pl-6">
                            {namespaces.map((ns) => (
                              <div key={ns.name} className="flex items-center gap-2">
                                <Checkbox
                                  checked={clusterPerm?.namespaces.includes(ns.name)}
                                  onCheckedChange={(checked) =>
                                    handleNamespaceToggle(cluster.name, ns.name, checked as boolean)
                                  }
                                />
                                <span className="text-sm">{ns.name}</span>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    </CardContent>
                  )}
                </Card>
              );
            })}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setPermDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={savePermissions} disabled={savingPermissions}>
              {savingPermissions ? 'Saving...' : 'Save Permissions'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
