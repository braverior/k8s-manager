import { useState, useEffect, useCallback, useRef } from 'react';
import { clusterManageApi } from '@/api';
import { useAuth } from '@/hooks/use-auth';
import { useCluster } from '@/hooks/use-cluster';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Loading } from '@/components/ui/spinner';
import { Textarea } from '@/components/ui/textarea';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import type { ClusterDetail } from '@/types';
import {
  RefreshCw,
  Server,
  Plus,
  Pencil,
  Trash2,
  TestTube,
  Upload,
} from 'lucide-react';

export function ClustersPage() {
  const { isAdmin } = useAuth();
  const { refreshClusters } = useCluster();
  const { toast } = useToast();

  const [clusters, setClusters] = useState<ClusterDetail[]>([]);
  const [loading, setLoading] = useState(true);

  // Add dialog
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [addName, setAddName] = useState('');
  const [addDescription, setAddDescription] = useState('');
  const [addKubeconfig, setAddKubeconfig] = useState('');
  const [adding, setAdding] = useState(false);
  const [testingNew, setTestingNew] = useState(false);
  const [testResult, setTestResult] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Edit dialog
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [editCluster, setEditCluster] = useState<ClusterDetail | null>(null);
  const [editDescription, setEditDescription] = useState('');
  const [editKubeconfig, setEditKubeconfig] = useState('');
  const [editing, setEditing] = useState(false);
  const editFileInputRef = useRef<HTMLInputElement>(null);

  // Delete dialog
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleteCluster, setDeleteCluster] = useState<ClusterDetail | null>(null);
  const [deleting, setDeleting] = useState(false);

  const fetchClusters = useCallback(async () => {
    try {
      setLoading(true);
      const data = await clusterManageApi.list();
      setClusters(data || []);
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch clusters',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    fetchClusters();
  }, [fetchClusters]);

  const handleTestConnection = async () => {
    if (!addKubeconfig.trim()) {
      toast({ title: 'Error', description: 'Please provide a kubeconfig', variant: 'destructive' });
      return;
    }
    setTestingNew(true);
    setTestResult(null);
    try {
      const result = await clusterManageApi.testConnection({ kubeconfig: addKubeconfig });
      if (result.success) {
        setTestResult(`Connection successful! Kubernetes ${result.version}`);
      } else {
        setTestResult(`Connection failed: ${result.message}`);
      }
    } catch (err) {
      setTestResult(`Error: ${err instanceof Error ? err.message : 'Test failed'}`);
    } finally {
      setTestingNew(false);
    }
  };

  const handleAdd = async () => {
    if (!addName.trim() || !addKubeconfig.trim()) {
      toast({ title: 'Error', description: 'Name and kubeconfig are required', variant: 'destructive' });
      return;
    }
    setAdding(true);
    try {
      await clusterManageApi.add({
        name: addName.trim(),
        description: addDescription.trim(),
        kubeconfig: addKubeconfig,
      });
      toast({ title: 'Success', description: `Cluster "${addName}" added successfully` });
      setAddDialogOpen(false);
      resetAddForm();
      fetchClusters();
      refreshClusters();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to add cluster',
        variant: 'destructive',
      });
    } finally {
      setAdding(false);
    }
  };

  const handleEdit = async () => {
    if (!editCluster) return;
    setEditing(true);
    try {
      const req: { description?: string; kubeconfig?: string } = {};
      if (editDescription !== editCluster.description) {
        req.description = editDescription;
      }
      if (editKubeconfig.trim()) {
        req.kubeconfig = editKubeconfig;
      }
      await clusterManageApi.update(editCluster.name, req);
      toast({ title: 'Success', description: `Cluster "${editCluster.name}" updated` });
      setEditDialogOpen(false);
      fetchClusters();
      refreshClusters();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to update cluster',
        variant: 'destructive',
      });
    } finally {
      setEditing(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteCluster) return;
    setDeleting(true);
    try {
      await clusterManageApi.delete(deleteCluster.name);
      toast({ title: 'Success', description: `Cluster "${deleteCluster.name}" deleted` });
      setDeleteDialogOpen(false);
      fetchClusters();
      refreshClusters();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to delete cluster',
        variant: 'destructive',
      });
    } finally {
      setDeleting(false);
    }
  };

  const resetAddForm = () => {
    setAddName('');
    setAddDescription('');
    setAddKubeconfig('');
    setTestResult(null);
  };

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>, setter: (v: string) => void) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (ev) => {
      setter(ev.target?.result as string);
    };
    reader.readAsText(file);
    e.target.value = '';
  };

  const openEditDialog = (cluster: ClusterDetail) => {
    setEditCluster(cluster);
    setEditDescription(cluster.description);
    setEditKubeconfig('');
    setEditDialogOpen(true);
  };

  const openDeleteDialog = (cluster: ClusterDetail) => {
    setDeleteCluster(cluster);
    setDeleteDialogOpen(true);
  };

  if (!isAdmin) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-muted-foreground">Access denied. Admin privileges required.</p>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold flex items-center gap-2">
            <Server className="w-7 h-7 text-primary" />
            Cluster Management
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage Kubernetes clusters
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="icon" onClick={fetchClusters}>
            <RefreshCw className="w-4 h-4" />
          </Button>
          <Button onClick={() => { resetAddForm(); setAddDialogOpen(true); }}>
            <Plus className="w-4 h-4 mr-2" />
            Add Cluster
          </Button>
        </div>
      </div>

      {/* Cluster List */}
      {loading ? (
        <Loading text="Loading clusters..." />
      ) : clusters.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Server className="w-12 h-12 text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">No clusters configured</p>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left p-4 font-medium text-muted-foreground">Name</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Description</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">API Server</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Status</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Created By</th>
                  <th className="text-right p-4 font-medium text-muted-foreground">Actions</th>
                </tr>
              </thead>
              <tbody>
                {clusters.map((cluster) => (
                  <tr key={cluster.name} className="border-b border-border hover:bg-muted/50">
                    <td className="p-4">
                      <span className="font-medium">{cluster.name}</span>
                    </td>
                    <td className="p-4">
                      <span className="text-sm text-muted-foreground">{cluster.description || '-'}</span>
                    </td>
                    <td className="p-4">
                      <span className="text-sm font-mono">{cluster.api_server || '-'}</span>
                    </td>
                    <td className="p-4">
                      <Badge variant={cluster.status === 'connected' ? 'success' : 'destructive'}>
                        {cluster.status}
                      </Badge>
                    </td>
                    <td className="p-4">
                      <span className="text-sm text-muted-foreground">{cluster.created_by || '-'}</span>
                    </td>
                    <td className="p-4">
                      <div className="flex items-center justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openEditDialog(cluster)}
                          title="Edit cluster"
                        >
                          <Pencil className="w-4 h-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openDeleteDialog(cluster)}
                          className="text-destructive hover:text-destructive"
                          title="Delete cluster"
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* Add Cluster Dialog */}
      <Dialog open={addDialogOpen} onOpenChange={setAddDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Add Cluster</DialogTitle>
            <DialogDescription>
              Add a new Kubernetes cluster by providing its kubeconfig
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Cluster Name</label>
              <Input
                placeholder="e.g., production, staging"
                value={addName}
                onChange={(e) => setAddName(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">Only letters, numbers, hyphens, and underscores</p>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">Description</label>
              <Input
                placeholder="e.g., Production cluster in AWS"
                value={addDescription}
                onChange={(e) => setAddDescription(e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label className="text-sm font-medium">Kubeconfig</label>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => fileInputRef.current?.click()}
                >
                  <Upload className="w-3.5 h-3.5 mr-1.5" />
                  Upload File
                </Button>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".yaml,.yml,.conf,.config,-config"
                  className="hidden"
                  onChange={(e) => handleFileUpload(e, setAddKubeconfig)}
                />
              </div>
              <Textarea
                placeholder="Paste kubeconfig content here..."
                value={addKubeconfig}
                onChange={(e) => setAddKubeconfig(e.target.value)}
                rows={10}
                className="font-mono text-xs"
              />
            </div>

            <div className="flex items-center gap-3">
              <Button
                variant="outline"
                onClick={handleTestConnection}
                disabled={testingNew || !addKubeconfig.trim()}
              >
                <TestTube className="w-4 h-4 mr-2" />
                {testingNew ? 'Testing...' : 'Test Connection'}
              </Button>
              {testResult && (
                <span className={`text-sm ${testResult.startsWith('Connection successful') ? 'text-green-600' : 'text-destructive'}`}>
                  {testResult}
                </span>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setAddDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleAdd} disabled={adding || !addName.trim() || !addKubeconfig.trim()}>
              {adding ? 'Adding...' : 'Add Cluster'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Cluster Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Cluster: {editCluster?.name}</DialogTitle>
            <DialogDescription>
              Update cluster settings
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Description</label>
              <Input
                value={editDescription}
                onChange={(e) => setEditDescription(e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label className="text-sm font-medium">
                  Kubeconfig (optional, leave empty to keep current)
                </label>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => editFileInputRef.current?.click()}
                >
                  <Upload className="w-3.5 h-3.5 mr-1.5" />
                  Upload File
                </Button>
                <input
                  ref={editFileInputRef}
                  type="file"
                  accept=".yaml,.yml,.conf,.config,-config"
                  className="hidden"
                  onChange={(e) => handleFileUpload(e, setEditKubeconfig)}
                />
              </div>
              <Textarea
                placeholder="Paste new kubeconfig to update..."
                value={editKubeconfig}
                onChange={(e) => setEditKubeconfig(e.target.value)}
                rows={10}
                className="font-mono text-xs"
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleEdit} disabled={editing}>
              {editing ? 'Saving...' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Cluster Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Cluster</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete cluster "{deleteCluster?.name}"?
              This will also remove all associated user permissions. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
              {deleting ? 'Deleting...' : 'Delete Cluster'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
