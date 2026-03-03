package service

import (
	"context"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"k8s_api_server/internal/k8s"
	k8soperator "k8s_api_server/internal/k8s/operator"
	"k8s_api_server/internal/model/dto"
	apperrors "k8s_api_server/internal/pkg/errors"
)

type DashboardService struct {
	clientManager *k8s.ClientManager
}

func NewDashboardService(clientManager *k8s.ClientManager) *DashboardService {
	return &DashboardService{
		clientManager: clientManager,
	}
}

// GetOverview 获取集群概览信息
func (s *DashboardService) GetOverview(ctx context.Context, clusterName string) (*dto.DashboardResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	clusterInfo, err := s.clientManager.GetClusterInfo(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 404, 404, "集群不存在")
	}

	resp := &dto.DashboardResponse{
		ClusterName: clusterName,
		Status:      clusterInfo.Status,
	}

	// 获取 Kubernetes 版本
	if version, err := client.Discovery().ServerVersion(); err == nil {
		resp.Version = version.GitVersion
	}

	// 获取节点列表
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err == nil {
		resp.Resources.Nodes = len(nodes.Items)
	}

	// 获取命名空间数量
	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err == nil {
		resp.Resources.Namespaces = len(namespaces.Items)
	}

	// 获取所有命名空间的 Pod
	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err == nil {
		resp.Resources.Pods = s.countPodsByPhase(pods.Items)
	}

	// 获取所有命名空间的 Deployment
	deployments, err := client.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err == nil {
		resp.Resources.Deployments = len(deployments.Items)
	}

	// 获取所有命名空间的 Service
	services, err := client.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err == nil {
		resp.Resources.Services = len(services.Items)
	}

	// 获取所有命名空间的 ConfigMap
	configMaps, err := client.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{})
	if err == nil {
		resp.Resources.ConfigMaps = len(configMaps.Items)
	}

	// 获取所有命名空间的 HPA（使用版本自适应的 operator）
	hpaOp := k8soperator.NewHPAOperator(client)
	hpas, err := hpaOp.List(ctx, "")
	if err == nil {
		resp.Resources.HPAs = len(hpas)
		// 添加 HPA 概要信息
		resp.HPASummaries = s.buildHPASummaries(hpas)
	}

	// 获取集群资源容量和使用情况
	if nodes != nil {
		resp.Capacity, resp.Usage = s.calculateClusterResources(ctx, clusterName, nodes.Items)
	}

	return resp, nil
}

// countPodsByPhase 按状态统计 Pod 数量
func (s *DashboardService) countPodsByPhase(pods []corev1.Pod) dto.PodStats {
	stats := dto.PodStats{
		Total: len(pods),
	}

	for _, pod := range pods {
		switch pod.Status.Phase {
		case corev1.PodRunning:
			stats.Running++
		case corev1.PodPending:
			stats.Pending++
		case corev1.PodSucceeded:
			stats.Succeeded++
		case corev1.PodFailed:
			stats.Failed++
		}
	}

	return stats
}

// calculateClusterResources 计算集群资源容量和使用情况
func (s *DashboardService) calculateClusterResources(ctx context.Context, clusterName string, nodes []corev1.Node) (*dto.ClusterCapacity, *dto.ClusterUsage) {
	// 计算总容量
	totalCPU := resource.NewQuantity(0, resource.DecimalSI)
	totalMemory := resource.NewQuantity(0, resource.BinarySI)
	totalPods := int64(0)

	for _, node := range nodes {
		totalCPU.Add(*node.Status.Allocatable.Cpu())
		totalMemory.Add(*node.Status.Allocatable.Memory())
		totalPods += node.Status.Allocatable.Pods().Value()
	}

	capacity := &dto.ClusterCapacity{
		CPU:    formatCPUValue(totalCPU),
		Memory: formatMemory(totalMemory),
		Pods:   totalPods,
	}

	// 尝试获取使用量（需要 metrics-server）
	metricsClient, err := s.clientManager.GetMetricsClient(clusterName)
	if err != nil || metricsClient == nil {
		return capacity, nil
	}

	// 获取节点指标
	nodeMetrics, err := metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return capacity, nil
	}

	// 计算已使用量
	usedCPU := resource.NewQuantity(0, resource.DecimalSI)
	usedMemory := resource.NewQuantity(0, resource.BinarySI)

	for _, metric := range nodeMetrics.Items {
		usedCPU.Add(*metric.Usage.Cpu())
		usedMemory.Add(*metric.Usage.Memory())
	}

	// 计算百分比
	cpuPercent := float64(0)
	if totalCPU.MilliValue() > 0 {
		cpuPercent = float64(usedCPU.MilliValue()) / float64(totalCPU.MilliValue()) * 100
	}

	memoryPercent := float64(0)
	if totalMemory.Value() > 0 {
		memoryPercent = float64(usedMemory.Value()) / float64(totalMemory.Value()) * 100
	}

	usage := &dto.ClusterUsage{
		CPU:              formatCPUValue(usedCPU),
		CPUPercentage:    round(cpuPercent, 2),
		Memory:           formatMemory(usedMemory),
		MemoryPercentage: round(memoryPercent, 2),
	}

	return capacity, usage
}

