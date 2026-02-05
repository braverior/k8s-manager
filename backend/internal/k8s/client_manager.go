package k8s

import (
	"context"
	"fmt"
	"os"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	"k8s_api_server/internal/config"
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

// LoadFromConfig 从配置文件加载所有集群
func (m *ClientManager) LoadFromConfig(clusterConfigs []config.ClusterConfig) error {
	for _, cc := range clusterConfigs {
		if err := m.addCluster(cc); err != nil {
			logger.Error(fmt.Sprintf("Failed to load cluster %s: %v", cc.Name, err))
			continue
		}
		logger.Info(fmt.Sprintf("Loaded cluster: %s", cc.Name))
	}
	return nil
}

func (m *ClientManager) addCluster(cc config.ClusterConfig) error {
	var cfg *rest.Config
	var err error

	// 根据类型选择配置方式
	clusterType := cc.Type
	if clusterType == "" {
		clusterType = "in-cluster" // 默认使用 in-cluster 模式
	}

	switch clusterType {
	case "in-cluster":
		// 使用 ServiceAccount 进行认证
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return fmt.Errorf("failed to create in-cluster config: %w", err)
		}
	case "kubeconfig":
		// 从 kubeconfig 文件读取配置
		if cc.KubeconfigPath == "" {
			return fmt.Errorf("kubeconfig_path is required for type=kubeconfig")
		}
		kubeconfigData, err := os.ReadFile(cc.KubeconfigPath)
		if err != nil {
			return fmt.Errorf("failed to read kubeconfig file: %w", err)
		}
		cfg, err = clientcmd.RESTConfigFromKubeConfig(kubeconfigData)
		if err != nil {
			return fmt.Errorf("failed to parse kubeconfig: %w", err)
		}
	default:
		return fmt.Errorf("unsupported cluster type: %s (use 'in-cluster' or 'kubeconfig')", clusterType)
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// 创建 metrics client（可选，失败不影响主功能）
	metricsClient, err := metricsv.NewForConfig(cfg)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to create metrics client for cluster %s: %v", cc.Name, err))
		metricsClient = nil
	}

	m.mu.Lock()
	m.clients[cc.Name] = client
	m.metricsClients[cc.Name] = metricsClient
	m.configs[cc.Name] = cfg
	m.clusters[cc.Name] = &ClusterInfo{
		Name:        cc.Name,
		Description: cc.Description,
		APIServer:   cfg.Host,
		Status:      "connected",
	}
	m.mu.Unlock()

	return nil
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
