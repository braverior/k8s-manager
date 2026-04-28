import type {
  ApiResponse,
  Cluster,
  Namespace,
  K8sResource,
  ResourceRequest,
  HistoryRecord,
  PaginatedResponse,
  HistoryDiff,
  ConnectionTest,
  Pod,
  PodLogResponse,
  PodEvent,
  ClusterDashboard,
  NodeListItem,
  NodeDetail,
  Deployment,
  HPA,
  FeishuConfig,
  LoginResponse,
  User,
  UserPermissions,
  ClusterPermission,
  BatchPermissionResult,
  ClusterDetail,
  AddClusterRequest,
  UpdateClusterRequest,
  TestNewConnectionRequest,
} from '@/types';

const API_PREFIX = '/api/v1';

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

// Get token from localStorage
function getAuthToken(): string | null {
  return localStorage.getItem('token');
}

// Build full URL: apiServer + /api/v1 + path
function buildUrl(apiServer: string, path: string): string {
  // Remove trailing slash from apiServer
  const base = apiServer.replace(/\/+$/, '');
  return `${base}${API_PREFIX}${path}`;
}

export async function fetchApi<T>(apiServer: string, url: string, options?: RequestInit): Promise<T> {
  const token = getAuthToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const fullUrl = buildUrl(apiServer, url);

  const response = await fetch(fullUrl, {
    ...options,
    headers,
  });

  // Handle 401 Unauthorized - redirect to login
  if (response.status === 401) {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    window.location.href = '/login';
    throw new Error('Unauthorized');
  }

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new ApiError(error.message || `HTTP ${response.status}`, response.status);
  }

  const data: ApiResponse<T> = await response.json();

  if (data.code !== 0) {
    throw new Error(data.message || 'API Error');
  }

  return data.data;
}

// Helper: create a fetchApi bound to same-origin (all clusters go through the same backend)
function clusterFetch<T>(_cluster: string, url: string, options?: RequestInit): Promise<T> {
  return fetchApi<T>('', url, options);
}

// Auth APIs - need a cluster's apiServer since auth is per-cluster
// Auth uses the first available cluster by default (or the selected one)
export const authApi = {
  getFeishuConfig: (apiServer: string) => fetchApi<FeishuConfig>(apiServer, '/auth/feishu/config'),

  feishuLogin: (apiServer: string, code: string, state?: string) =>
    fetchApi<LoginResponse>(apiServer, '/auth/feishu/login', {
      method: 'POST',
      body: JSON.stringify({ code, state }),
    }),

  getMe: (apiServer: string) => fetchApi<User>(apiServer, '/auth/me'),

  logout: (apiServer: string) =>
    fetchApi<void>(apiServer, '/auth/logout', {
      method: 'POST',
    }),
};