// formatMemory 格式化内存为 Mi 单位
func formatMemory(q *resource.Quantity) string {
	bytes := q.Value()
	if bytes <= 0 {
		return "0"
	}
	mi := float64(bytes) / (1024 * 1024)
	if mi >= 1024 {
		gi := mi / 1024
		return fmt.Sprintf("%.1fGi", gi)
	}
	return fmt.Sprintf("%.0fMi", mi)
}

// formatCPUValue 格式化 CPU 为毫核(m)格式
func formatCPUValue(q *resource.Quantity) string {
	if q == nil {
		return "0"
	}
	milliValue := q.MilliValue()
	if milliValue >= 1000 {
		cores := float64(milliValue) / 1000
		if cores == float64(int64(cores)) {
			return fmt.Sprintf("%d", int64(cores))
		}
		return fmt.Sprintf("%.2f", cores)
	}
	return fmt.Sprintf("%dm", milliValue)
}

// round 四舍五入到指定小数位
func round(val float64, precision int) float64 {
	p := float64(1)
	for i := 0; i < precision; i++ {
		p *= 10
	}
	return float64(int(val*p+0.5)) / p
}

// buildHPASummaries 构建 HPA 概要列表
func (s *DashboardService) buildHPASummaries(hpas []autoscalingv2.HorizontalPodAutoscaler) []dto.HPASummary {
	summaries := make([]dto.HPASummary, 0, len(hpas))

	for _, hpa := range hpas {
		summary := dto.HPASummary{
			Name:            hpa.Name,
			Namespace:       hpa.Namespace,
			TargetKind:      hpa.Spec.ScaleTargetRef.Kind,
			TargetName:      hpa.Spec.ScaleTargetRef.Name,
			MaxReplicas:     hpa.Spec.MaxReplicas,
			CurrentReplicas: hpa.Status.CurrentReplicas,
			DesiredReplicas: hpa.Status.DesiredReplicas,
		}

		// MinReplicas 默认为 1
		if hpa.Spec.MinReplicas != nil {
			summary.MinReplicas = *hpa.Spec.MinReplicas
		} else {
			summary.MinReplicas = 1
		}

		// 解析指标配置
		for _, metric := range hpa.Spec.Metrics {
			if metric.Type == autoscalingv2.ResourceMetricSourceType && metric.Resource != nil {
				if metric.Resource.Name == corev1.ResourceCPU {
					if metric.Resource.Target.Type == autoscalingv2.UtilizationMetricType && metric.Resource.Target.AverageUtilization != nil {
						summary.CPUTargetUtilization = metric.Resource.Target.AverageUtilization
					}
				} else if metric.Resource.Name == corev1.ResourceMemory {
					if metric.Resource.Target.Type == autoscalingv2.UtilizationMetricType && metric.Resource.Target.AverageUtilization != nil {
						summary.MemoryTargetUtilization = metric.Resource.Target.AverageUtilization
					}
				}
			}
		}

		// 解析当前指标值
		for _, metric := range hpa.Status.CurrentMetrics {
			if metric.Type == autoscalingv2.ResourceMetricSourceType && metric.Resource != nil {
				if metric.Resource.Name == corev1.ResourceCPU {
					if metric.Resource.Current.AverageUtilization != nil {
						summary.CPUCurrentUtilization = metric.Resource.Current.AverageUtilization
					}
				} else if metric.Resource.Name == corev1.ResourceMemory {
					if metric.Resource.Current.AverageUtilization != nil {
						summary.MemoryCurrentUtilization = metric.Resource.Current.AverageUtilization
					}
				}
			}
		}

		summaries = append(summaries, summary)
	}

	return summaries
}
