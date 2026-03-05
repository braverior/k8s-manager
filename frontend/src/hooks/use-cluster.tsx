import { useState, useEffect, createContext, useContext, useCallback, type ReactNode } from 'react';
import { clusterApi, fetchApi } from '@/api';
import { useAuth } from '@/hooks/use-auth';
import type { Cluster, Namespace } from '@/types';

interface ClusterContextType {
  clusters: Cluster[];
  namespaces: Namespace[];
  selectedCluster: string;
  selectedNamespace: string;
  setSelectedCluster: (cluster: string) => void;
  setSelectedNamespace: (namespace: string) => void;
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
  refreshClusters: () => Promise<void>;
}

const ClusterContext = createContext<ClusterContextType | null>(null);

export function ClusterProvider({ children }: { children: ReactNode }) {
  const { user, hasClusterPermission, hasNamespacePermission } = useAuth();
  const [allClusters, setAllClusters] = useState<Cluster[]>([]);
  const [allNamespaces, setAllNamespaces] = useState<Namespace[]>([]);
  const [selectedCluster, setSelectedCluster] = useState<string>('');
  const [selectedNamespace, setSelectedNamespace] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [error] = useState<string | null>(null);

  // Fetch clusters from API
  const fetchClusters = useCallback(async () => {
    try {
      setLoading(true);
      const data = await fetchApi<Cluster[]>('', '/clusters');
      setAllClusters(data || []);
    } catch (err) {
      console.error('Failed to fetch clusters:', err);
      setAllClusters([]);
    } finally {
      setLoading(false);
    }
  }, []);

  // Load clusters on mount
  useEffect(() => {
    fetchClusters();
  }, [fetchClusters]);

  // Filter clusters based on permissions
  const clusters = allClusters.filter((c) => hasClusterPermission(c.name));

  // Filter namespaces based on permissions
  const namespaces = allNamespaces.filter((ns) =>
    hasNamespacePermission(selectedCluster, ns.name)
  );

  const fetchNamespaces = useCallback(async (cluster: string) => {
    if (!cluster) return;
    try {
      const data = await clusterApi.getNamespaces(cluster);
      setAllNamespaces(data || []);
    } catch (err) {
      console.error('Failed to fetch namespaces:', err);
      setAllNamespaces([]);
    }
  }, []);

  // Auto-select first available cluster when clusters change or user permissions change
  useEffect(() => {
    if (clusters.length > 0) {
      if (!selectedCluster || !clusters.some((c) => c.name === selectedCluster)) {
        setSelectedCluster(clusters[0].name);
      }
    } else {
      setSelectedCluster('');
    }
  }, [clusters, selectedCluster, user]);

  // Fetch namespaces when cluster changes
  useEffect(() => {
    if (selectedCluster) {
      fetchNamespaces(selectedCluster);
    } else {
      setAllNamespaces([]);
    }
  }, [selectedCluster, fetchNamespaces]);

  // Auto-select first available namespace when namespaces change
  useEffect(() => {
    if (namespaces.length > 0) {
      if (!selectedNamespace || !namespaces.some((ns) => ns.name === selectedNamespace)) {
        const defaultNs = namespaces.find((ns) => ns.name === 'default');
        setSelectedNamespace(defaultNs ? 'default' : namespaces[0].name);
      }
    } else {
      setSelectedNamespace('');
    }
  }, [namespaces, selectedNamespace]);

  const refresh = useCallback(async () => {
    if (selectedCluster) {
      await fetchNamespaces(selectedCluster);
    }
  }, [fetchNamespaces, selectedCluster]);

  const refreshClusters = useCallback(async () => {
    await fetchClusters();
  }, [fetchClusters]);

  return (
    <ClusterContext.Provider
      value={{
        clusters,
        namespaces,
        selectedCluster,
        selectedNamespace,
        setSelectedCluster,
        setSelectedNamespace,
        loading,
        error,
        refresh,
        refreshClusters,
      }}
    >
      {children}
    </ClusterContext.Provider>
  );
}

export function useCluster() {
  const context = useContext(ClusterContext);
  if (!context) {
    throw new Error('useCluster must be used within a ClusterProvider');
  }
  return context;
}
