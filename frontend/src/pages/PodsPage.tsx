import { useState, useEffect, useCallback, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import { podApi, deploymentApi } from '@/api';
import { useCluster } from '@/hooks/use-cluster';
import { useToast } from '@/hooks/use-toast';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Loading } from '@/components/ui/spinner';
import { DeleteConfirmDialog } from '@/components/DeleteConfirmDialog';
import { PodTerminalDialog } from '@/components/PodTerminalDialog';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import type { Pod, PodLogResponse, PodEvent, Deployment, PaginatedResponse } from '@/types';
import {
  Search,
  RefreshCw,
  Trash2,
  Eye,
  Server,
  Cpu,
  HardDrive,
  Clock,
  Container,
  Activity,
  CircleDot,
  RotateCcw,
  LayoutGrid,
  List,
  Box,
  Terminal,
  FileText,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Info,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';

type ViewMode = 'card' | 'table';

const PAGE_SIZE = 50;

export function PodsPage() {
  const { selectedCluster, selectedNamespace } = useCluster();
  const { toast } = useToast();
  const [searchParams, setSearchParams] = useSearchParams();
  const deploymentFilter = searchParams.get('deployment');

  const [pods, setPods] = useState<Pod[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [viewMode, setViewMode] = useState<ViewMode>('card');
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [deploymentsLoading, setDeploymentsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  // Debounce timer ref
  const searchTimerRef = useRef<ReturnType<typeof setTimeout>>();

  // Dialog states
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [terminalDialogOpen, setTerminalDialogOpen] = useState(false);
  const [selectedPod, setSelectedPod] = useState<Pod | null>(null);
  const [deleting, setDeleting] = useState(false);

  // Logs & Events states
  const [podLogs, setPodLogs] = useState<PodLogResponse | null>(null);
  const [logsLoading, setLogsLoading] = useState(false);
  const [logsPrevious, setLogsPrevious] = useState(false);
  const [selectedLogContainer, setSelectedLogContainer] = useState('');
  const [podEvents, setPodEvents] = useState<PodEvent[]>([]);
  const [eventsLoading, setEventsLoading] = useState(false);

  const fetchPods = useCallback(async () => {
    if (!selectedCluster || !selectedNamespace) return;
    try {
      setLoading(true);
      if (deploymentFilter) {
        // Non-paginated path for deployment filter
        const data = await podApi.list(selectedCluster, selectedNamespace, { deployment: deploymentFilter }) as Pod[];
        setPods(data || []);
        setTotal(data?.length || 0);
      } else {
        // Paginated path
        const data = await podApi.list(selectedCluster, selectedNamespace, {
          search: searchTerm || undefined,
          page,
          page_size: PAGE_SIZE,
        }) as PaginatedResponse<Pod>;
        setPods(data?.items || []);
        setTotal(data?.total || 0);
      }
    } catch (err) {
      setPods([]);
      setTotal(0);
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch Pods',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [selectedCluster, selectedNamespace, deploymentFilter, searchTerm, page, toast]);

  useEffect(() => {
    fetchPods();
  }, [fetchPods]);

  // Reset page on cluster/namespace/deployment change
  useEffect(() => {
    setPage(1);
    setSearchTerm('');
  }, [selectedCluster, selectedNamespace, deploymentFilter]);

  const fetchDeployments = useCallback(async () => {
    if (!selectedCluster || !selectedNamespace) return;
    try {
      setDeploymentsLoading(true);
      const data = await deploymentApi.list(selectedCluster, selectedNamespace);
      setDeployments(data || []);
    } catch {
      setDeployments([]);
    } finally {
      setDeploymentsLoading(false);
    }
  }, [selectedCluster, selectedNamespace]);

  useEffect(() => {
    fetchDeployments();
  }, [fetchDeployments]);

  const clearDeploymentFilter = () => {
    searchParams.delete('deployment');
    setSearchParams(searchParams);
  };

  const setDeploymentFilterParam = (name: string) => {
    searchParams.set('deployment', name);
    setSearchParams(searchParams);
  };

  const handleSearchChange = (value: string) => {
    setSearchTerm(value);
    if (searchTimerRef.current) {
      clearTimeout(searchTimerRef.current);
    }
    searchTimerRef.current = setTimeout(() => {
      setPage(1);
    }, 300);
  };

  const handleDelete = async () => {
    if (!selectedPod) return;
    try {
      setDeleting(true);
      await podApi.delete(selectedCluster, selectedNamespace, selectedPod.name);
      toast({ title: 'Success', description: 'Pod deleted successfully. A new pod will be scheduled.' });
      setDeleteDialogOpen(false);
      fetchPods();
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to delete Pod',
        variant: 'destructive',
      });
    } finally {
      setDeleting(false);
    }
  };

  const totalPages = Math.ceil(total / PAGE_SIZE);

  const fetchLogs = useCallback(async (podName: string, container?: string, previous?: boolean) => {
    if (!selectedCluster || !selectedNamespace) return;
    try {
      setLogsLoading(true);
      const data = await podApi.getLogs(selectedCluster, selectedNamespace, podName, {
        container: container || undefined,
        tail_lines: 500,
        previous: previous || false,
      });
      setPodLogs(data);
    } catch (err) {
      setPodLogs(null);
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch logs',
        variant: 'destructive',
      });
    } finally {
      setLogsLoading(false);
    }
  }, [selectedCluster, selectedNamespace, toast]);

  const fetchEvents = useCallback(async (podName: string) => {
    if (!selectedCluster || !selectedNamespace) return;
    try {
      setEventsLoading(true);
      const data = await podApi.getEvents(selectedCluster, selectedNamespace, podName);
      setPodEvents(data || []);
    } catch (err) {
      setPodEvents([]);
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch events',
        variant: 'destructive',
      });
    } finally {
      setEventsLoading(false);
    }
  }, [selectedCluster, selectedNamespace, toast]);

  const handleTabChange = (value: string) => {
    if (!selectedPod) return;
    if (value === 'logs') {
      setLogsPrevious(false);
      setSelectedLogContainer('');
      fetchLogs(selectedPod.name);
    } else if (value === 'events') {
      fetchEvents(selectedPod.name);
    }
  };

  const getPhaseVariant = (phase: string) => {
    switch (phase) {
      case 'Running':
        return 'success';
      case 'Succeeded':
        return 'secondary';
      case 'Pending':
        return 'warning';
      case 'Failed':
        return 'destructive';
      default:
        return 'outline';
    }
  };

  const getPhaseIcon = (phase: string) => {
    switch (phase) {
      case 'Running':
        return <Activity className="w-3 h-3" />;
      case 'Pending':
        return <Clock className="w-3 h-3" />;
      default:
        return <CircleDot className="w-3 h-3" />;
    }
  };

  const formatAge = (dateStr: string) => {
    const created = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - created.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffDays > 0) return `${diffDays}d`;
    if (diffHours > 0) return `${diffHours}h`;
    return `${diffMins}m`;
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
            <Container className="w-7 h-7 text-primary" />
            Pods
          </h1>
          <p className="text-muted-foreground mt-1">
            View and manage pods in {selectedNamespace}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="text-sm">
            {total} pods
          </Badge>
        </div>
      </div>

      {/* Deployment Filter Tags */}
      {!deploymentsLoading && deployments.length > 0 && (
        <div className="flex items-center gap-2 flex-wrap">
          <Box className="w-4 h-4 text-muted-foreground shrink-0" />
          <Badge
            variant={deploymentFilter ? 'outline' : 'default'}
            className="cursor-pointer hover:opacity-80"
            onClick={clearDeploymentFilter}
          >
            全部({deployments.reduce((sum, d) => sum + d.pod_count, 0)})
          </Badge>
          {deployments.map((dep) => (
            <Badge
              key={dep.name}
              variant={deploymentFilter === dep.name ? 'default' : 'outline'}
              className="cursor-pointer hover:opacity-80"
              onClick={() => setDeploymentFilterParam(dep.name)}
            >
              {dep.name}({dep.pod_count})
            </Badge>
          ))}
        </div>
      )}

      {/* Search & Controls */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search Pods..."
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
        <Button variant="outline" size="icon" onClick={fetchPods}>
          <RefreshCw className="w-4 h-4" />
        </Button>
      </div>

      {/* Pod List */}
      {loading ? (
        <Loading text="Loading Pods..." />
      ) : pods.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Container className="w-12 h-12 text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">
              {searchTerm ? 'No Pods found matching your search' : 'No Pods in this namespace'}
            </p>
          </CardContent>
        </Card>
      ) : viewMode === 'card' ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {pods.map((pod) => (
            <Card
              key={pod.name}
              className="group hover:border-primary/50 transition-colors cursor-pointer"
              onClick={() => {
                setSelectedPod(pod);
                setViewDialogOpen(true);
              }}
            >
              <CardHeader className="pb-3">
                <CardTitle className="flex items-center justify-between text-base">
                  <div className="flex items-center gap-2 truncate flex-1 min-w-0">
                    <Container className="w-4 h-4 text-primary shrink-0" />
                    <span className="truncate" title={pod.name}>{pod.name}</span>
                  </div>
                  <div className="flex items-center gap-1 shrink-0">
                    <Badge variant={getPhaseVariant(pod.phase)} className="flex items-center gap-1">
                      {getPhaseIcon(pod.phase)}
                      {pod.phase}
                    </Badge>
                  </div>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {/* Ready Status */}
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">Ready</span>
                    <span className={pod.ready_containers === pod.total_containers ? 'text-green-500' : 'text-yellow-500'}>
                      {pod.ready_containers}/{pod.total_containers}
                    </span>
                  </div>

                  {/* Restarts */}
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <RotateCcw className="w-3.5 h-3.5" />
                      Restarts
                    </span>
                    <span className={pod.restart_count > 0 ? 'text-yellow-500' : ''}>
                      {pod.restart_count}
                    </span>
                  </div>

                  {/* Node */}
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <Server className="w-3.5 h-3.5" />
                      Node
                    </span>
                    <span className="font-mono text-xs truncate max-w-[120px]" title={pod.node_name}>
                      {pod.node_name || '-'}
                    </span>
                  </div>

                  {/* Age */}
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <Clock className="w-3.5 h-3.5" />
                      Age
                    </span>
                    <span>{formatAge(pod.created_at)}</span>
                  </div>

                  {/* IP */}
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">Pod IP</span>
                    <span className="font-mono text-xs">{pod.pod_ip || '-'}</span>
                  </div>

                  {/* Actions */}
                  <div className="flex items-center gap-2 pt-2 border-t border-border">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="flex-1"
                      onClick={(e) => {
                        e.stopPropagation();
                        setSelectedPod(pod);
                        setTerminalDialogOpen(true);
                      }}
                      disabled={pod.phase !== 'Running'}
                      title={pod.phase !== 'Running' ? 'Pod must be running' : 'Open terminal'}
                    >
                      <Terminal className="w-4 h-4 mr-1" />
                      Terminal
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="flex-1"
                      onClick={(e) => {
                        e.stopPropagation();
                        setSelectedPod(pod);
                        setViewDialogOpen(true);
                      }}
                    >
                      <Eye className="w-4 h-4 mr-1" />
                      Details
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-destructive hover:text-destructive"
                      onClick={(e) => {
                        e.stopPropagation();
                        setSelectedPod(pod);
                        setDeleteDialogOpen(true);
                      }}
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
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
                  <th className="text-left p-4 font-medium text-muted-foreground">Ready</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Restarts</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Node</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Pod IP</th>
                  <th className="text-left p-4 font-medium text-muted-foreground">Age</th>
                  <th className="text-right p-4 font-medium text-muted-foreground">Actions</th>
                </tr>
              </thead>
              <tbody>
                {pods.map((pod) => (
                  <tr
                    key={pod.name}
                    className="border-b border-border hover:bg-muted/50 transition-colors cursor-pointer"
                    onClick={() => {
                      setSelectedPod(pod);
                      setViewDialogOpen(true);
                    }}
                  >
                    <td className="p-4">
                      <div className="flex items-center gap-2">
                        <Container className="w-4 h-4 text-primary shrink-0" />
                        <span className="font-medium truncate max-w-[200px]" title={pod.name}>
                          {pod.name}
                        </span>
                      </div>
                    </td>
                    <td className="p-4">
                      <Badge variant={getPhaseVariant(pod.phase)} className="flex items-center gap-1 w-fit">
                        {getPhaseIcon(pod.phase)}
                        {pod.phase}
                      </Badge>
                    </td>
                    <td className="p-4">
                      <span className={pod.ready_containers === pod.total_containers ? 'text-green-500' : 'text-yellow-500'}>
                        {pod.ready_containers}/{pod.total_containers}
                      </span>
                    </td>
                    <td className="p-4">
                      <span className={pod.restart_count > 0 ? 'text-yellow-500' : ''}>
                        {pod.restart_count}
                      </span>
                    </td>
                    <td className="p-4">
                      <span className="font-mono text-xs truncate max-w-[120px] block" title={pod.node_name}>
                        {pod.node_name || '-'}
                      </span>
                    </td>
                    <td className="p-4">
                      <span className="font-mono text-xs">{pod.pod_ip || '-'}</span>
                    </td>
                    <td className="p-4 text-sm text-muted-foreground">{formatAge(pod.created_at)}</td>
                    <td className="p-4 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation();
                            setSelectedPod(pod);
                            setTerminalDialogOpen(true);
                          }}
                          disabled={pod.phase !== 'Running'}
                          title={pod.phase !== 'Running' ? 'Pod must be running' : 'Open terminal'}
                        >
                          <Terminal className="w-4 h-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation();
                            setSelectedPod(pod);
                            setViewDialogOpen(true);
                          }}
                        >
                          <Eye className="w-4 h-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-destructive hover:text-destructive"
                          onClick={(e) => {
                            e.stopPropagation();
                            setSelectedPod(pod);
                            setDeleteDialogOpen(true);
                          }}
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

      {/* Pagination */}
      {!deploymentFilter && totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {(page - 1) * PAGE_SIZE + 1} to {Math.min(page * PAGE_SIZE, total)} of {total} pods
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

      {/* View Dialog */}
      <Dialog open={viewDialogOpen} onOpenChange={(open) => {
        setViewDialogOpen(open);
        if (!open) {
          setPodLogs(null);
          setPodEvents([]);
        }
      }}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 min-w-0">
              <Container className="w-5 h-5 text-primary shrink-0" />
              <span className="truncate" title={selectedPod?.name}>{selectedPod?.name}</span>
            </DialogTitle>
            <DialogDescription>Pod details, logs, and events</DialogDescription>
          </DialogHeader>

          {selectedPod && (
            <Tabs defaultValue="overview" className="mt-4" onValueChange={handleTabChange}>
              <TabsList className="grid w-full grid-cols-5">
                <TabsTrigger value="overview">Overview</TabsTrigger>
                <TabsTrigger value="containers">Containers</TabsTrigger>
                <TabsTrigger value="logs">Logs</TabsTrigger>
                <TabsTrigger value="events">Events</TabsTrigger>
                <TabsTrigger value="metrics">Metrics</TabsTrigger>
              </TabsList>

              <TabsContent value="overview" className="space-y-4 mt-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Status</p>
                    <Badge variant={getPhaseVariant(selectedPod.phase)}>{selectedPod.phase}</Badge>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Ready</p>
                    <p className="font-medium">{selectedPod.ready_containers}/{selectedPod.total_containers}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Restarts</p>
                    <p className="font-medium">{selectedPod.restart_count}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Age</p>
                    <p className="font-medium">{formatAge(selectedPod.created_at)}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Pod IP</p>
                    <p className="font-mono text-sm">{selectedPod.pod_ip || '-'}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-sm text-muted-foreground">Host IP</p>
                    <p className="font-mono text-sm">{selectedPod.host_ip || '-'}</p>
                  </div>
                  <div className="space-y-1 col-span-2">
                    <p className="text-sm text-muted-foreground">Node</p>
                    <p className="font-mono text-sm">{selectedPod.node_name || '-'}</p>
                  </div>
                  <div className="space-y-1 col-span-2">
                    <p className="text-sm text-muted-foreground">Created</p>
                    <p className="text-sm">{new Date(selectedPod.created_at).toLocaleString()}</p>
                  </div>
                </div>

                {/* Conditions */}
                {selectedPod.conditions && selectedPod.conditions.length > 0 && (
                  <div className="space-y-2 pt-4 border-t">
                    <p className="text-sm font-medium">Conditions</p>
                    <div className="space-y-2">
                      {selectedPod.conditions.map((cond) => (
                        <div key={cond.type} className="flex items-start gap-2 text-sm">
                          {cond.status === 'True' ? (
                            <CheckCircle2 className="w-4 h-4 text-green-500 mt-0.5 shrink-0" />
                          ) : (
                            <XCircle className="w-4 h-4 text-red-500 mt-0.5 shrink-0" />
                          )}
                          <div className="min-w-0">
                            <span className="font-medium">{cond.type}</span>
                            {cond.reason && <span className="text-muted-foreground"> - {cond.reason}</span>}
                            {cond.message && (
                              <p className="text-muted-foreground text-xs mt-0.5">{cond.message}</p>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </TabsContent>

              <TabsContent value="containers" className="space-y-4 mt-4">
                {selectedPod.containers?.map((container) => (
                  <Card key={container.name}>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm flex items-center justify-between">
                        <span className="flex items-center gap-2">
                          <Container className="w-4 h-4" />
                          {container.name}
                        </span>
                        <Badge variant={container.ready ? 'success' : 'destructive'}>
                          {container.state}
                        </Badge>
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-2 text-sm">
                      <div className="flex justify-between gap-2">
                        <span className="text-muted-foreground shrink-0">Image</span>
                        <span className="font-mono text-xs truncate" title={container.image}>
                          {container.image}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Restarts</span>
                        <span>{container.restart_count}</span>
                      </div>
                      {container.started_at && (
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Started</span>
                          <span>{new Date(container.started_at).toLocaleString()}</span>
                        </div>
                      )}
                      {container.reason && (
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Reason</span>
                          <span className="text-yellow-500">{container.reason}</span>
                        </div>
                      )}
                      {container.message && (
                        <div className="pt-1">
                          <span className="text-muted-foreground text-xs">Message: </span>
                          <span className="text-xs text-red-500">{container.message}</span>
                        </div>
                      )}
                      {container.exit_code != null && (
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Exit Code</span>
                          <span className={container.exit_code !== 0 ? 'text-red-500' : ''}>{container.exit_code}</span>
                        </div>
                      )}
                      {container.last_state && (
                        <div className="pt-2 border-t border-border space-y-1">
                          <p className="text-xs text-muted-foreground font-medium">Last State</p>
                          <div className="flex justify-between">
                            <span className="text-muted-foreground text-xs">State</span>
                            <span className="text-xs">{container.last_state}</span>
                          </div>
                          {container.last_reason && (
                            <div className="flex justify-between">
                              <span className="text-muted-foreground text-xs">Reason</span>
                              <span className="text-xs text-yellow-500">{container.last_reason}</span>
                            </div>
                          )}
                          {container.last_message && (
                            <p className="text-xs text-red-500">{container.last_message}</p>
                          )}
                        </div>
                      )}
                    </CardContent>
                  </Card>
                )) || <p className="text-muted-foreground">No container information available</p>}
              </TabsContent>

              <TabsContent value="logs" className="space-y-3 mt-4">
                <div className="flex items-center gap-2 flex-wrap">
                  {selectedPod.containers && selectedPod.containers.length > 1 && (
                    <select
                      className="text-sm border rounded-md px-2 py-1 bg-background"
                      value={selectedLogContainer}
                      onChange={(e) => {
                        setSelectedLogContainer(e.target.value);
                        fetchLogs(selectedPod.name, e.target.value, logsPrevious);
                      }}
                    >
                      <option value="">Default container</option>
                      {selectedPod.containers.map((c) => (
                        <option key={c.name} value={c.name}>{c.name}</option>
                      ))}
                    </select>
                  )}
                  <label className="flex items-center gap-1.5 text-sm">
                    <input
                      type="checkbox"
                      checked={logsPrevious}
                      onChange={(e) => {
                        setLogsPrevious(e.target.checked);
                        fetchLogs(selectedPod.name, selectedLogContainer, e.target.checked);
                      }}
                      className="rounded"
                    />
                    Previous container
                  </label>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => fetchLogs(selectedPod.name, selectedLogContainer, logsPrevious)}
                  >
                    <RefreshCw className="w-3.5 h-3.5 mr-1" />
                    Refresh
                  </Button>
                </div>
                {logsLoading ? (
                  <Loading text="Loading logs..." />
                ) : podLogs?.logs ? (
                  <pre className="bg-muted p-3 rounded-lg text-xs font-mono overflow-auto max-h-[400px] whitespace-pre-wrap break-all">
                    {podLogs.logs}
                  </pre>
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    <FileText className="w-12 h-12 mx-auto mb-4 opacity-50" />
                    <p>No logs available</p>
                    <p className="text-xs mt-1">Container may not have started yet</p>
                  </div>
                )}
              </TabsContent>

              <TabsContent value="events" className="space-y-3 mt-4">
                <div className="flex items-center justify-between">
                  <p className="text-sm text-muted-foreground">{podEvents.length} events</p>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => fetchEvents(selectedPod.name)}
                  >
                    <RefreshCw className="w-3.5 h-3.5 mr-1" />
                    Refresh
                  </Button>
                </div>
                {eventsLoading ? (
                  <Loading text="Loading events..." />
                ) : podEvents.length > 0 ? (
                  <div className="space-y-2">
                    {podEvents.map((event, idx) => (
                      <div
                        key={`${event.reason}-${idx}`}
                        className={`p-3 rounded-lg border text-sm ${
                          event.type === 'Warning' ? 'border-yellow-500/30 bg-yellow-500/5' : 'border-border'
                        }`}
                      >
                        <div className="flex items-start gap-2">
                          {event.type === 'Warning' ? (
                            <AlertTriangle className="w-4 h-4 text-yellow-500 mt-0.5 shrink-0" />
                          ) : (
                            <Info className="w-4 h-4 text-blue-500 mt-0.5 shrink-0" />
                          )}
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center gap-2 flex-wrap">
                              <span className="font-medium">{event.reason}</span>
                              <Badge variant={event.type === 'Warning' ? 'warning' : 'secondary'} className="text-xs">
                                {event.type}
                              </Badge>
                              {event.count > 1 && (
                                <span className="text-xs text-muted-foreground">x{event.count}</span>
                              )}
                            </div>
                            <p className="text-muted-foreground mt-1 break-all">{event.message}</p>
                            <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
                              <span>{event.source}</span>
                              {event.last_time && (
                                <span>{new Date(event.last_time).toLocaleString()}</span>
                              )}
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    <Activity className="w-12 h-12 mx-auto mb-4 opacity-50" />
                    <p>No events found</p>
                    <p className="text-xs mt-1">Events are typically retained for 1 hour</p>
                  </div>
                )}
              </TabsContent>

              <TabsContent value="metrics" className="space-y-4 mt-4">
                {selectedPod.metrics?.containers?.map((metric) => (
                  <Card key={metric.name}>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm flex items-center gap-2">
                        <Container className="w-4 h-4" />
                        {metric.name}
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="grid grid-cols-2 gap-4">
                        <div className="flex items-center gap-3 p-3 bg-muted rounded-lg">
                          <Cpu className="w-5 h-5 text-blue-500" />
                          <div>
                            <p className="text-xs text-muted-foreground">CPU</p>
                            <p className="font-mono font-medium">{metric.cpu}</p>
                          </div>
                        </div>
                        <div className="flex items-center gap-3 p-3 bg-muted rounded-lg">
                          <HardDrive className="w-5 h-5 text-green-500" />
                          <div>
                            <p className="text-xs text-muted-foreground">Memory</p>
                            <p className="font-mono font-medium">{metric.memory}</p>
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                )) || (
                  <div className="text-center py-8 text-muted-foreground">
                    <Activity className="w-12 h-12 mx-auto mb-4 opacity-50" />
                    <p>No metrics available</p>
                    <p className="text-xs mt-1">Metrics server may not be installed</p>
                  </div>
                )}
              </TabsContent>
            </Tabs>
          )}
        </DialogContent>
      </Dialog>

      {/* Delete Dialog */}
      <DeleteConfirmDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="Restart Pod"
        description="This will delete the pod. If it's managed by a Deployment or ReplicaSet, a new pod will be automatically created."
        resourceName={selectedPod?.name || ''}
        onConfirm={handleDelete}
        deleting={deleting}
      />

      {/* Terminal Dialog */}
      <PodTerminalDialog
        open={terminalDialogOpen}
        onOpenChange={setTerminalDialogOpen}
        pod={selectedPod}
        cluster={selectedCluster}
        namespace={selectedNamespace}
      />
    </div>
  );
}
