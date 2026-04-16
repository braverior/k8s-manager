import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { clusterApi, deploymentApi } from '@/api';
import { useCluster } from '@/hooks/use-cluster';
import { useToast } from '@/hooks/use-toast';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Loading } from '@/components/ui/spinner';
import { formatMemory, formatCpu, cn } from '@/lib/utils';
import type { ClusterDashboard, Deployment } from '@/types';
import {
  Server,
  Box,
  Settings,
  Layers,
  Activity,
  ArrowRight,
  Container,
  Network,
  RefreshCw,
  CheckCircle2,
  XCircle,
  Clock,
  AlertTriangle,
  Gauge,
  Star,
  RotateCw,
} from 'lucide-react';

function ProgressRing({ percentage, size = 120, strokeWidth = 10, color = 'text-primary' }: {
  percentage: number;
  size?: number;
  strokeWidth?: number;
  color?: string;
}) {
  const radius = (size - strokeWidth) / 2;
  const circumference = radius * 2 * Math.PI;
  const offset = circumference - (percentage / 100) * circumference;

  return (
    <div className="relative inline-flex items-center justify-center">
      <svg width={size} height={size} className="-rotate-90">
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="currentColor"
          strokeWidth={strokeWidth}
          className="text-muted/30"
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="currentColor"
          strokeWidth={strokeWidth}
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          strokeLinecap="round"
          className={color}
        />
      </svg>
      <div className="absolute flex flex-col items-center justify-center">
        <span className="text-2xl font-bold">{percentage.toFixed(1)}%</span>
      </div>
    </div>
  );
}

