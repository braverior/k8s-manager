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
  BatchPermissionResult
} from '@/types';

const API_BASE = '/api/v1';

// Get token from localStorage
function getAuthToken(): string | null {
  return localStorage.getItem('token');
}

async function fetchApi<T>(url: string, options?: RequestInit): Promise<T> {
  const token = getAuthToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}${url}`, {
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
    throw new Error(error.message || `HTTP ${response.status}`);
  }

  const data: ApiResponse<T> = await response.json();

  if (data.code !== 0) {
    throw new Error(data.message || 'API Error');
  }

  return data.data;
}

// Auth APIs
export const authApi = {
  getFeishuConfig: () => fetchApi<FeishuConfig>('/auth/feishu/config'),

  feishuLogin: (code: string, state?: string) =>
    fetchApi<LoginResponse>('/auth/feishu/login', {
      method: 'POST',
      body: JSON.stringify({ code, state }),
    }),

  getMe: () => fetchApi<User>('/auth/me'),

  logout: () =>
    fetchApi<void>('/auth/logout', {
      method: 'POST',
    }),
};

// Admin APIs
export const adminApi = {
  // User management
  listUsers: (params?: {
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
    return fetchApi<PaginatedResponse<User>>(`/admin/users${query ? `?${query}` : ''}`);
  },

  getUser: (userId: string) => fetchApi<User>(`/admin/users/${userId}`),

  updateUserRole: (userId: string, role: 'admin' | 'user') =>
    fetchApi<void>(`/admin/users/${userId}/role`, {
      method: 'PUT',
      body: JSON.stringify({ role }),
    }),

  updateUserStatus: (userId: string, status: 'active' | 'disabled') =>
    fetchApi<void>(`/admin/users/${userId}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    }),

  // Permission management
  getUserPermissions: (userId: string) =>
    fetchApi<UserPermissions>(`/admin/users/${userId}/permissions`),

  setUserPermissions: (userId: string, permissions: ClusterPermission[]) =>
    fetchApi<void>(`/admin/users/${userId}/permissions`, {
      method: 'PUT',
      body: JSON.stringify({ permissions }),
    }),

  addClusterPermission: (userId: string, cluster: string, namespaces: string[]) =>
    fetchApi<void>(`/admin/users/${userId}/permissions/clusters`, {
      method: 'POST',
      body: JSON.stringify({ cluster, namespaces }),
    }),

  removeClusterPermission: (userId: string, cluster: string) =>
    fetchApi<void>(`/admin/users/${userId}/permissions/clusters/${cluster}`, {
      method: 'DELETE',
    }),

  updateClusterNamespaces: (userId: string, cluster: string, namespaces: string[]) =>
    fetchApi<void>(`/admin/users/${userId}/permissions/clusters/${cluster}/namespaces`, {
      method: 'PUT',
      body: JSON.stringify({ namespaces }),
    }),

  batchSetPermissions: (userIds: string[], permissions: ClusterPermission[]) =>
    fetchApi<BatchPermissionResult>('/admin/permissions/batch', {
      method: 'POST',
      body: JSON.stringify({ user_ids: userIds, permissions }),
    }),
};

// Cluster APIs
export const clusterApi = {
  list: () => fetchApi<Cluster[]>('/clusters'),

  get: (cluster: string) => fetchApi<Cluster>(`/clusters/${cluster}`),

  testConnection: (cluster: string) =>
    fetchApi<ConnectionTest>(`/clusters/${cluster}/test-connection`, { method: 'POST' }),

  getNamespaces: (cluster: string) =>
    fetchApi<Namespace[]>(`/clusters/${cluster}/namespaces`),

  getDashboard: (cluster: string) =>
    fetchApi<ClusterDashboard>(`/clusters/${cluster}/dashboard`),
};

// Node APIs
export const nodeApi = {
  list: (cluster: string) =>
    fetchApi<NodeListItem[]>(`/clusters/${cluster}/nodes`),

  get: (cluster: string, name: string) =>
    fetchApi<NodeDetail>(`/clusters/${cluster}/nodes/${name}`),
};

