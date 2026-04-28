import { useState, useEffect, useCallback } from 'react';
import { serviceApi } from '@/api';
import { useCluster } from '@/hooks/use-cluster';
import { useAuth } from '@/hooks/use-auth';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Loading } from '@/components/ui/spinner';
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
import type { K8sResource } from '@/types';
import { ApiError } from '@/api';
import {
  Plus,
  Search,
  Network,
  Edit,
  Trash2,
  Eye,
  RefreshCw,
  ArrowRight,
  AlertTriangle,
} from 'lucide-react';

const DEFAULT_SERVICE_YAML = `apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  type: ClusterIP
  selector:
    app: my-app
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
`;

interface ServiceInfo {
  type: string;
  ports: { port: string; targetPort: string; protocol: string; nodePort?: string }[];
  selector: Record<string, string>;
}

function parseServiceYaml(yaml: string): ServiceInfo {
  const typeMatch = yaml.match(/type:\s*(\S+)/);
  const svcType = typeMatch ? typeMatch[1] : 'ClusterIP';

  const ports: ServiceInfo['ports'] = [];
  const portsSection = yaml.match(/ports:\n((?:\s+-[\s\S]*?)*)(?=\n\S|\n\s{0,3}\S|$)/);
  if (portsSection) {
    const portBlocks = portsSection[1].split(/\n\s*-\s+/).filter(Boolean);
    for (const block of portBlocks) {
      const portMatch = block.match(/port:\s*(\S+)/);
      const targetMatch = block.match(/targetPort:\s*(\S+)/);
      const protoMatch = block.match(/protocol:\s*(\S+)/);
      const nodePortMatch = block.match(/nodePort:\s*(\S+)/);
      if (portMatch) {
        ports.push({
          port: portMatch[1],
          targetPort: targetMatch ? targetMatch[1] : portMatch[1],
          protocol: protoMatch ? protoMatch[1] : 'TCP',
          nodePort: nodePortMatch ? nodePortMatch[1] : undefined,
        });
      }
    }
  }

  const selector: Record<string, string> = {};
  const selectorMatch = yaml.match(/selector:\n((?:\s+\S+:\s*\S+\n?)*)/);
  if (selectorMatch) {
    const lines = selectorMatch[1].match(/(\S+):\s*(\S+)/g);
    lines?.forEach((line) => {
      const [key, value] = line.split(':').map((s) => s.trim());
      selector[key] = value;
    });
  }

  return { type: svcType, ports, selector };
}

