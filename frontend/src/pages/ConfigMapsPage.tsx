import { useState, useEffect, useCallback } from 'react';
import { configMapApi } from '@/api';
import { useCluster } from '@/hooks/use-cluster';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Loading } from '@/components/ui/spinner';
import { YamlEditorDialog } from '@/components/YamlEditorDialog';
import { DeleteConfirmDialog } from '@/components/DeleteConfirmDialog';
import { formatConfigMapYaml } from '@/lib/utils';
import type { K8sResource } from '@/types';
import {
  Plus,
  Search,
  Settings,
  Edit,
  Trash2,
  Eye,
  RefreshCw,
  FileCode,
} from 'lucide-react';

const DEFAULT_CONFIGMAP_YAML = `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  key1: value1
  key2: value2
`;

export function ConfigMapsPage() {
  const { selectedCluster, selectedNamespace } = useCluster();
  const { toast } = useToast();

  const [configMaps, setConfigMaps] = useState<K8sResource[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');

  // Dialog states
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedResource, setSelectedResource] = useState<K8sResource | null>(null);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const fetchConfigMaps = useCallback(async () => {
    if (!selectedCluster || !selectedNamespace) return;
    try {
      setLoading(true);
      const data = await configMapApi.list(selectedCluster, selectedNamespace);
      setConfigMaps(data || []);
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch ConfigMaps',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [selectedCluster, selectedNamespace, toast]);

  useEffect(() => {
    fetchConfigMaps();
  }, [fetchConfigMaps]);

  const handleCreate = async (yaml: string) => {
    try {
      setSaving(true);
      await configMapApi.create(selectedCluster, selectedNamespace, { yaml });
      toast({ title: 'Success', description: 'ConfigMap created successfully' });
      setCreateDialogOpen(false);
      fetchConfigMaps();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to create ConfigMap',
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
      await configMapApi.update(selectedCluster, selectedNamespace, selectedResource.name, { yaml });
      toast({ title: 'Success', description: 'ConfigMap updated successfully' });
      setEditDialogOpen(false);
      fetchConfigMaps();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to update ConfigMap',
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
      await configMapApi.delete(selectedCluster, selectedNamespace, selectedResource.name);
      toast({ title: 'Success', description: 'ConfigMap deleted successfully' });
      setDeleteDialogOpen(false);
      fetchConfigMaps();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to delete ConfigMap',
        variant: 'destructive',
      });
    } finally {
      setDeleting(false);
    }
  };

  const filteredConfigMaps = configMaps.filter((cm) =>
    cm.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const parseDataKeys = (yaml: string): string[] => {
    const match = yaml.match(/data:\n([\s\S]*?)(?=\n[^\s]|$)/);
    if (!match) return [];
    const dataSection = match[1];
    const keys = dataSection.match(/^\s{2}(\w+):/gm);
    return keys ? keys.map((k) => k.trim().replace(':', '')) : [];
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
            <Settings className="w-7 h-7 text-primary" />
            ConfigMaps
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage configuration data in {selectedNamespace}
          </p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Create ConfigMap
        </Button>
      </div>

      {/* Search & Refresh */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search ConfigMaps..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-9"
          />
        </div>
        <Button variant="outline" size="icon" onClick={fetchConfigMaps}>
          <RefreshCw className="w-4 h-4" />
        </Button>
      </div>

      {/* ConfigMap List */}
      {loading ? (
        <Loading text="Loading ConfigMaps..." />
      ) : filteredConfigMaps.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <FileCode className="w-12 h-12 text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">
              {searchTerm ? 'No ConfigMaps found matching your search' : 'No ConfigMaps in this namespace'}
            </p>
            <Button variant="outline" className="mt-4" onClick={() => setCreateDialogOpen(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Create your first ConfigMap
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filteredConfigMaps.map((cm) => {
            const dataKeys = parseDataKeys(cm.yaml);
            return (
              <Card
                key={cm.name}
                className="group hover:border-primary/50 transition-colors cursor-pointer"
                onClick={() => {
                  setSelectedResource(cm);
                  setViewDialogOpen(true);
                }}
              >
                <CardHeader className="pb-3">
                  <CardTitle className="flex items-center justify-between text-base">
                    <div className="flex items-center gap-2 truncate">
                      <Settings className="w-4 h-4 text-primary shrink-0" />
                      <span className="truncate">{cm.name}</span>
                    </div>
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedResource(cm);
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
                          setSelectedResource(cm);
                          setEditDialogOpen(true);
                        }}
                      >
                        <Edit className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-destructive hover:text-destructive"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedResource(cm);
                          setDeleteDialogOpen(true);
                        }}
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    <p className="text-xs text-muted-foreground">
                      {dataKeys.length} data {dataKeys.length === 1 ? 'key' : 'keys'}
                    </p>
                    <div className="flex flex-wrap gap-1">
                      {dataKeys.slice(0, 5).map((key) => (
                        <Badge key={key} variant="secondary" className="text-xs">
                          {key}
                        </Badge>
                      ))}
                      {dataKeys.length > 5 && (
                        <Badge variant="outline" className="text-xs">
                          +{dataKeys.length - 5} more
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
        title="Create ConfigMap"
        description="Define a new ConfigMap using YAML"
        initialYaml={DEFAULT_CONFIGMAP_YAML}
        onSave={handleCreate}
        saving={saving}
      />

      {/* Edit Dialog */}
      <YamlEditorDialog
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        title={`Edit ${selectedResource?.name}`}
        description="Modify the ConfigMap configuration"
        initialYaml={selectedResource ? formatConfigMapYaml(selectedResource.yaml) : ''}
        onSave={handleEdit}
        saving={saving}
      />

      {/* View Dialog */}
      <YamlEditorDialog
        open={viewDialogOpen}
        onOpenChange={setViewDialogOpen}
        title={`View ${selectedResource?.name}`}
        description="ConfigMap configuration (read-only)"
        initialYaml={selectedResource ? formatConfigMapYaml(selectedResource.yaml) : ''}
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
        title="Delete ConfigMap"
        description="This action cannot be undone. This will permanently delete the ConfigMap from the cluster."
        resourceName={selectedResource?.name || ''}
        onConfirm={handleDelete}
        deleting={deleting}
      />
    </div>
  );
}