// Admin APIs - use specific cluster's apiServer
export const adminApi = {
  listUsers: (apiServer: string, params?: {
    keyword?: string;
    department_id?: string;
    role?: string;
    page?: number;
    page_size?: number;
  }) => {
    const searchParams = new URLSearchParams();
    if (params?.keyword) searchParams.set('keyword', params.keyword);
    if (params?.department_id) searchParams.set('department_id', params.department_id);
    if (params?.role) searchParams.set('role', params.role);
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.page_size) searchParams.set('page_size', params.page_size.toString());
    const query = searchParams.toString();
    return fetchApi<PaginatedResponse<User>>(apiServer, `/admin/users${query ? `?${query}` : ''}`);
  },

  getUser: (apiServer: string, userId: string) => fetchApi<User>(apiServer, `/admin/users/${userId}`),

  updateUserRole: (apiServer: string, userId: string, role: 'admin' | 'user') =>
    fetchApi<void>(apiServer, `/admin/users/${userId}/role`, {
      method: 'PUT',
      body: JSON.stringify({ role }),
    }),

  updateUserStatus: (apiServer: string, userId: string, status: 'active' | 'disabled') =>
    fetchApi<void>(apiServer, `/admin/users/${userId}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    }),

  getUserPermissions: (apiServer: string, userId: string) =>
    fetchApi<UserPermissions>(apiServer, `/admin/users/${userId}/permissions`),

  setUserPermissions: (apiServer: string, userId: string, permissions: ClusterPermission[]) =>
    fetchApi<void>(apiServer, `/admin/users/${userId}/permissions`, {
      method: 'PUT',
      body: JSON.stringify({ permissions }),
    }),

  addClusterPermission: (apiServer: string, userId: string, cluster: string, namespaces: string[]) =>
    fetchApi<void>(apiServer, `/admin/users/${userId}/permissions/clusters`, {
      method: 'POST',
      body: JSON.stringify({ cluster, namespaces }),
    }),

  removeClusterPermission: (apiServer: string, userId: string, cluster: string) =>
    fetchApi<void>(apiServer, `/admin/users/${userId}/permissions/clusters/${cluster}`, {
      method: 'DELETE',
    }),

  updateClusterNamespaces: (apiServer: string, userId: string, cluster: string, namespaces: string[]) =>
    fetchApi<void>(apiServer, `/admin/users/${userId}/permissions/clusters/${cluster}/namespaces`, {
      method: 'PUT',
      body: JSON.stringify({ namespaces }),
    }),

  batchSetPermissions: (apiServer: string, userIds: string[], permissions: ClusterPermission[]) =>
    fetchApi<BatchPermissionResult>(apiServer, '/admin/permissions/batch', {
      method: 'POST',
      body: JSON.stringify({ user_ids: userIds, permissions }),
    }),
};

// Cluster APIs
export const clusterApi = {
  get: (cluster: string) => clusterFetch<Cluster>(cluster, `/clusters/${cluster}`),

  testConnection: (cluster: string) =>
    clusterFetch<ConnectionTest>(cluster, `/clusters/${cluster}/test-connection`, { method: 'POST' }),

  getNamespaces: (cluster: string) =>
    clusterFetch<Namespace[]>(cluster, `/clusters/${cluster}/namespaces`),

  getDashboard: (cluster: string, namespace?: string) => {
    const query = namespace ? `?namespace=${encodeURIComponent(namespace)}` : '';
    return clusterFetch<ClusterDashboard>(cluster, `/clusters/${cluster}/dashboard${query}`);
  },
};

// Node APIs
export const nodeApi = {
  list: (cluster: string, params?: { search?: string; page?: number; page_size?: number }) => {
    const searchParams = new URLSearchParams();
    if (params?.search) searchParams.set('search', params.search);
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.page_size) searchParams.set('page_size', params.page_size.toString());
    const query = searchParams.toString();
    return clusterFetch<PaginatedResponse<NodeListItem>>(
      cluster,
      `/clusters/${cluster}/nodes${query ? `?${query}` : ''}`
    );
  },

  get: (cluster: string, name: string) =>
    clusterFetch<NodeDetail>(cluster, `/clusters/${cluster}/nodes/${name}`),
};

// ConfigMap APIs
export const configMapApi = {
  list: (cluster: string, namespace: string) =>
    clusterFetch<K8sResource[]>(cluster, `/clusters/${cluster}/namespaces/${namespace}/configmaps`),

  get: (cluster: string, namespace: string, name: string) =>
    clusterFetch<K8sResource>(cluster, `/clusters/${cluster}/namespaces/${namespace}/configmaps/${name}`),

  create: (cluster: string, namespace: string, data: ResourceRequest) =>
    clusterFetch<K8sResource>(cluster, `/clusters/${cluster}/namespaces/${namespace}/configmaps`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (cluster: string, namespace: string, name: string, data: ResourceRequest) =>
    clusterFetch<K8sResource>(cluster, `/clusters/${cluster}/namespaces/${namespace}/configmaps/${name}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (cluster: string, namespace: string, name: string) =>
    clusterFetch<void>(cluster, `/clusters/${cluster}/namespaces/${namespace}/configmaps/${name}`, {
      method: 'DELETE',
    }),
};

