package k8s

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	"k8s_api_server/internal/pkg/logger"
)

type ClusterInfo struct {
	Name        string
	Description string
	APIServer   string
	Status      string
}

type ClientManager struct {
	clients        map[string]*kubernetes.Clientset
	metricsClients map[string]metricsv.Interface
	configs        map[string]*rest.Config
	clusters       map[string]*ClusterInfo
	mu             sync.RWMutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:        make(map[string]*kubernetes.Clientset),
		metricsClients: make(map[string]metricsv.Interface),
		configs:        make(map[string]*rest.Config),
		clusters:       make(map[string]*ClusterInfo),
	}
}

func (m *ClientManager) GetClient(clusterName string) (*kubernetes.Clientset, error) {
	m.mu.RLock()
	client, ok := m.clients[clusterName]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("cluster %s not found", clusterName)
	}

	return client, nil
}

func (m *ClientManager) GetConfig(clusterName string) (*rest.Config, error) {
	m.mu.RLock()
	cfg, ok := m.configs[clusterName]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("config for cluster %s not found", clusterName)
	}

	return cfg, nil
}

func (m *ClientManager) GetMetricsClient(clusterName string) (metricsv.Interface, error) {
	m.mu.RLock()
	client, ok := m.metricsClients[clusterName]
	m.mu.RUnlock()

	if !ok || client == nil {
		return nil, fmt.Errorf("metrics client for cluster %s not available", clusterName)
	}

	return client, nil
}

func (m *ClientManager) ListClusters() []ClusterInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []ClusterInfo
	for _, c := range m.clusters {
		list = append(list, *c)
	}
	return list
}

func (m *ClientManager) GetClusterInfo(clusterName string) (*ClusterInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, ok := m.clusters[clusterName]
	if !ok {
		return nil, fmt.Errorf("cluster %s not found", clusterName)
	}
	return info, nil
}

func (m *ClientManager) TestConnection(ctx context.Context, clusterName string) (string, error) {
	client, err := m.GetClient(clusterName)
	if err != nil {
		return "", err
	}

	version, err := client.Discovery().ServerVersion()
	if err != nil {
		m.mu.Lock()
		if info, ok := m.clusters[clusterName]; ok {
			info.Status = "disconnected"
		}
		m.mu.Unlock()
		return "", fmt.Errorf("failed to connect to cluster: %w", err)
	}

	m.mu.Lock()
	if info, ok := m.clusters[clusterName]; ok {
		info.Status = "connected"
	}
	m.mu.Unlock()

	return version.GitVersion, nil
}

// AddClusterFromKubeconfig 从 kubeconfig 字节创建客户端并加载到管理器
func (m *ClientManager) AddClusterFromKubeconfig(name, description string, kubeconfigData []byte) error {
	cfg, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigData)
	if err != nil {
		return fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	metricsClient, err := metricsv.NewForConfig(cfg)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to create metrics client for cluster %s: %v", name, err))
		metricsClient = nil
	}

	m.mu.Lock()
	m.clients[name] = client
	m.metricsClients[name] = metricsClient
	m.configs[name] = cfg
	m.clusters[name] = &ClusterInfo{
		Name:        name,
		Description: description,
		APIServer:   cfg.Host,
		Status:      "connected",
	}
	m.mu.Unlock()

	return nil
}

// RemoveCluster 从管理器中移除集群
func (m *ClientManager) RemoveCluster(name string) {
	m.mu.Lock()
	delete(m.clients, name)
	delete(m.metricsClients, name)
	delete(m.configs, name)
	delete(m.clusters, name)
	m.mu.Unlock()
}

// HasCluster 检查集群是否已加载
func (m *ClientManager) HasCluster(name string) bool {
	m.mu.RLock()
	_, ok := m.clients[name]
	m.mu.RUnlock()
	return ok
}
