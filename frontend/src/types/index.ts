// API Types
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface PaginatedResponse<T> {
  total: number;
  items: T[];
}

// Auth Types
export interface Department {
  id: string;
  name: string;
  path: string;
}

export interface ClusterPermission {
  cluster: string;
  namespaces: string[];
}

export interface User {
  id: string;
  name: string;
  email: string;
  avatar_url: string;
  mobile?: string;
  employee_id?: string;
  department?: Department;
  role: 'admin' | 'user';
  is_admin: boolean;
  permissions?: ClusterPermission[];
  status?: 'active' | 'disabled';
  last_login_at?: string;
  created_at?: string;
}

export interface FeishuConfig {
  app_id: string;
  redirect_uri: string;
  authorize_url: string;
}

export interface LoginResponse {
  token: string;
  expires_in: number;
  user: User;
}

export interface UserPermissions {
  user_id: string;
  user_name: string;
  is_admin: boolean;
  permissions: ClusterPermission[];
}

export interface BatchPermissionResult {
  success_count: number;
  failed_count: number;
  failed_users: string[];
}

// Cluster Types
export interface Cluster {
  name: string;
  description: string;
  api_server: string;
  status: 'connected' | 'disconnected' | 'unknown';
}

export interface ClusterDetail {
  id: number;
  name: string;
  description: string;
  cluster_type: string;
  api_server: string;
  status: string;
  source: string;
  has_kubeconfig: boolean;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface AddClusterRequest {
  name: string;
  description: string;
  kubeconfig: string;
}

export interface UpdateClusterRequest {
  description?: string;
  kubeconfig?: string;
}

export interface TestNewConnectionRequest {
  kubeconfig: string;
}

export interface Namespace {
  name: string;
  status: string;
}

// Resource Types
export interface K8sResource {
  name: string;
  namespace: string;
  yaml: string;
}

// Deployment Type (extends K8sResource with status fields)
export interface Deployment extends K8sResource {
  replicas: number;
  ready_replicas: number;
  updated_replicas: number;
  available_replicas: number;
  pod_count: number;
  pod_status_counts?: Record<string, number>;
}

export interface ResourceRequest {
  yaml?: string;
  content?: string;
}

// History Types
export interface HistoryRecord {
  id: number;
  cluster_name: string;
  namespace: string;
  resource_type: 'ConfigMap' | 'Deployment' | 'Service' | 'HPA';
  resource_name: string;
  version: number;
  operation: 'create' | 'update' | 'delete' | 'restart' | string;
  operator: string;
  created_at: string;
  content?: string;
}

export interface HistoryDiff {
  source_version: number;
  target_version: number;
  source_content: string;
  target_content: string;
  diff: string;
}

// Connection Test
export interface ConnectionTest {
  success: boolean;
  message: string;
  version?: string;
}

// Pod Types
export interface PodContainer {
  name: string;
  ready: boolean;
  restart_count: number;
  state: string;
  reason?: string;
  message?: string;
  image: string;
  started_at?: string;
  exit_code?: number | null;
  last_state?: string;
  last_reason?: string;
  last_message?: string;
}

export interface PodContainerMetrics {
  name: string;
  cpu: string;
  memory: string;
}

export interface PodMetrics {
  containers: PodContainerMetrics[];
}

export interface PodCondition {
  type: string;
  status: string;
  reason?: string;
  message?: string;
  last_transition_time?: string;
}

export interface Pod {
  name: string;
  namespace: string;
  phase: 'Pending' | 'Running' | 'Succeeded' | 'Failed' | 'Unknown';
  pod_ip: string;
  host_ip: string;
  node_name: string;
  ready_containers: number;
  total_containers: number;
  restart_count: number;
  created_at: string;
  containers: PodContainer[];
  conditions?: PodCondition[];
  metrics?: PodMetrics;
}

export interface PodLogResponse {
  pod_name: string;
  container_name: string;
  logs: string;
}

export interface PodEvent {
  type: string;
  reason: string;
  message: string;
  source: string;
  count: number;
  first_time: string;
  last_time: string;
}

// Dashboard Types
export interface DashboardPodStats {
  total: number;
  running: number;
  pending: number;
  succeeded: number;
  failed: number;
}

export interface DashboardResources {
  nodes: number;
  namespaces: number;
  pods: DashboardPodStats;
  deployments: number;
  services: number;
  configmaps: number;
  hpas: number;
}

export interface DashboardCapacity {
  cpu: string;
  memory: string;
  pods: number;
}

export interface DashboardUsage {
  cpu: string;
  cpu_percentage: number;
  memory: string;
  memory_percentage: number;
}

export interface DashboardNode {
  name: string;
  status: string;
  roles: string[];
  internal_ip: string;
  kubelet_version: string;
  os: string;
  container_runtime: string;
  cpu_capacity: string;
  memory_capacity: string;
  cpu_usage?: string;
  memory_usage?: string;
}

export interface ClusterDashboard {
  cluster_name: string;
  status: string;
  version: string;
  resources: DashboardResources;
  capacity: DashboardCapacity;
  usage?: DashboardUsage;
  nodes: DashboardNode[];
  hpa_summaries?: HPASummary[];
}

// HPA Summary (for Dashboard)
export interface HPASummary {
  name: string;
  namespace: string;
  target_kind: string;
  target_name: string;
  min_replicas: number;
  max_replicas: number;
  current_replicas: number;
  desired_replicas: number;
  cpu_target_utilization?: number;
  cpu_current_utilization?: number;
  memory_target_utilization?: number;
  memory_current_utilization?: number;
}

// HPA Types
export interface HPAScaleTargetRef {
  api_version: string;
  kind: string;
  name: string;
}

export interface HPAMetric {
  type: string;
  name: string;
  target_type: string;
  target_value?: string;
  target_utilization?: number;
  current_value?: string;
  current_utilization?: number;
}

export interface HPACondition {
  type: string;
  status: string;
  reason: string;
  message: string;
  last_transition_time: string;
}

export interface HPAResourceMetric {
  target_type: string;
  target_utilization?: number;
  target_average_value?: string;
  current_utilization?: number;
  current_average_value?: string;
}

export interface HPA {
  name: string;
  namespace: string;
  scale_target_ref: HPAScaleTargetRef;
  min_replicas: number;
  max_replicas: number;
  current_replicas: number;
  desired_replicas: number;
  cpu?: HPAResourceMetric;
  memory?: HPAResourceMetric;
  metrics?: HPAMetric[];
  conditions?: HPACondition[];
  yaml: string;
}

// Node Types (for Node list API)
export interface NodeListItem {
  name: string;
  status: string;
  roles: string[];
  internal_ip: string;
  kubelet_version: string;
  cpu_capacity: string;
  memory_capacity: string;
  cpu_usage?: string;
  cpu_percentage?: number;
  memory_usage?: string;
  memory_percentage?: number;
  pod_count: number;
  created_at: string;
}

export interface NodeCondition {
  type: string;
  status: string;
  reason: string;
  message: string;
  last_heartbeat_time?: string;
}

export interface NodeTaint {
  key: string;
  value?: string;
  effect: string;
}

export interface NodeCapacity {
  cpu: string;
  memory: string;
  pods: string;
  ephemeral_storage?: string;
}

export interface NodeUsage {
  cpu: string;
  cpu_percentage: number;
  memory: string;
  memory_percentage: number;
}

export interface NodeDetail {
  name: string;
  status: string;
  roles: string[];
  internal_ip: string;
  external_ip?: string;
  kubelet_version: string;
  kernel_version: string;
  os_image: string;
  os: string;
  architecture: string;
  container_runtime: string;
  created_at: string;
  capacity: NodeCapacity;
  allocatable: NodeCapacity;
  usage?: NodeUsage;
  conditions: NodeCondition[];
  taints?: NodeTaint[];
  labels: Record<string, string>;
  pod_count: number;
}

// WebSocket Terminal Message Types
export interface TerminalInputMessage {
  type: 'input';
  data: string;
}

export interface TerminalOutputMessage {
  type: 'output';
  data: string;
}

export interface TerminalResizeMessage {
  type: 'resize';
  cols: number;
  rows: number;
}

export interface TerminalPingMessage {
  type: 'ping';
}

export interface TerminalPongMessage {
  type: 'pong';
}

export interface TerminalErrorMessage {
  type: 'error';
  data: string;
}

export type TerminalMessage =
  | TerminalInputMessage
  | TerminalOutputMessage
  | TerminalResizeMessage
  | TerminalPingMessage
  | TerminalPongMessage
  | TerminalErrorMessage;
