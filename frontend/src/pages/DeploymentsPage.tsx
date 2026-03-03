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
import { Loading } from '@/components/ui/spinner';
import { YamlEditorDialog } from '@/components/YamlEditorDialog';
import { DeleteConfirmDialog } from '@/components/DeleteConfirmDialog';
import type { Deployment } from '@/types';
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

  // Dialog states
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedResource, setSelectedResource] = useState<Deployment | null>(null);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

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
      await deploymentApi.update(selectedCluster, selectedNamespace, selectedResource.name, { yaml });
      toast({ title: 'Success', description: 'Deployment updated successfully' });
      setEditDialogOpen(false);
      fetchDeployments();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to update Deployment',
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
            className="pl-9"
          />
        </div>
        <Button variant="outline" size="icon" onClick={fetchDeployments}>
          <RefreshCw className="w-4 h-4" />
        </Button>
      </div>

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
            const isHealthy = deployment.ready_replicas === deployment.replicas;
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
                      <Badge variant={isHealthy ? 'success' : 'warning'}>
                        {deployment.ready_replicas}/{deployment.replicas}
                      </Badge>
                    </div>

                    {/* Pod Count - Clickable */}
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground flex items-center gap-1.5">
                        <Container className="w-3.5 h-3.5" />
                        Pods
                      </span>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 px-2 text-primary hover:text-primary"
                        onClick={(e) => {
                          e.stopPropagation();
                          navigate(`/pods?deployment=${encodeURIComponent(deployment.name)}`);
                        }}
                      >
                        {deployment.pod_count}
                        <ExternalLink className="w-3 h-3 ml-1" />
                      </Button>
                    </div>

                    {/* Image */}
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Image</span>
                      <span className="font-mono text-xs truncate max-w-[150px]" title={info.image}>
                        {info.image}
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
    </div>
  );
}