function StatCard({ title, value, icon: Icon, subValue, color = 'text-primary' }: {
  title: string;
  value: number | string;
  icon: React.ElementType;
  subValue?: string;
  color?: string;
}) {
  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-muted-foreground">{title}</p>
            <p className="text-2xl font-bold mt-1">{value}</p>
            {subValue && <p className="text-xs text-muted-foreground mt-1">{subValue}</p>}
          </div>
          <div className={`p-3 rounded-lg bg-muted ${color}`}>
            <Icon className="w-6 h-6" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function DashboardPage() {
  const { selectedCluster, selectedNamespace, clusters, defaultCluster, setDefaultCluster } = useCluster();
  const { toast } = useToast();

  const [dashboard, setDashboard] = useState<ClusterDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const [rollingDeployments, setRollingDeployments] = useState<Deployment[]>([]);
  const [allDeployments, setAllDeployments] = useState<Deployment[]>([]);

  const fetchDashboard = useCallback(async () => {
    if (!selectedCluster) return;
    try {
      setLoading(true);
      const data = await clusterApi.getDashboard(selectedCluster, selectedNamespace || undefined);
      setDashboard(data);
    } catch (err) {
      setDashboard(null);
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to fetch dashboard',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [selectedCluster, selectedNamespace, toast]);

  useEffect(() => {
    fetchDashboard();
  }, [fetchDashboard]);

  // 轮询 Deployment 滚动更新状态，每 5s 刷新一次
  useEffect(() => {
    if (!selectedCluster || !selectedNamespace) return;

    const fetchDeployments = async () => {
      try {
        const list = await deploymentApi.list(selectedCluster, selectedNamespace);
        setAllDeployments(list);
        setRollingDeployments(
          list.filter(d => d.updated_replicas < d.replicas)
        );
      } catch {
        // 静默失败，不影响主 dashboard
      }
    };

    fetchDeployments();
    const interval = setInterval(fetchDeployments, 5000);
    return () => clearInterval(interval);
  }, [selectedCluster, selectedNamespace]);

  const clusterInfo = clusters.find((c) => c.name === selectedCluster);

  const quickLinks = [
    { title: 'Nodes', icon: Server, href: '/nodes', color: 'text-blue-500' },
    { title: 'ConfigMaps', icon: Settings, href: '/configmaps', color: 'text-cyan-500' },
    { title: 'Deployments', icon: Box, href: '/deployments', color: 'text-green-500' },
    { title: 'Pods', icon: Container, href: '/pods', color: 'text-purple-500' },
  ];

  if (!selectedCluster) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-muted-foreground">Please select a cluster</p>
      </div>
    );
  }

  if (loading) {
    return <Loading text="Loading dashboard..." />;
  }

  if (!dashboard) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-muted-foreground">Failed to load dashboard data</p>
      </div>
    );
  }

  const podStats = dashboard.resources?.pods || { total: 0, running: 0, pending: 0, succeeded: 0, failed: 0 };
  const podTotal = podStats.total || 1;
  const capacity = dashboard.capacity || { cpu: '-', memory: '-', pods: 0 };
  const resources = dashboard.resources || { nodes: 0, namespaces: 0, deployments: 0, services: 0, configmaps: 0, hpas: 0, pods: podStats };

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold flex items-center gap-2">
            <Layers className="w-7 h-7 text-primary" />
            {dashboard.cluster_name}
          </h1>
          {clusterInfo?.description && (
            <p className="text-sm text-muted-foreground mt-1">{clusterInfo.description}</p>
          )}
          <div className="flex items-center gap-2 mt-2">
            <Badge variant={dashboard.status === 'connected' ? 'success' : 'destructive'}>
              {dashboard.status}
            </Badge>
            <Badge variant="outline">
              {dashboard.version}
            </Badge>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant={defaultCluster === selectedCluster ? 'default' : 'outline'}
            size="sm"
            onClick={() => {
              if (defaultCluster === selectedCluster) {
                setDefaultCluster('');
                toast({
                  title: 'Default cluster cleared',
                  description: 'The first available cluster will be auto-selected on page load.',
                });
              } else {
                setDefaultCluster(selectedCluster);
                toast({
                  title: 'Default cluster set',
                  description: `"${selectedCluster}" will be auto-selected on page load.`,
                });
              }
            }}
            title={
              defaultCluster === selectedCluster
                ? 'Click to clear default cluster'
                : 'Set as default cluster'
            }
          >
            <Star
              className={cn(
                'w-4 h-4 mr-1.5',
                defaultCluster === selectedCluster ? 'fill-current' : ''
              )}
            />
            {defaultCluster === selectedCluster ? 'Default' : 'Set as Default'}
          </Button>
          <Button variant="outline" size="icon" onClick={fetchDashboard}>
            <RefreshCw className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Resource Stats */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Nodes"
          value={resources.nodes}
          icon={Server}
          color="text-blue-500"
        />
        <StatCard
          title="Namespaces"
          value={resources.namespaces}
          icon={Layers}
          color="text-purple-500"
        />
        <StatCard
          title="Deployments"
          value={resources.deployments}
          icon={Box}
          color="text-green-500"
        />
        <StatCard
          title="Services"
          value={resources.services}
          icon={Network}
          color="text-orange-500"
        />
      </div>

      {/* Usage & Pod Stats */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* CPU & Memory Usage */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Gauge className="w-4 h-4" />
              Resource Usage
            </CardTitle>
          </CardHeader>
          <CardContent>
            {dashboard.usage ? (
              <div className="flex justify-around items-center">
                <div className="text-center">
                  <ProgressRing
                    percentage={dashboard.usage.cpu_percentage || 0}
                    color={(dashboard.usage.cpu_percentage || 0) > 80 ? 'text-red-500' : (dashboard.usage.cpu_percentage || 0) > 60 ? 'text-yellow-500' : 'text-green-500'}
                  />
                  <p className="text-sm text-muted-foreground mt-2">CPU</p>
                  <p className="text-xs font-mono">{formatCpu(dashboard.usage.cpu)} / {formatCpu(capacity.cpu)}</p>
                </div>
                <div className="text-center">
                  <ProgressRing
                    percentage={dashboard.usage.memory_percentage || 0}
                    color={(dashboard.usage.memory_percentage || 0) > 80 ? 'text-red-500' : (dashboard.usage.memory_percentage || 0) > 60 ? 'text-yellow-500' : 'text-green-500'}
                  />
                  <p className="text-sm text-muted-foreground mt-2">Memory</p>
                  <p className="text-xs font-mono">{formatMemory(dashboard.usage.memory)} / {formatMemory(capacity.memory)}</p>
                </div>
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
                <AlertTriangle className="w-8 h-8 mb-2" />
                <p className="text-sm">Metrics server not available</p>
                <p className="text-xs mt-1">Capacity: {formatCpu(capacity.cpu)} CPU, {formatMemory(capacity.memory)}</p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Pod Status */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Container className="w-4 h-4" />
              Pod Status
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="text-center">
                <p className="text-4xl font-bold">{podStats.total}</p>
                <p className="text-sm text-muted-foreground">Total Pods</p>
              </div>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <CheckCircle2 className="w-4 h-4 text-green-500" />
                    <span className="text-sm">Running</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{podStats.running}</span>
                    <div className="w-24 h-2 bg-muted rounded-full overflow-hidden">
                      <div
                        className="h-full bg-green-500 rounded-full"
                        style={{ width: `${(podStats.running / podTotal) * 100}%` }}
                      />
                    </div>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Clock className="w-4 h-4 text-yellow-500" />
                    <span className="text-sm">Pending</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{podStats.pending}</span>
                    <div className="w-24 h-2 bg-muted rounded-full overflow-hidden">
                      <div
                        className="h-full bg-yellow-500 rounded-full"
                        style={{ width: `${(podStats.pending / podTotal) * 100}%` }}
                      />
                    </div>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Activity className="w-4 h-4 text-blue-500" />
                    <span className="text-sm">Succeeded</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{podStats.succeeded}</span>
                    <div className="w-24 h-2 bg-muted rounded-full overflow-hidden">
                      <div
                        className="h-full bg-blue-500 rounded-full"
                        style={{ width: `${(podStats.succeeded / podTotal) * 100}%` }}
                      />
                    </div>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <XCircle className="w-4 h-4 text-red-500" />
                    <span className="text-sm">Failed</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{podStats.failed}</span>
                    <div className="w-24 h-2 bg-muted rounded-full overflow-hidden">
                      <div
                        className="h-full bg-red-500 rounded-full"
                        style={{ width: `${(podStats.failed / podTotal) * 100}%` }}
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Quick Access */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="text-base">Quick Access</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {quickLinks.map((link) => {
              const Icon = link.icon;
              return (
                <Link key={link.href} to={link.href}>
                  <div className="flex items-center justify-between p-3 rounded-lg bg-muted/50 hover:bg-muted transition-colors cursor-pointer group">
                    <div className="flex items-center gap-3">
                      <Icon className={`w-5 h-5 ${link.color}`} />
                      <span className="font-medium">{link.title}</span>
                    </div>
                    <ArrowRight className="w-4 h-4 text-muted-foreground group-hover:text-foreground transition-colors" />
                  </div>
                </Link>
              );
            })}
            <div className="grid grid-cols-2 gap-3 pt-2">
              <div className="p-3 rounded-lg bg-muted/50 text-center">
                <p className="text-2xl font-bold">{resources.configmaps}</p>
                <p className="text-xs text-muted-foreground">ConfigMaps</p>
              </div>
              <Link to="/hpas">
                <div className="p-3 rounded-lg bg-muted/50 text-center hover:bg-muted transition-colors cursor-pointer">
                  <p className="text-2xl font-bold">{resources.hpas}</p>
                  <p className="text-xs text-muted-foreground">HPAs</p>
                </div>
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Rolling Updates */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base flex items-center gap-2">
            <RotateCw className={cn('w-4 h-4', rollingDeployments.length > 0 && 'animate-spin text-blue-500')} />
            滚动更新状态
            {rollingDeployments.length > 0 && (
              <Badge variant="outline" className="ml-auto text-blue-500 border-blue-500 animate-pulse">
                {rollingDeployments.length} 滚动中
              </Badge>
            )}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {allDeployments.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-4">暂无 Deployment 数据</p>
          ) : rollingDeployments.length === 0 ? (
            <div className="flex items-center gap-2 text-sm text-green-600 py-2">
              <CheckCircle2 className="w-4 h-4" />
              所有 Deployment 均已更新完成
            </div>
          ) : (
            <div className="space-y-3">
              {rollingDeployments.map(d => {
                const updatedPct = d.replicas > 0 ? (d.updated_replicas / d.replicas) * 100 : 0;
                const readyPct = d.replicas > 0 ? (d.ready_replicas / d.replicas) * 100 : 0;
                return (
                  <div key={`${d.namespace}/${d.name}`} className="rounded-lg border p-3 space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-sm">{d.name}</span>
                      <span className="text-xs text-muted-foreground">{d.namespace}</span>
                    </div>
                    <div className="grid grid-cols-2 gap-3 text-xs">
                      <div className="space-y-1">
                        <div className="flex justify-between text-muted-foreground">
                          <span>UP-TO-DATE</span>
                          <span className="font-mono font-medium text-foreground">{d.updated_replicas}/{d.replicas}</span>
                        </div>
                        <div className="h-2 bg-muted rounded-full overflow-hidden">
                          <div
                            className={cn('h-full rounded-full transition-all duration-500', updatedPct === 100 ? 'bg-green-500' : 'bg-blue-500')}
                            style={{ width: `${updatedPct}%` }}
                          />
                        </div>
                      </div>
                      <div className="space-y-1">
                        <div className="flex justify-between text-muted-foreground">
                          <span>READY</span>
                          <span className="font-mono font-medium text-foreground">{d.ready_replicas}/{d.replicas}</span>
                        </div>
                        <div className="h-2 bg-muted rounded-full overflow-hidden">
                          <div
                            className={cn('h-full rounded-full transition-all duration-500', readyPct === 100 ? 'bg-green-500' : 'bg-yellow-500')}
                            style={{ width: `${readyPct}%` }}
                          />
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
