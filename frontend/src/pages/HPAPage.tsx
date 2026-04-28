import { useState, useEffect, useCallback } from 'react';
import { hpaApi } from '@/api';
import { useCluster } from '@/hooks/use-cluster';
import { useAuth } from '@/hooks/use-auth';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Loading } from '@/components/ui/spinner';
import { Progress } from '@/components/ui/progress';
import { YamlEditorDialog } from '@/components/YamlEditorDialog';
import { DeleteConfirmDialog } from '@/components/DeleteConfirmDialog';
import type { HPA } from '@/types';
import { ApiError } from '@/api';
import {
  Plus,
  Search,
  Edit,
  Trash2,
  Eye,
  RefreshCw,
  FileCode,
  Gauge,
  ArrowUpDown,
  Cpu,
  MemoryStick,
  CheckCircle2,
  AlertCircle,
  Target,
} from 'lucide-react';

const DEFAULT_HPA_YAML = `apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-deployment
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 80
`;

export function HPAPage() {
  const { selectedCluster, selectedNamespace } = useCluster();
  const { isAdmin } = useAuth();
  const { toast } = useToast();

  const [hpas, setHpas] = useState<HPA[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');

  // Dialog states
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedResource, setSelectedResource] = useState<HPA | null>(null);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const fetchHPAs = useCallback(async () => {
    if (!selectedCluster || !selectedNamespace) return;
    try {
      setLoading(true);
      const data = await hpaApi.list(selectedCluster, selectedNamespace);
      setHpas(data || []);
    } catch (err) {
      setHpas([]);
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch HPAs',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [selectedCluster, selectedNamespace, toast]);

  useEffect(() => {
    fetchHPAs();
  }, [fetchHPAs]);

  const handleCreate = async (yaml: string) => {
    try {
      setSaving(true);
      await hpaApi.create(selectedCluster, selectedNamespace, { yaml });
      toast({ title: 'Success', description: 'HPA created successfully' });
      setCreateDialogOpen(false);
      fetchHPAs();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to create HPA',
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
      await hpaApi.update(selectedCluster, selectedNamespace, selectedResource.name, {
        yaml,
        resourceVersion: selectedResource.resourceVersion,
      });
      toast({ title: 'Success', description: 'HPA updated successfully' });
      setEditDialogOpen(false);
      fetchHPAs();
    } catch (err) {
      const isConflict = err instanceof ApiError && err.status === 409;
      toast({
        title: isConflict ? '版本冲突' : 'Error',
        description: isConflict
          ? '资源已被其他用户修改，请关闭编辑器后重新打开获取最新版本'
          : err instanceof Error ? err.message : 'Failed to update HPA',
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
      await hpaApi.delete(selectedCluster, selectedNamespace, selectedResource.name);
      toast({ title: 'Success', description: 'HPA deleted successfully' });
      setDeleteDialogOpen(false);
      fetchHPAs();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to delete HPA',
        variant: 'destructive',
      });
    } finally {
      setDeleting(false);
    }
  };

  const filteredHPAs = hpas.filter((hpa) =>
    hpa.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const getScalingStatus = (hpa: HPA) => {
    const scalingActive = hpa.conditions?.find(c => c.type === 'ScalingActive');
    const ableToScale = hpa.conditions?.find(c => c.type === 'AbleToScale');

    if (scalingActive?.status === 'True') {
      return { status: 'active', label: 'Active', variant: 'success' as const, message: scalingActive.message };
    }

    if (scalingActive?.status === 'False') {
      // Show warning with reason
      return {
        status: 'warning',
        label: scalingActive.reason || 'Warning',
        variant: 'destructive' as const,
        message: scalingActive.message
      };
    }

    if (ableToScale?.status === 'True') {
      return { status: 'ready', label: 'Ready', variant: 'secondary' as const, message: ableToScale.message };
    }

    return { status: 'unknown', label: 'Unknown', variant: 'outline' as const, message: '' };
  };

  const getUtilizationColor = (current: number | undefined, target: number | undefined) => {
    if (current === undefined || target === undefined) return 'bg-muted';
    const ratio = current / target;
    if (ratio < 0.7) return 'bg-green-500';
    if (ratio < 0.9) return 'bg-yellow-500';
    return 'bg-red-500';
  };

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
            <Gauge className="w-7 h-7 text-primary" />
            Horizontal Pod Autoscalers
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage auto-scaling policies in {selectedNamespace}
          </p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Create HPA
        </Button>
      </div>

      {/* Search & Refresh */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search HPAs..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-9"
          />
        </div>
        <Button variant="outline" size="icon" onClick={fetchHPAs}>
          <RefreshCw className="w-4 h-4" />
        </Button>
      </div>

      {/* HPA List */}
      {loading ? (
        <Loading text="Loading HPAs..." />
      ) : filteredHPAs.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <FileCode className="w-12 h-12 text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">
              {searchTerm ? 'No HPAs found matching your search' : 'No HPAs in this namespace'}
            </p>
            <Button variant="outline" className="mt-4" onClick={() => setCreateDialogOpen(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Create your first HPA
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {filteredHPAs.map((hpa) => {
            const scalingStatus = getScalingStatus(hpa);

            return (
              <Card
                key={hpa.name}
                className="group hover:border-primary/50 transition-colors cursor-pointer"
                onClick={() => {
                  setSelectedResource(hpa);
                  setViewDialogOpen(true);
                }}
              >
                <CardHeader className="pb-3">
                  <CardTitle className="flex items-center justify-between text-base">
                    <div className="flex items-center gap-2 truncate">
                      <Gauge className="w-4 h-4 text-primary shrink-0" />
                      <span className="truncate">{hpa.name}</span>
                    </div>
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedResource(hpa);
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
                          setSelectedResource(hpa);
                          setEditDialogOpen(true);
                        }}
                      >
                        <Edit className="w-4 h-4" />
                      </Button>
                      {isAdmin && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-destructive hover:text-destructive"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedResource(hpa);
                          setDeleteDialogOpen(true);
                        }}
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                      )}
                    </div>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Target & Status */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2 text-sm">
                      <Target className="w-4 h-4 text-muted-foreground" />
                      <Badge variant="outline">
                        {hpa.scale_target_ref?.kind}/{hpa.scale_target_ref?.name}
                      </Badge>
                    </div>
                    <div className="relative group/status">
                      <Badge variant={scalingStatus.variant} className="cursor-help">
                        {scalingStatus.status === 'active' ? (
                          <CheckCircle2 className="w-3 h-3 mr-1" />
                        ) : scalingStatus.status === 'warning' ? (
                          <AlertCircle className="w-3 h-3 mr-1" />
                        ) : null}
                        {scalingStatus.label}
                      </Badge>
                      {scalingStatus.message && (
                        <div className="absolute right-0 top-full mt-1 z-50 hidden group-hover/status:block w-64 p-2 text-xs bg-popover border rounded-md shadow-md">
                          {scalingStatus.message}
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Replicas */}
                  <div className="flex items-center gap-2 text-sm">
                    <ArrowUpDown className="w-4 h-4 text-muted-foreground" />
                    <span className="text-muted-foreground">Replicas:</span>
                    <span className="font-medium">
                      {hpa.current_replicas} / {hpa.desired_replicas}
                    </span>
                    <span className="text-muted-foreground">
                      (min: {hpa.min_replicas}, max: {hpa.max_replicas})
                    </span>
                  </div>

                  {/* CPU Metric */}
                  {hpa.cpu && (
                    <div className="space-y-1.5">
                      <div className="flex items-center justify-between text-sm">
                        <div className="flex items-center gap-2">
                          <Cpu className="w-4 h-4 text-blue-500" />
                          <span>CPU</span>
                        </div>
                        <span className="text-muted-foreground">
                          {hpa.cpu.current_utilization ?? '-'}% / {hpa.cpu.target_utilization ?? '-'}%
                        </span>
                      </div>
                      <Progress
                        value={hpa.cpu.current_utilization ?? 0}
                        max={100}
                        className="h-2"
                        indicatorClassName={getUtilizationColor(hpa.cpu.current_utilization, hpa.cpu.target_utilization)}
                      />
                    </div>
                  )}

                  {/* Memory Metric */}
                  {hpa.memory && (
                    <div className="space-y-1.5">
                      <div className="flex items-center justify-between text-sm">
                        <div className="flex items-center gap-2">
                          <MemoryStick className="w-4 h-4 text-purple-500" />
                          <span>Memory</span>
                        </div>
                        <span className="text-muted-foreground">
                          {hpa.memory.current_utilization ?? '-'}% / {hpa.memory.target_utilization ?? '-'}%
                        </span>
                      </div>
                      <Progress
                        value={hpa.memory.current_utilization ?? 0}
                        max={100}
                        className="h-2"
                        indicatorClassName={getUtilizationColor(hpa.memory.current_utilization, hpa.memory.target_utilization)}
                      />
                    </div>
                  )}

                  {/* Other Metrics */}
                  {hpa.metrics && hpa.metrics.filter(m => m.name !== 'cpu' && m.name !== 'memory').length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {hpa.metrics.filter(m => m.name !== 'cpu' && m.name !== 'memory').map((metric, idx) => (
                        <Badge key={idx} variant="secondary" className="text-xs">
                          {metric.name}: {metric.current_value || metric.current_utilization || '-'}
                        </Badge>
                      ))}
                    </div>
                  )}
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
        title="Create HPA"
        description="Define a new Horizontal Pod Autoscaler using YAML"
        initialYaml={DEFAULT_HPA_YAML}
        onSave={handleCreate}
        saving={saving}
      />

      {/* Edit Dialog */}
      <YamlEditorDialog
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        title={`Edit ${selectedResource?.name}`}
        description="Modify the HPA configuration"
        initialYaml={selectedResource?.yaml || ''}
        onSave={handleEdit}
        saving={saving}
      />

      {/* View Dialog */}
      <YamlEditorDialog
        open={viewDialogOpen}
        onOpenChange={setViewDialogOpen}
        title={`View ${selectedResource?.name}`}
        description="HPA configuration (read-only)"
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
        title="Delete HPA"
        description="This action cannot be undone. This will permanently delete the HPA from the cluster."
        resourceName={selectedResource?.name || ''}
        onConfirm={handleDelete}
        deleting={deleting}
      />
    </div>
  );
}
