import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { deploymentApi } from '@/api';
import { useCluster } from '@/hooks/use-cluster';
import { useAuth } from '@/hooks/use-auth';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Loading, Spinner } from '@/components/ui/spinner';
import { YamlEditorDialog } from '@/components/YamlEditorDialog';
import { DeleteConfirmDialog } from '@/components/DeleteConfirmDialog';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import type { Deployment } from '@/types';
import { ApiError } from '@/api';
import { classifyPhase } from '@/lib/utils';
import { useSearchHistory } from '@/hooks/use-search-history';
import { SearchHistoryChips } from '@/components/SearchHistoryChips';
import {
  Plus,
  Search,
  Box,
  Edit,
  Trash2,
  Eye,
  RefreshCw,
  Layers,
  Container,
  ExternalLink,
  RotateCw,
} from 'lucide-react';

const DEFAULT_DEPLOYMENT_YAML = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: my-container
        image: nginx:latest
        ports:
        - containerPort: 80
`;

interface DeploymentInfo {
  replicas: number;
  image: string;
  labels: Record<string, string>;
}

function parseDeploymentYaml(yaml: string): DeploymentInfo {
  const replicasMatch = yaml.match(/replicas:\s*(\d+)/);
  const imageMatch = yaml.match(/image:\s*([^\s\n]+)/);
  const labelsMatch = yaml.match(/labels:\n((?:\s+\w+:\s*\w+\n?)+)/);

  const labels: Record<string, string> = {};
  if (labelsMatch) {
    const labelLines = labelsMatch[1].match(/(\w+):\s*(\w+)/g);
    labelLines?.forEach((line) => {
      const [key, value] = line.split(':').map((s) => s.trim());
      labels[key] = value;
    });
  }

  return {
    replicas: replicasMatch ? parseInt(replicasMatch[1], 10) : 1,
    image: imageMatch ? imageMatch[1] : 'unknown',
    labels,
  };
}

export function DeploymentsPage() {
  const { selectedCluster, selectedNamespace } = useCluster();
  const { isAdmin } = useAuth();
  const { toast } = useToast();
  const navigate = useNavigate();

  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const { history, add: addHistory, remove: removeHistory, clear: clearHistory } = useSearchHistory('searchHistory:deployments');

  // 输入停止 1.2s 后自动入库
  useEffect(() => {
    if (!searchTerm.trim()) return;
    const t = setTimeout(() => addHistory(searchTerm), 1200);
    return () => clearTimeout(t);
  }, [searchTerm, addHistory]);

  // Dialog states
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [restartDialogOpen, setRestartDialogOpen] = useState(false);
  const [selectedResource, setSelectedResource] = useState<Deployment | null>(null);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [restarting, setRestarting] = useState(false);

  const fetchDeployments = useCallback(async () => {
    if (!selectedCluster || !selectedNamespace) return;
    try {
      setLoading(true);
      const data = await deploymentApi.list(selectedCluster, selectedNamespace);
      setDeployments(data || []);
    } catch (err) {
      setDeployments([]);
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch Deployments',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [selectedCluster, selectedNamespace, toast]);

  useEffect(() => {
    fetchDeployments();
  }, [fetchDeployments]);

  // 每 5s 静默轮询，自动感知滚动更新状态
  useEffect(() => {
    if (!selectedCluster || !selectedNamespace) return;
    const interval = setInterval(async () => {
      try {
        const data = await deploymentApi.list(selectedCluster, selectedNamespace);
        setDeployments(data || []);
      } catch {
        // 静默失败
      }
    }, 5000);
    return () => clearInterval(interval);
  }, [selectedCluster, selectedNamespace]);

  const handleCreate = async (yaml: string) => {
    try {
      setSaving(true);
      await deploymentApi.create(selectedCluster, selectedNamespace, { yaml });
      toast({ title: 'Success', description: 'Deployment created successfully' });
      setCreateDialogOpen(false);
      fetchDeployments();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to create Deployment',
        variant: 'destructive',
      });
    } finally {
      setSaving(false);
    }
  };

  const handleEdit = async (yaml: string) => {
    if (!selectedResource) return;
    try {
      setSaving(true);
      await deploymentApi.update(selectedCluster, selectedNamespace, selectedResource.name, {
        yaml,
        resourceVersion: selectedResource.resourceVersion,
      });
      toast({ title: 'Success', description: 'Deployment updated successfully' });
      setEditDialogOpen(false);
      fetchDeployments();
    } catch (err) {
      const isConflict = err instanceof ApiError && err.status === 409;
      toast({
        title: isConflict ? '版本冲突' : 'Error',
        description: isConflict
          ? '资源已被其他用户修改，请关闭编辑器后重新打开获取最新版本'
          : err instanceof Error ? err.message : 'Failed to update Deployment',
        variant: 'destructive',
      });
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!selectedResource) return;
    try {
      setDeleting(true);
      await deploymentApi.delete(selectedCluster, selectedNamespace, selectedResource.name);
      toast({ title: 'Success', description: 'Deployment deleted successfully' });
      setDeleteDialogOpen(false);
      fetchDeployments();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to delete Deployment',
        variant: 'destructive',
      });
    } finally {
      setDeleting(false);
    }
  };

  const handleRestart = async () => {
    if (!selectedResource) return;
    try {
      setRestarting(true);
      await deploymentApi.restart(selectedCluster, selectedNamespace, selectedResource.name);
      toast({ title: 'Success', description: `Deployment "${selectedResource.name}" restarting` });
      setRestartDialogOpen(false);
      fetchDeployments();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to restart Deployment',
        variant: 'destructive',
      });
    } finally {
      setRestarting(false);
    }
  };

  const filteredDeployments = deployments.filter((d) =>
    d.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  if (!selectedCluster || !selectedNamespace) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-muted-foreground">Please select a cluster and namespace</p>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold flex items-center gap-2">
            <Box className="w-7 h-7 text-primary" />
            Deployments
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage application deployments in {selectedNamespace}
          </p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Create Deployment
        </Button>
      </div>

      {/* Search & Refresh */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search Deployments..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            onKeyDown={(e) => { if (e.key === 'Enter' && searchTerm.trim()) addHistory(searchTerm); }}
            className="pl-9"
          />
        </div>
        <Button variant="outline" size="icon" onClick={fetchDeployments}>
          <RefreshCw className="w-4 h-4" />
        </Button>
      </div>

      <SearchHistoryChips
        history={history}
        current={searchTerm}
        onSelect={setSearchTerm}
        onRemove={removeHistory}
        onClear={clearHistory}
      />

      {/* Deployment List */}
      {loading ? (
        <Loading text="Loading Deployments..." />
      ) : filteredDeployments.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Layers className="w-12 h-12 text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">
              {searchTerm ? 'No Deployments found matching your search' : 'No Deployments in this namespace'}
            </p>
            <Button variant="outline" className="mt-4" onClick={() => setCreateDialogOpen(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Create your first Deployment
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filteredDeployments.map((deployment) => {
            const info = parseDeploymentYaml(deployment.yaml);
            const isRolling = deployment.updated_replicas < deployment.replicas;
            const errorPodCount = Object.entries(deployment.pod_status_counts || {})
              .filter(([status]) => classifyPhase(status) === 'error')
              .reduce((sum, [, count]) => sum + count, 0);
            const hasError = errorPodCount > 0 || (deployment.replicas > 0 && deployment.ready_replicas === 0);
            const isHealthy = !isRolling && !hasError && deployment.ready_replicas === deployment.replicas;
            const replicasVariant = hasError ? 'destructive' : isHealthy ? 'success' : 'warning';
            const updatedPct = deployment.replicas > 0 ? (deployment.updated_replicas / deployment.replicas) * 100 : 0;
            const readyPct = deployment.replicas > 0 ? (deployment.ready_replicas / deployment.replicas) * 100 : 0;
            return (
              <Card
                key={deployment.name}
                className="group hover:border-primary/50 transition-colors cursor-pointer"
                onClick={() => {
                  setSelectedResource(deployment);
                  setViewDialogOpen(true);
                }}
              >
                <CardHeader className="pb-3">
                  <CardTitle className="flex items-center justify-between text-base">
                    <div className="flex items-center gap-2 truncate">
                      <Box className="w-4 h-4 text-primary shrink-0" />
                      <span className="truncate">{deployment.name}</span>
                      {isRolling && (
                        <RotateCw className="w-3.5 h-3.5 text-blue-500 animate-spin shrink-0" />
                      )}
                    </div>
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedResource(deployment);
                          setViewDialogOpen(true);
                        }}
                      >
                        <Eye className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedResource(deployment);
                          setEditDialogOpen(true);
                        }}
                      >
                        <Edit className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        title="Restart Deployment"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedResource(deployment);
                          setRestartDialogOpen(true);
                        }}
                      >
                        <RotateCw className="w-4 h-4" />
                      </Button>
                      {isAdmin && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-destructive hover:text-destructive"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedResource(deployment);
                          setDeleteDialogOpen(true);
                        }}
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                      )}
                    </div>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {/* Replicas Status */}
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground flex items-center gap-1.5">
                        <Layers className="w-3.5 h-3.5" />
                        Replicas
                      </span>
                      <Badge variant={replicasVariant}>
                        {deployment.ready_replicas}/{deployment.replicas}
                      </Badge>
                    </div>

                    {/* Rolling Update Progress */}
                    {isRolling && (
                      <div className="border-t pt-2 space-y-1.5">
                        <div className="flex items-center gap-1 text-xs text-blue-500 font-medium mb-1">
                          <RotateCw className="w-3 h-3 animate-spin" />
                          滚动更新中
                        </div>
                        <div className="space-y-1">
                          <div className="flex justify-between text-xs text-muted-foreground">
                            <span>UP-TO-DATE</span>
                            <span className="font-mono font-medium text-foreground">{deployment.updated_replicas}/{deployment.replicas}</span>
                          </div>
                          <div className="h-1.5 bg-muted rounded-full overflow-hidden">
                            <div
                              className="h-full bg-blue-500 rounded-full transition-all duration-500"
                              style={{ width: `${updatedPct}%` }}
                            />
                          </div>
                        </div>
                        <div className="space-y-1">
                          <div className="flex justify-between text-xs text-muted-foreground">
                            <span>READY</span>
                            <span className="font-mono font-medium text-foreground">{deployment.ready_replicas}/{deployment.replicas}</span>
                          </div>
                          <div className="h-1.5 bg-muted rounded-full overflow-hidden">
                            <div
                              className="h-full bg-yellow-500 rounded-full transition-all duration-500"
                              style={{ width: `${readyPct}%` }}
                            />
                          </div>
                        </div>
                      </div>
                    )}

                    {/* Pod Count - Clickable with status breakdown */}
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground flex items-center gap-1.5">
                        <Container className="w-3.5 h-3.5" />
                        Pods
                      </span>
                      <div className="flex items-center gap-2">
                        {deployment.pod_status_counts && Object.entries(deployment.pod_status_counts)
                          .filter(([, count]) => count > 0)
                          .map(([status, count]) => {
                            const category = classifyPhase(status);
                            const dotClass =
                              category === 'healthy' ? (status === 'Succeeded' ? 'bg-blue-400' : 'bg-green-500') :
                              category === 'pending' ? 'bg-yellow-500' :
                              category === 'error' ? 'bg-red-500' :
                              'bg-gray-400';
                            return (
                              <span
                                key={status}
                                className="flex items-center gap-0.5 text-xs font-medium"
                                title={status}
                              >
                                <span className={`inline-block w-2 h-2 rounded-full ${dotClass}`} />
                                {count}
                              </span>
                            );
                          })}
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 px-1 text-primary hover:text-primary"
                          onClick={(e) => {
                            e.stopPropagation();
                            navigate(`/pods?deployment=${encodeURIComponent(deployment.name)}`);
                          }}
                        >
                          <ExternalLink className="w-3 h-3" />
                        </Button>
                      </div>
                    </div>

                    {/* Image */}
                    <div className="flex items-start justify-between text-sm gap-2">
                      <span className="text-muted-foreground shrink-0 flex items-center gap-1.5">
                        <Container className="w-3.5 h-3.5" />
                        Image
                      </span>
                      <span className="font-mono text-xs text-right break-all" title={info.image}>
                        {info.image.includes('/') ? info.image.split('/').pop() : info.image}
                      </span>
                    </div>

                    {/* Labels */}
                    <div className="flex flex-wrap gap-1 pt-1">
                      {Object.entries(info.labels).slice(0, 3).map(([key, value]) => (
                        <Badge key={key} variant="outline" className="text-xs">
                          {key}: {value}
                        </Badge>
                      ))}
                      {Object.keys(info.labels).length > 3 && (
                        <Badge variant="outline" className="text-xs">
                          +{Object.keys(info.labels).length - 3} more
                        </Badge>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {/* Create Dialog */}
      <YamlEditorDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        title="Create Deployment"
        description="Define a new Deployment using YAML"
        initialYaml={DEFAULT_DEPLOYMENT_YAML}
        onSave={handleCreate}
        saving={saving}
      />

      {/* Edit Dialog */}
      <YamlEditorDialog
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        title={`Edit ${selectedResource?.name}`}
        description="Modify the Deployment configuration"
        initialYaml={selectedResource?.yaml || ''}
        onSave={handleEdit}
        saving={saving}
      />

      {/* View Dialog */}
      <YamlEditorDialog
        open={viewDialogOpen}
        onOpenChange={setViewDialogOpen}
        title={`View ${selectedResource?.name}`}
        description="Deployment configuration (read-only)"
        initialYaml={selectedResource?.yaml || ''}
        readOnly={true}
        onSave={async () => {
          setViewDialogOpen(false);
          setEditDialogOpen(true);
        }}
        saving={false}
      />

      {/* Delete Dialog */}
      <DeleteConfirmDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="Delete Deployment"
        description="This action cannot be undone. This will permanently delete the Deployment and stop all its pods."
        resourceName={selectedResource?.name || ''}
        onConfirm={handleDelete}
        deleting={deleting}
      />

      {/* Restart Confirm Dialog */}
      <Dialog open={restartDialogOpen} onOpenChange={setRestartDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <RotateCw className="w-5 h-5" />
              重启 Deployment
            </DialogTitle>
            <DialogDescription>
              此操作将触发滚动重启，所有 Pod 将依次重建。
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <p className="text-sm text-muted-foreground">即将重启：</p>
            <p className="mt-2 font-mono text-sm bg-muted px-3 py-2 rounded-md">
              {selectedResource?.name || ''}
            </p>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRestartDialogOpen(false)} disabled={restarting}>
              取消
            </Button>
            <Button onClick={handleRestart} disabled={restarting}>
              {restarting && <Spinner size="sm" className="mr-2" />}
              {restarting ? '重启中...' : '确认重启'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