export function ServicesPage() {
  const { selectedCluster, selectedNamespace } = useCluster();
  const { isAdmin } = useAuth();
  const { toast } = useToast();

  const [services, setServices] = useState<K8sResource[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');

  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editConfirmOpen, setEditConfirmOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedResource, setSelectedResource] = useState<K8sResource | null>(null);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const fetchServices = useCallback(async () => {
    if (!selectedCluster || !selectedNamespace) return;
    try {
      setLoading(true);
      const data = await serviceApi.list(selectedCluster, selectedNamespace);
      setServices(data || []);
    } catch (err) {
      setServices([]);
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch Services',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [selectedCluster, selectedNamespace, toast]);

  useEffect(() => {
    fetchServices();
  }, [fetchServices]);

  const handleCreate = async (yaml: string) => {
    try {
      setSaving(true);
      await serviceApi.create(selectedCluster, selectedNamespace, { yaml });
      toast({ title: 'Success', description: 'Service created successfully' });
      setCreateDialogOpen(false);
      fetchServices();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to create Service',
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
      await serviceApi.update(selectedCluster, selectedNamespace, selectedResource.name, {
        yaml,
        resourceVersion: selectedResource.resourceVersion,
      });
      toast({ title: 'Success', description: 'Service updated successfully' });
      setEditDialogOpen(false);
      fetchServices();
    } catch (err) {
      const isConflict = err instanceof ApiError && err.status === 409;
      toast({
        title: isConflict ? '版本冲突' : 'Error',
        description: isConflict
          ? '资源已被其他用户修改，请关闭编辑器后重新打开获取最新版本'
          : err instanceof Error ? err.message : 'Failed to update Service',
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
      await serviceApi.delete(selectedCluster, selectedNamespace, selectedResource.name);
      toast({ title: 'Success', description: 'Service deleted successfully' });
      setDeleteDialogOpen(false);
      fetchServices();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to delete Service',
        variant: 'destructive',
      });
    } finally {
      setDeleting(false);
    }
  };

  const filteredServices = services.filter((s) =>
    s.name.toLowerCase().includes(searchTerm.toLowerCase())
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
            <Network className="w-7 h-7 text-primary" />
            Services
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage network services in {selectedNamespace}
          </p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Create Service
        </Button>
      </div>

      {/* Search & Refresh */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search Services..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-9"
          />
        </div>
        <Button variant="outline" size="icon" onClick={fetchServices}>
          <RefreshCw className="w-4 h-4" />
        </Button>
      </div>

      {/* Service List */}
      {loading ? (
        <Loading text="Loading Services..." />
      ) : filteredServices.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Network className="w-12 h-12 text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">
              {searchTerm ? 'No Services found matching your search' : 'No Services in this namespace'}
            </p>
            <Button variant="outline" className="mt-4" onClick={() => setCreateDialogOpen(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Create your first Service
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filteredServices.map((svc) => {
            const info = parseServiceYaml(svc.yaml);
            return (
              <Card
                key={svc.name}
                className="group hover:border-primary/50 transition-colors cursor-pointer"
                onClick={() => {
                  setSelectedResource(svc);
                  setViewDialogOpen(true);
                }}
              >
                <CardHeader className="pb-3">
                  <CardTitle className="flex items-center justify-between text-base">
                    <div className="flex items-center gap-2 truncate">
                      <Network className="w-4 h-4 text-primary shrink-0" />
                      <span className="truncate">{svc.name}</span>
                    </div>
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedResource(svc);
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
                          setSelectedResource(svc);
                          setEditConfirmOpen(true);
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
                          setSelectedResource(svc);
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
                    {/* Type */}
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Type</span>
                      <Badge variant={
                        info.type === 'LoadBalancer' ? 'default' :
                        info.type === 'NodePort' ? 'warning' :
                        'outline'
                      }>
                        {info.type}
                      </Badge>
                    </div>

                    {/* Ports */}
                    {info.ports.length > 0 && (
                      <div className="space-y-1">
                        <span className="text-sm text-muted-foreground">Ports</span>
                        <div className="flex flex-wrap gap-1.5">
                          {info.ports.map((p, i) => (
                            <Badge key={i} variant="secondary" className="text-xs font-mono">
                              {p.port}
                              <ArrowRight className="w-3 h-3 mx-0.5" />
                              {p.targetPort}
                              {p.nodePort && <span className="ml-1 text-muted-foreground">:{p.nodePort}</span>}
                              <span className="ml-1 text-muted-foreground">/{p.protocol}</span>
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Selector */}
                    {Object.keys(info.selector).length > 0 && (
                      <div className="flex flex-wrap gap-1 pt-1">
                        {Object.entries(info.selector).slice(0, 3).map(([key, value]) => (
                          <Badge key={key} variant="outline" className="text-xs">
                            {key}: {value}
                          </Badge>
                        ))}
                        {Object.keys(info.selector).length > 3 && (
                          <Badge variant="outline" className="text-xs">
                            +{Object.keys(info.selector).length - 3} more
                          </Badge>
                        )}
                      </div>
                    )}
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
        title="Create Service"
        description="Define a new Service using YAML"
        initialYaml={DEFAULT_SERVICE_YAML}
        onSave={handleCreate}
        saving={saving}
      />

      {/* Edit Confirm Dialog */}
      <Dialog open={editConfirmOpen} onOpenChange={setEditConfirmOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-yellow-500" />
              确认编辑
            </DialogTitle>
            <DialogDescription>
              编辑 Service 属于风险操作，可能会影响正在运行的服务网络流量。
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <p className="text-sm text-muted-foreground">即将编辑：</p>
            <p className="mt-2 font-mono text-sm bg-muted px-3 py-2 rounded-md">
              {selectedResource?.name || ''}
            </p>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditConfirmOpen(false)}>
              取消
            </Button>
            <Button
              variant="default"
              onClick={() => {
                setEditConfirmOpen(false);
                setEditDialogOpen(true);
              }}
            >
              继续编辑
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <YamlEditorDialog
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        title={`Edit ${selectedResource?.name}`}
        description="Modify the Service configuration"
        initialYaml={selectedResource?.yaml || ''}
        onSave={handleEdit}
        saving={saving}
      />

      {/* View Dialog */}
      <YamlEditorDialog
        open={viewDialogOpen}
        onOpenChange={setViewDialogOpen}
        title={`View ${selectedResource?.name}`}
        description="Service configuration (read-only)"
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
        title="Delete Service"
        description="This action cannot be undone. This will permanently delete the Service from the cluster."
        resourceName={selectedResource?.name || ''}
        onConfirm={handleDelete}
        deleting={deleting}
      />
    </div>
  );
}
