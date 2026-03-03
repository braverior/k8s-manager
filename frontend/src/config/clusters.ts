import type { Cluster } from '@/types';

// Cluster config is injected at runtime via window.__K8S_CLUSTERS__
// In Docker/nginx, this is set by the entrypoint script from K8S_MANAGER_CLUSTERS env var
// In development, it falls back to VITE_CLUSTERS env var
declare global {
  interface Window {
    __K8S_CLUSTERS__?: Cluster[];
  }
}

function loadClusters(): Cluster[] {
  // Priority 1: Runtime injection (Docker/nginx)
  if (window.__K8S_CLUSTERS__ && Array.isArray(window.__K8S_CLUSTERS__)) {
    return window.__K8S_CLUSTERS__;
  }

  // Priority 2: Vite env var (development)
  const envClusters = import.meta.env.VITE_CLUSTERS;
  if (envClusters) {
    try {
      const parsed = JSON.parse(envClusters);
      if (Array.isArray(parsed)) {
        return parsed;
      }
    } catch (e) {
      console.error('Failed to parse VITE_CLUSTERS:', e);
    }
  }

  // Fallback: empty
  console.warn('No cluster configuration found. Set K8S_MANAGER_CLUSTERS env var or VITE_CLUSTERS.');
  return [];
}

let _clusters: Cluster[] | null = null;

export function getClusters(): Cluster[] {
  if (_clusters === null) {
    _clusters = loadClusters();
  }
  return _clusters;
}

export function getClusterApiServer(clusterName: string): string {
  const cluster = getClusters().find((c) => c.name === clusterName);
  if (!cluster) {
    throw new Error(`Cluster "${clusterName}" not found in config`);
  }
  return cluster.api_server;
}