// ConfigMap APIs
export const configMapApi = {
  list: (cluster: string, namespace: string) =>
    fetchApi<K8sResource[]>(`/clusters/${cluster}/namespaces/${namespace}/configmaps`),

  get: (cluster: string, namespace: string, name: string) =>
    fetchApi<K8sResource>(`/clusters/${cluster}/namespaces/${namespace}/configmaps/${name}`),

  create: (cluster: string, namespace: string, data: ResourceRequest) =>
    fetchApi<K8sResource>(`/clusters/${cluster}/namespaces/${namespace}/configmaps`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (cluster: string, namespace: string, name: string, data: ResourceRequest) =>
    fetchApi<K8sResource>(`/clusters/${cluster}/namespaces/${namespace}/configmaps/${name}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (cluster: string, namespace: string, name: string) =>
    fetchApi<void>(`/clusters/${cluster}/namespaces/${namespace}/configmaps/${name}`, {
      method: 'DELETE',
    }),
};

// Deployment APIs
export const deploymentApi = {
  list: (cluster: string, namespace: string) =>
    fetchApi<Deployment[]>(`/clusters/${cluster}/namespaces/${namespace}/deployments`),

  get: (cluster: string, namespace: string, name: string) =>
    fetchApi<Deployment>(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}`),

  create: (cluster: string, namespace: string, data: ResourceRequest) =>
    fetchApi<K8sResource>(`/clusters/${cluster}/namespaces/${namespace}/deployments`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (cluster: string, namespace: string, name: string, data: ResourceRequest) =>
    fetchApi<K8sResource>(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (cluster: string, namespace: string, name: string) =>
    fetchApi<void>(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}`, {
      method: 'DELETE',
    }),

  getPods: (cluster: string, namespace: string, name: string) =>
    fetchApi<Pod[]>(`/clusters/${cluster}/namespaces/${namespace}/deployments/${name}/pods`),
};

// Pod APIs
export const podApi = {
  list: (cluster: string, namespace: string, params?: { deployment?: string }) => {
    const searchParams = new URLSearchParams();
    if (params?.deployment) searchParams.set('deployment', params.deployment);
    const query = searchParams.toString();
    return fetchApi<Pod[]>(
      `/clusters/${cluster}/namespaces/${namespace}/pods${query ? `?${query}` : ''}`
    );
  },

  get: (cluster: string, namespace: string, name: string) =>
    fetchApi<Pod>(`/clusters/${cluster}/namespaces/${namespace}/pods/${name}`),

  delete: (cluster: string, namespace: string, name: string) =>
    fetchApi<void>(`/clusters/${cluster}/namespaces/${namespace}/pods/${name}`, {
      method: 'DELETE',
    }),
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
    return fetchApi<PaginatedResponse<HistoryRecord>>(
      `/clusters/${cluster}/namespaces/${namespace}/histories${query ? `?${query}` : ''}`
    );
  },

  get: (cluster: string, namespace: string, id: number) =>
    fetchApi<HistoryRecord>(`/clusters/${cluster}/namespaces/${namespace}/histories/${id}`),

  diff: (cluster: string, namespace: string, sourceVersion: number, targetVersion: number) =>
    fetchApi<HistoryDiff>(
      `/clusters/${cluster}/namespaces/${namespace}/histories/diff?source_version=${sourceVersion}&target_version=${targetVersion}`
    ),

  rollback: (cluster: string, namespace: string, id: number, operator?: string) =>
    fetchApi<{ success: boolean; message: string; restored_version: number; new_version: number }>(
      `/clusters/${cluster}/namespaces/${namespace}/histories/${id}/rollback`,
      {
        method: 'POST',
        body: JSON.stringify({ operator }),
      }
    ),
};

// HPA APIs
export const hpaApi = {
  list: (cluster: string, namespace: string) =>
    fetchApi<HPA[]>(`/clusters/${cluster}/namespaces/${namespace}/hpas`),

  get: (cluster: string, namespace: string, name: string) =>
    fetchApi<HPA>(`/clusters/${cluster}/namespaces/${namespace}/hpas/${name}`),

  create: (cluster: string, namespace: string, data: ResourceRequest) =>
    fetchApi<K8sResource>(`/clusters/${cluster}/namespaces/${namespace}/hpas`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  update: (cluster: string, namespace: string, name: string, data: ResourceRequest) =>
    fetchApi<K8sResource>(`/clusters/${cluster}/namespaces/${namespace}/hpas/${name}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  delete: (cluster: string, namespace: string, name: string) =>
    fetchApi<void>(`/clusters/${cluster}/namespaces/${namespace}/hpas/${name}`, {
      method: 'DELETE',
    }),
};