// Service APIs
export const serviceApi = {
  list: (cluster: string, namespace: string) =>
    clusterFetch<K8sResource[]>(cluster, `/clusters/${cluster}/namespaces/${namespace}/services`),

  get: (cluster: string, namespace: string, name: string) =>
    clusterFetch<K8sResource>(cluster, `/clusters/${cluster}/namespaces/${namespace}/services/${name}`),

  create: (cluster: string, namespace: string, data: ResourceRequest) =>
    clusterFetch<K8sResource>(cluster, `/clusters/${cluster}/namespaces/${namespace}/services`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (cluster: string, namespace: string, name: string, data: ResourceRequest) =>
    clusterFetch<K8sResource>(cluster, `/clusters/${cluster}/namespaces/${namespace}/services/${name}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (cluster: string, namespace: string, name: string) =>
    clusterFetch<void>(cluster, `/clusters/${cluster}/namespaces/${namespace}/services/${name}`, {
      method: 'DELETE',
    }),
};

// Deployment APIs
export const deploymentApi = {
  list: (cluster: string, namespace: string) =>
    clusterFetch<Deployment[]>(cluster, `/clusters/${cluster}/namespaces/${namespace}/deployments`),

  get: (cluster: string, namespace: string, name: string) =>
    clusterFetch<Deployment>(cluster, `/clusters/${cluster}/namespaces/${namespace}/deployments/${name}`),

  create: (cluster: string, namespace: string, data: ResourceRequest) =>
    clusterFetch<K8sResource>(cluster, `/clusters/${cluster}/namespaces/${namespace}/deployments`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (cluster: string, namespace: string, name: string, data: ResourceRequest) =>
    clusterFetch<K8sResource>(cluster, `/clusters/${cluster}/namespaces/${namespace}/deployments/${name}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (cluster: string, namespace: string, name: string) =>
    clusterFetch<void>(cluster, `/clusters/${cluster}/namespaces/${namespace}/deployments/${name}`, {
      method: 'DELETE',
    }),

  getPods: (cluster: string, namespace: string, name: string) =>
    clusterFetch<Pod[]>(cluster, `/clusters/${cluster}/namespaces/${namespace}/deployments/${name}/pods`),

  restart: (cluster: string, namespace: string, name: string) =>
    clusterFetch<void>(cluster, `/clusters/${cluster}/namespaces/${namespace}/deployments/${name}/restart`, {
      method: 'POST',
    }),
};

// Pod APIs
export const podApi = {
  list: (cluster: string, namespace: string, params?: { deployment?: string; search?: string; status?: string; page?: number; page_size?: number }) => {
    const searchParams = new URLSearchParams();
    if (params?.deployment) searchParams.set('deployment', params.deployment);
    if (params?.search) searchParams.set('search', params.search);
    if (params?.status) searchParams.set('status', params.status);
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.page_size) searchParams.set('page_size', params.page_size.toString());
    const query = searchParams.toString();
    // When deployment filter is present, backend returns non-paginated array via response.Success
    // When no deployment filter, backend returns paginated response via response.SuccessWithPage
    if (params?.deployment) {
      return clusterFetch<Pod[]>(
        cluster,
        `/clusters/${cluster}/namespaces/${namespace}/pods${query ? `?${query}` : ''}`
      );
    }
    return clusterFetch<PaginatedResponse<Pod>>(
      cluster,
      `/clusters/${cluster}/namespaces/${namespace}/pods${query ? `?${query}` : ''}`
    );
  },

  get: (cluster: string, namespace: string, name: string) =>
    clusterFetch<Pod>(cluster, `/clusters/${cluster}/namespaces/${namespace}/pods/${name}`),

  delete: (cluster: string, namespace: string, name: string) =>
    clusterFetch<void>(cluster, `/clusters/${cluster}/namespaces/${namespace}/pods/${name}`, {
      method: 'DELETE',
    }),

  getLogs: (cluster: string, namespace: string, name: string, params?: {
    container?: string;
    tail_lines?: number;
    previous?: boolean;
    timestamps?: boolean;
  }) => {
    const searchParams = new URLSearchParams();
    if (params?.container) searchParams.set('container', params.container);
    if (params?.tail_lines) searchParams.set('tail_lines', params.tail_lines.toString());
    if (params?.previous) searchParams.set('previous', 'true');
    if (params?.timestamps) searchParams.set('timestamps', 'true');
    const query = searchParams.toString();
    return clusterFetch<PodLogResponse>(
      cluster,
      `/clusters/${cluster}/namespaces/${namespace}/pods/${name}/logs${query ? `?${query}` : ''}`
    );
  },

  getEvents: (cluster: string, namespace: string, name: string) =>
    clusterFetch<PodEvent[]>(cluster, `/clusters/${cluster}/namespaces/${namespace}/pods/${name}/events`),
};

// History APIs
export const historyApi = {
  list: (
    cluster: string,
    namespace: string,
    params?: { resource_type?: string; resource_name?: string; page?: number; page_size?: number }
  ) => {
    const searchParams = new URLSearchParams();
    if (params?.resource_type) searchParams.set('resource_type', params.resource_type);
    if (params?.resource_name) searchParams.set('resource_name', params.resource_name);
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.page_size) searchParams.set('page_size', params.page_size.toString());

    const query = searchParams.toString();
    return clusterFetch<PaginatedResponse<HistoryRecord>>(
      cluster,
      `/clusters/${cluster}/namespaces/${namespace}/histories${query ? `?${query}` : ''}`
    );
  },

  get: (cluster: string, namespace: string, id: number) =>
    clusterFetch<HistoryRecord>(cluster, `/clusters/${cluster}/namespaces/${namespace}/histories/${id}`),

  diff: (cluster: string, namespace: string, sourceVersion: number, targetVersion: number) =>
    clusterFetch<HistoryDiff>(
      cluster,
      `/clusters/${cluster}/namespaces/${namespace}/histories/diff?source_version=${sourceVersion}&target_version=${targetVersion}`
    ),

  rollback: (cluster: string, namespace: string, id: number, operator?: string) =>
    clusterFetch<{ success: boolean; message: string; restored_version: number; new_version: number }>(
      cluster,
      `/clusters/${cluster}/namespaces/${namespace}/histories/${id}/rollback`,
      {
        method: 'POST',
        body: JSON.stringify({ operator }),
      }
    ),

  diffWithPrevious: (cluster: string, namespace: string, id: number) =>
    clusterFetch<HistoryDiff>(
      cluster,
      `/clusters/${cluster}/namespaces/${namespace}/histories/${id}/diff-previous`
    ),
};

