import { useState, useEffect, useCallback, useRef } from 'react';
import { nodeApi } from '@/api';
import { useCluster } from '@/hooks/use-cluster';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Loading } from '@/components/ui/spinner';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { formatMemory, formatCpu } from '@/lib/utils';
import type { NodeListItem, NodeDetail } from '@/types';
import {
  Search,
  RefreshCw,
  Server,
  Cpu,
  HardDrive,
  Clock,
  Eye,
  Activity,
  CheckCircle2,
  AlertTriangle,
  Container,
  Tag,
  Shield,
  LayoutGrid,
  List,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';

type ViewMode = 'card' | 'table';

const PAGE_SIZE = 50;

export function NodesPage() {
  const { selectedCluster } = useCluster();
  const { toast } = useToast();

  const [nodes, setNodes] = useState<NodeListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [viewMode, setViewMode] = useState<ViewMode>('card');
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  // Debounce timer ref
  const searchTimerRef = useRef<ReturnType<typeof setTimeout>>();

  // Dialog states
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [selectedNode, setSelectedNode] = useState<NodeDetail | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  const fetchNodes = useCallback(async () => {
    if (!selectedCluster) return;
    try {
      setLoading(true);
      const data = await nodeApi.list(selectedCluster, {
        search: searchTerm || undefined,
        page,
        page_size: PAGE_SIZE,
      });
      setNodes(data?.items || []);
      setTotal(data?.total || 0);
    } catch (err) {
      setNodes([]);
      setTotal(0);
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch nodes',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [selectedCluster, searchTerm, page, toast]);

  useEffect(() => {
    fetchNodes();
  }, [fetchNodes]);

  // Reset page on cluster change
  useEffect(() => {
    setPage(1);
    setSearchTerm('');
  }, [selectedCluster]);

  const handleSearchChange = (value: string) => {
    setSearchTerm(value);
    if (searchTimerRef.current) {
      clearTimeout(searchTimerRef.current);
    }
    searchTimerRef.current = setTimeout(() => {
      setPage(1);
    }, 300);
  };

  const fetchNodeDetail = async (nodeName: string) => {
    if (!selectedCluster) return;
    try {
      setLoadingDetail(true);
      const data = await nodeApi.get(selectedCluster, nodeName);
      setSelectedNode(data);
      setViewDialogOpen(true);
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch node details',
        variant: 'destructive',
      });
    } finally {
      setLoadingDetail(false);
    }
  };

  const totalPages = Math.ceil(total / PAGE_SIZE);

  const formatAge = (dateStr: string) => {
    const created = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - created.getTime();
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffDays > 365) return `${Math.floor(diffDays / 365)}y`;
    if (diffDays > 30) return `${Math.floor(diffDays / 30)}mo`;
    if (diffDays > 0) return `${diffDays}d`;

    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    if (diffHours > 0) return `${diffHours}h`;

    const diffMins = Math.floor(diffMs / (1000 * 60));
    return `${diffMins}m`;
  };

  if (!selectedCluster) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-muted-foreground">Please select a cluster</p>
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
            Nodes
          </h1>
          <p className="text-muted-foreground mt-1">
            View cluster nodes in {selectedCluster}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="text-sm">
            {total} nodes
          </Badge>
        </div>
      </div>

      {/* Search & Controls */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search nodes..."
            value={searchTerm}
            onChange={(e) => handleSearchChange(e.target.value)}
            className="pl-9"
          />
        </div>
        <div className="flex items-center border rounded-md">
          <Button
            variant={viewMode === 'card' ? 'secondary' : 'ghost'}
            size="icon"
            className="rounded-r-none"
            onClick={() => setViewMode('card')}
          >
            <LayoutGrid className="w-4 h-4" />
          </Button>
          <Button
            variant={viewMode === 'table' ? 'secondary' : 'ghost'}
            size="icon"
            className="rounded-l-none"
            onClick={() => setViewMode('table')}
          >
            <List className="w-4 h-4" />
          </Button>
        </div>
        <Button variant="outline" size="icon" onClick={fetchNodes}>
          <RefreshCw className="w-4 h-4" />
        </Button>
      </div>

      {/* Node List */}
      {loading ? (
        <Loading text="Loading nodes..." />
      ) : nodes.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Server className="w-12 h-12 text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">
              {searchTerm ? 'No nodes found matching your search' : 'No nodes in this cluster'}
            </p>
          </CardContent>
        </Card>
      ) : viewMode === 'card' ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {nodes.map((node) => (
            <Card
              key={node.name}
              className="group hover:border-primary/50 transition-colors cursor-pointer"
              onClick={() => fetchNodeDetail(node.name)}
            >
              <CardHeader className="pb-3">
                <CardTitle className="flex items-center justify-between text-base">
                  <div className="flex items-center gap-2 truncate flex-1 min-w-0">
                    <Server className="w-4 h-4 text-primary shrink-0" />
                    <span className="truncate" title={node.name}>{node.name}</span>
                  </div>
                  <Badge variant={node.status === 'Ready' ? 'success' : 'destructive'}>
                    {node.status}
                  </Badge>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {/* Roles */}
                  <div className="flex flex-wrap gap-1">
                    {node.roles?.map((role) => (
                      <Badge key={role} variant="outline" className="text-xs">
                        {role}
                      </Badge>
                    ))}
                  </div>

                  {/* CPU Usage */}
                  <div className="space-y-1">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground flex items-center gap-1">
                        <Cpu className="w-3.5 h-3.5" />
                        CPU
                      </span>
                      <span>
                        {formatCpu(node.cpu_usage)} / {formatCpu(node.cpu_capacity)}
                      </span>
                    </div>
                    {node.cpu_percentage !== undefined && (
                      <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
                        <div
                          className={`h-full rounded-full ${
                            node.cpu_percentage > 80 ? 'bg-red-500' : node.cpu_percentage > 60 ? 'bg-yellow-500' : 'bg-green-500'
                          }`}
                          style={{ width: `${node.cpu_percentage}%` }}
                        />
                      </div>
                    )}
                  </div>

                  {/* Memory Usage */}
                  <div className="space-y-1">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground flex items-center gap-1">
                        <HardDrive className="w-3.5 h-3.5" />
                        Memory
                      </span>
                      <span>
                        {formatMemory(node.memory_usage)} / {formatMemory(node.memory_capacity)}
                      </span>
                    </div>
                    {node.memory_percentage !== undefined && (
                      <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
                        <div
                          className={`h-full rounded-full ${
                            node.memory_percentage > 80 ? 'bg-red-500' : node.memory_percentage > 60 ? 'bg-yellow-500' : 'bg-green-500'
                          }`}
                          style={{ width: `${node.memory_percentage}%` }}
                        />
                      </div>
                    )}
                  </div>

                  {/* Info Row */}
                  <div className="flex items-center justify-between text-sm pt-2 border-t border-border">
                    <div className="flex items-center gap-1 text-muted-foreground">
                      <Container className="w-3.5 h-3.5" />
                      <span>{node.pod_count} pods</span>
                    </div>
                    <div className="flex items-center gap-1 text-muted-foreground">
                      <Clock className="w-3.5 h-3.5" />
                      <span>{formatAge(node.created_at)}</span>
                    </div>
                  </div>

                  {/* View Button */}
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full"
                    onClick={(e) => {
                      e.stopPropagation();
                      fetchNodeDetail(node.name);
                    }}
                  >
                    <Eye className="w-4 h-4 mr-1" />
                    View Details
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <Card>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left p-4 font-medium text-muted-foreground">Name</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Status</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Roles</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">CPU</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Memory</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Pods</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Age</th>
                  <th className="text-right p-4 font-medium text-muted-foreground">Actions</th>
                </tr>
              </thead>
              <tbody>
                {nodes.map((node) => (
                  <tr
                    key={node.name}
                    className="border-b border-border hover:bg-muted/50 transition-colors cursor-pointer"
                    onClick={() => fetchNodeDetail(node.name)}
                  >
                    <td className="p-4">
                      <div className="flex items-center gap-2">
                        <Server className="w-4 h-4 text-primary shrink-0" />
                        <span className="font-medium truncate max-w-[200px]" title={node.name}>
                          {node.name}
                        </span>
                      </div>
                    </td>
                    <td className="p-4">
                      <Badge variant={node.status === 'Ready' ? 'success' : 'destructive'}>
                        {node.status}
                      </Badge>
                    </td>
                    <td className="p-4">
                      <div className="flex flex-wrap gap-1">
                        {node.roles?.map((role) => (
                          <Badge key={role} variant="outline" className="text-xs">
                            {role}
                          </Badge>
                        ))}
                      </div>
                    </td>
                    <td className="p-4">
                      <div className="space-y-1">
                        <span className="text-sm">
                          {formatCpu(node.cpu_usage)} / {formatCpu(node.cpu_capacity)}
                        </span>
                        {node.cpu_percentage !== undefined && (
                          <div className="w-20 h-1.5 bg-muted rounded-full overflow-hidden">
                            <div
                              className={`h-full rounded-full ${
                                node.cpu_percentage > 80 ? 'bg-red-500' : node.cpu_percentage > 60 ? 'bg-yellow-500' : 'bg-green-500'
                              }`}
                              style={{ width: `${node.cpu_percentage}%` }}
                            />
                          </div>
                        )}
                      </div>
                    </td>
                    <td className="p-4">
                      <div className="space-y-1">
                        <span className="text-sm">
                          {formatMemory(node.memory_usage)} / {formatMemory(node.memory_capacity)}
                        </span>
                        {node.memory_percentage !== undefined && (
                          <div className="w-20 h-1.5 bg-muted rounded-full overflow-hidden">
                            <div
                              className={`h-full rounded-full ${
                                node.memory_percentage > 80 ? 'bg-red-500' : node.memory_percentage > 60 ? 'bg-yellow-500' : 'bg-green-500'
                              }`}
                              style={{ width: `${node.memory_percentage}%` }}
                            />
                          </div>
                        )}
                      </div>
                    </td>
                    <td className="p-4 text-sm">{node.pod_count}</td>
                    <td className="p-4 text-sm text-muted-foreground">{formatAge(node.created_at)}</td>
                    <td className="p-4 text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation();
                          fetchNodeDetail(node.name);
                        }}
                      >
                        <Eye className="w-4 h-4" />
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {(page - 1) * PAGE_SIZE + 1} to {Math.min(page * PAGE_SIZE, total)} of {total} nodes
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="icon"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              <ChevronLeft className="w-4 h-4" />
            </Button>
            <span className="text-sm">
              Page {page} of {totalPages}
            </span>
            <Button
              variant="outline"
              size="icon"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
            >
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}

      {/* Node Detail Dialog */}
      <Dialog open={viewDialogOpen} onOpenChange={setViewDialogOpen}>
        <DialogContent className="max-w-3xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Server className="w-5 h-5 text-primary" />
              {selectedNode?.name}
            </DialogTitle>
            <DialogDescription>Node details and resource information</DialogDescription>
          </DialogHeader>

          {loadingDetail ? (
            <Loading text="Loading node details..." />
          ) : selectedNode && (
            <Tabs defaultValue="overview" className="mt-4">
              <TabsList className="grid w-full grid-cols-4">
                <TabsTrigger value="overview">Overview</TabsTrigger>
                <TabsTrigger value="resources">Resources</TabsTrigger>
                <TabsTrigger value="conditions">Conditions</TabsTrigger>
                <TabsTrigger value="labels">Labels</TabsTrigger>
              </TabsList>

              <TabsContent value="overview" className="space-y-4 mt-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Status</p>
                    <Badge variant={selectedNode.status === 'Ready' ? 'success' : 'destructive'}>
                      {selectedNode.status}
                    </Badge>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Roles</p>
                    <div className="flex gap-1 flex-wrap">
                      {selectedNode.roles?.map((role) => (
                        <Badge key={role} variant="outline">{role}</Badge>
                      ))}
                    </div>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Internal IP</p>
                    <p className="font-mono text-sm">{selectedNode.internal_ip}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">External IP</p>
                    <p className="font-mono text-sm">{selectedNode.external_ip || '-'}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Kubelet Version</p>
                    <p className="text-sm">{selectedNode.kubelet_version}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Container Runtime</p>
                    <p className="text-sm">{selectedNode.container_runtime}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">OS Image</p>
                    <p className="text-sm">{selectedNode.os_image}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Architecture</p>
                    <p className="text-sm">{selectedNode.os} / {selectedNode.architecture}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Kernel Version</p>
                    <p className="font-mono text-sm">{selectedNode.kernel_version}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Pod Count</p>
                    <p className="text-sm">{selectedNode.pod_count}</p>
                  </div>
                  <div className="space-y-1 col-span-2">
                    <p className="text-sm text-muted-foreground">Created</p>
                    <p className="text-sm">{new Date(selectedNode.created_at).toLocaleString()}</p>
                  </div>
                </div>

                {/* Taints */}
                {selectedNode.taints && selectedNode.taints.length > 0 && (
                  <div className="pt-4 border-t border-border">
                    <p className="text-sm font-medium flex items-center gap-2 mb-2">
                      <Shield className="w-4 h-4" />
                      Taints
                    </p>
                    <div className="flex flex-wrap gap-2">
                      {selectedNode.taints.map((taint, i) => (
                        <Badge key={i} variant="secondary" className="font-mono text-xs">
                          {taint.key}={taint.value || ''}:{taint.effect}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}
              </TabsContent>

              <TabsContent value="resources" className="space-y-4 mt-4">
                {/* Usage */}
                {selectedNode.usage && (
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm flex items-center gap-2">
                        <Activity className="w-4 h-4" />
                        Current Usage
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="grid grid-cols-2 gap-4">
                      <div>
                        <p className="text-xs text-muted-foreground mb-1">CPU</p>
                        <p className="text-lg font-bold">{selectedNode.usage.cpu_percentage?.toFixed(1)}%</p>
                        <p className="text-xs text-muted-foreground">
                          {formatCpu(selectedNode.usage.cpu)} used
                        </p>
                      </div>
                      <div>
                        <p className="text-xs text-muted-foreground mb-1">Memory</p>
                        <p className="text-lg font-bold">{selectedNode.usage.memory_percentage?.toFixed(1)}%</p>
                        <p className="text-xs text-muted-foreground">
                          {formatMemory(selectedNode.usage.memory)} used
                        </p>
                      </div>
                    </CardContent>
                  </Card>
                )}

                {/* Capacity vs Allocatable */}
                <div className="grid grid-cols-2 gap-4">
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm">Capacity</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">CPU</span>
                        <span>{formatCpu(selectedNode.capacity?.cpu)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Memory</span>
                        <span>{formatMemory(selectedNode.capacity?.memory)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Pods</span>
                        <span>{selectedNode.capacity?.pods}</span>
                      </div>
                      {selectedNode.capacity?.ephemeral_storage && (
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Storage</span>
                          <span>{selectedNode.capacity.ephemeral_storage}</span>
                        </div>
                      )}
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm">Allocatable</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">CPU</span>
                        <span>{formatCpu(selectedNode.allocatable?.cpu)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Memory</span>
                        <span>{formatMemory(selectedNode.allocatable?.memory)}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Pods</span>
                        <span>{selectedNode.allocatable?.pods}</span>
                      </div>
                      {selectedNode.allocatable?.ephemeral_storage && (
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Storage</span>
                          <span>{selectedNode.allocatable.ephemeral_storage}</span>
                        </div>
                      )}
                    </CardContent>
                  </Card>
                </div>
              </TabsContent>

              <TabsContent value="conditions" className="space-y-3 mt-4">
                {selectedNode.conditions?.map((condition, i) => (
                  <Card key={i}>
                    <CardContent className="p-4">
                      <div className="flex items-start justify-between">
                        <div className="flex items-center gap-2">
                          {condition.status === 'True' && condition.type === 'Ready' ? (
                            <CheckCircle2 className="w-4 h-4 text-green-500" />
                          ) : condition.status === 'False' && condition.type !== 'Ready' ? (
                            <CheckCircle2 className="w-4 h-4 text-green-500" />
                          ) : (
                            <AlertTriangle className="w-4 h-4 text-yellow-500" />
                          )}
                          <span className="font-medium">{condition.type}</span>
                        </div>
                        <Badge variant={condition.status === 'True' ? (condition.type === 'Ready' ? 'success' : 'destructive') : 'secondary'}>
                          {condition.status}
                        </Badge>
                      </div>
                      <p className="text-sm text-muted-foreground mt-2">{condition.reason}</p>
                      <p className="text-xs text-muted-foreground mt-1">{condition.message}</p>
                      {condition.last_heartbeat_time && (
                        <p className="text-xs text-muted-foreground mt-2">
                          Last heartbeat: {new Date(condition.last_heartbeat_time).toLocaleString()}
                        </p>
                      )}
                    </CardContent>
                  </Card>
                )) || (
                  <div className="text-center py-8 text-muted-foreground">
                    <Activity className="w-12 h-12 mx-auto mb-4 opacity-50" />
                    <p>No conditions available</p>
                  </div>
                )}
              </TabsContent>

              <TabsContent value="labels" className="mt-4">
                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm flex items-center gap-2">
                      <Tag className="w-4 h-4" />
                      Labels ({Object.keys(selectedNode.labels || {}).length})
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    {Object.keys(selectedNode.labels || {}).length > 0 ? (
                      <div className="space-y-2">
                        {Object.entries(selectedNode.labels || {}).map(([key, value]) => (
                          <div key={key} className="flex items-center gap-2 text-sm">
                            <span className="font-mono text-xs bg-muted px-2 py-1 rounded">{key}</span>
                            <span className="text-muted-foreground">=</span>
                            <span className="font-mono text-xs">{value || '""'}</span>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <p className="text-muted-foreground text-center py-4">No labels</p>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