// HPA APIs
export const hpaApi = {
  list: (cluster: string, namespace: string) =>
    clusterFetch<HPA[]>(cluster, `/clusters/${cluster}/namespaces/${namespace}/hpas`),

  get: (cluster: string, namespace: string, name: string) =>
    clusterFetch<HPA>(cluster, `/clusters/${cluster}/namespaces/${namespace}/hpas/${name}`),

  create: (cluster: string, namespace: string, data: ResourceRequest) =>
    clusterFetch<K8sResource>(cluster, `/clusters/${cluster}/namespaces/${namespace}/hpas`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (cluster: string, namespace: string, name: string, data: ResourceRequest) =>
    clusterFetch<K8sResource>(cluster, `/clusters/${cluster}/namespaces/${namespace}/hpas/${name}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (cluster: string, namespace: string, name: string) =>
    clusterFetch<void>(cluster, `/clusters/${cluster}/namespaces/${namespace}/hpas/${name}`, {
      method: 'DELETE',
    }),
};

// Cluster Management APIs (admin only)
export const clusterManageApi = {
  list: () => fetchApi<ClusterDetail[]>('', '/admin/clusters'),

  get: (name: string) => fetchApi<ClusterDetail>('', `/admin/clusters/${name}`),

  add: (req: AddClusterRequest) =>
    fetchApi<ClusterDetail>('', '/admin/clusters', {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  update: (name: string, req: UpdateClusterRequest) =>
    fetchApi<ClusterDetail>('', `/admin/clusters/${name}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    }),

  delete: (name: string) =>
    fetchApi<void>('', `/admin/clusters/${name}`, {
      method: 'DELETE',
    }),

  testConnection: (req: TestNewConnectionRequest) =>
    fetchApi<ConnectionTest>('', '/admin/clusters/test-connection', {
      method: 'POST',
      body: JSON.stringify(req),
    }),
};
