package service

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"golang.org/x/sync/errgroup"

	"k8s_api_server/internal/k8s"
	"k8s_api_server/internal/model/dto"
	apperrors "k8s_api_server/internal/pkg/errors"
)

type NodeService struct {
	clientManager *k8s.ClientManager
}

func NewNodeService(clientManager *k8s.ClientManager) *NodeService {
	return &NodeService{
		clientManager: clientManager,
	}
}

// List 获取节点列表（支持分页和搜索）
func (s *NodeService) List(ctx context.Context, clusterName string, query *dto.ResourceQuery) ([]dto.NodeListItem, int64, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	var nodes *corev1.NodeList
	var pods *corev1.PodList
	var nodeMetricsList *metricsv1beta1.NodeMetricsList

	g, gCtx := errgroup.WithContext(ctx)

	// 获取节点列表（fatal）
	g.Go(func() error {
		var err error
		nodes, err = client.CoreV1().Nodes().List(gCtx, cachedListOptions())
		if err != nil {
			return apperrors.Wrap(err, 500, 500, "获取节点列表失败")
		}
		return nil
	})

	// 获取所有 Pod 统计每个节点的 Pod 数量（non-fatal）
	g.Go(func() error {
		podList, err := client.CoreV1().Pods("").List(gCtx, cachedListOptions())
		if err == nil {
			pods = podList
		}
		return nil
	})

	// 获取节点指标（non-fatal）
	g.Go(func() error {
		metricsClient, _ := s.clientManager.GetMetricsClient(clusterName)
		if metricsClient != nil {
			if ml, err := metricsClient.MetricsV1beta1().NodeMetricses().List(gCtx, cachedListOptions()); err == nil {
				nodeMetricsList = ml
			}
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, 0, err
	}

	// 构建 Pod 计数映射
	podCountByNode := make(map[string]int)
	if pods != nil {
		for _, pod := range pods.Items {
			if pod.Spec.NodeName != "" {
				podCountByNode[pod.Spec.NodeName]++
			}
		}
	}

	// 构建指标映射
	metricsMap := make(map[string]*nodeMetrics)
	if nodeMetricsList != nil {
		for _, m := range nodeMetricsList.Items {
			metricsMap[m.Name] = &nodeMetrics{
				CPU:    m.Usage.Cpu(),
				Memory: m.Usage.Memory(),
			}
		}
	}

	result := make([]dto.NodeListItem, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		item := dto.NodeListItem{
			Name:           node.Name,
			Status:         s.getNodeStatus(&node),
			Roles:          s.getNodeRoles(&node),
			KubeletVersion: node.Status.NodeInfo.KubeletVersion,
			CPUCapacity:    formatCPU(node.Status.Allocatable.Cpu()),
			MemoryCapacity: formatMemoryQuantity(node.Status.Allocatable.Memory()),
			PodCount:       podCountByNode[node.Name],
			CreatedAt:      node.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		}

		// 获取内部 IP
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				item.InternalIP = addr.Address
				break
			}
		}

		// 添加使用量信息
		if metrics, ok := metricsMap[node.Name]; ok {
			item.CPUUsage = formatCPU(metrics.CPU)
			item.MemoryUsage = formatMemoryQuantity(metrics.Memory)

			allocatableCPU := node.Status.Allocatable.Cpu().MilliValue()
			allocatableMemory := node.Status.Allocatable.Memory().Value()

			if allocatableCPU > 0 {
				item.CPUPercentage = roundFloat(float64(metrics.CPU.MilliValue())/float64(allocatableCPU)*100, 2)
			}
			if allocatableMemory > 0 {
				item.MemoryPercentage = roundFloat(float64(metrics.Memory.Value())/float64(allocatableMemory)*100, 2)
			}
		}

		result = append(result, item)
	}

	// 搜索过滤
	if query != nil && query.Search != "" {
		search := strings.ToLower(query.Search)
		filtered := make([]dto.NodeListItem, 0)
		for _, item := range result {
			if strings.Contains(strings.ToLower(item.Name), search) {
				filtered = append(filtered, item)
			}
		}
		result = filtered
	}

	total := int64(len(result))

	// 分页
	if query != nil && query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		if offset >= len(result) {
			return []dto.NodeListItem{}, total, nil
		}
		end := offset + query.PageSize
		if end > len(result) {
			end = len(result)
		}
		result = result[offset:end]
	}

	return result, total, nil
}

// Get 获取单个节点详情
func (s *NodeService) Get(ctx context.Context, clusterName, nodeName string) (*dto.NodeResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, apperrors.Wrap(err, 404, 404, "节点不存在")
	}

	// 统计该节点上的 Pod 数量
	pods, _ := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	podCount := 0
	if pods != nil {
		podCount = len(pods.Items)
	}

	resp := &dto.NodeResponse{
		Name:             node.Name,
		Status:           s.getNodeStatus(node),
		Roles:            s.getNodeRoles(node),
		KubeletVersion:   node.Status.NodeInfo.KubeletVersion,
		KernelVersion:    node.Status.NodeInfo.KernelVersion,
		OSImage:          node.Status.NodeInfo.OSImage,
		OS:               node.Status.NodeInfo.OperatingSystem,
		Architecture:     node.Status.NodeInfo.Architecture,
		ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
		CreatedAt:        node.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
		PodCount:         podCount,
		Labels:           node.Labels,
	}

	// 获取 IP 地址
	for _, addr := range node.Status.Addresses {
		switch addr.Type {
		case corev1.NodeInternalIP:
			resp.InternalIP = addr.Address
		case corev1.NodeExternalIP:
			resp.ExternalIP = addr.Address
		}
	}

	// 资源容量
	resp.Capacity = dto.NodeResourceInfo{
		CPU:              formatCPU(node.Status.Capacity.Cpu()),
		Memory:           formatMemoryQuantity(node.Status.Capacity.Memory()),
		Pods:             node.Status.Capacity.Pods().String(),
		EphemeralStorage: formatMemoryQuantity(node.Status.Capacity.StorageEphemeral()),
	}

	// 可分配资源
	resp.Allocatable = dto.NodeResourceInfo{
		CPU:              formatCPU(node.Status.Allocatable.Cpu()),
		Memory:           formatMemoryQuantity(node.Status.Allocatable.Memory()),
		Pods:             node.Status.Allocatable.Pods().String(),
		EphemeralStorage: formatMemoryQuantity(node.Status.Allocatable.StorageEphemeral()),
	}

	// 节点条件
	resp.Conditions = make([]dto.NodeCondition, 0, len(node.Status.Conditions))
	for _, cond := range node.Status.Conditions {
		resp.Conditions = append(resp.Conditions, dto.NodeCondition{
			Type:              string(cond.Type),
			Status:           string(cond.Status),
			Reason:           cond.Reason,
			Message:          cond.Message,
			LastHeartbeatTime: cond.LastHeartbeatTime.Format("2006-01-02T15:04:05Z"),
		})
	}

	// 节点污点
	if len(node.Spec.Taints) > 0 {
		resp.Taints = make([]dto.NodeTaint, 0, len(node.Spec.Taints))
		for _, taint := range node.Spec.Taints {
			resp.Taints = append(resp.Taints, dto.NodeTaint{
				Key:    taint.Key,
				Value:  taint.Value,
				Effect: string(taint.Effect),
			})
		}
	}

	// 获取资源使用情况
	metricsClient, _ := s.clientManager.GetMetricsClient(clusterName)
	if metricsClient != nil {
		if nodeMetrics, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, nodeName, metav1.GetOptions{}); err == nil {
			allocatableCPU := node.Status.Allocatable.Cpu().MilliValue()
			allocatableMemory := node.Status.Allocatable.Memory().Value()

			cpuPercent := float64(0)
			memoryPercent := float64(0)

			if allocatableCPU > 0 {
				cpuPercent = float64(nodeMetrics.Usage.Cpu().MilliValue()) / float64(allocatableCPU) * 100
			}
			if allocatableMemory > 0 {
				memoryPercent = float64(nodeMetrics.Usage.Memory().Value()) / float64(allocatableMemory) * 100
			}

			resp.Usage = &dto.NodeUsageInfo{
				CPU:              formatCPU(nodeMetrics.Usage.Cpu()),
				CPUPercentage:    roundFloat(cpuPercent, 2),
				Memory:           formatMemoryQuantity(nodeMetrics.Usage.Memory()),
				MemoryPercentage: roundFloat(memoryPercent, 2),
			}
		}
	}

	return resp, nil
}

type nodeMetrics struct {
	CPU    *resource.Quantity
	Memory *resource.Quantity
}

// getNodeStatus 获取节点状态
func (s *NodeService) getNodeStatus(node *corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}

// getNodeRoles 获取节点角色
func (s *NodeService) getNodeRoles(node *corev1.Node) []string {
	var roles []string
	for label := range node.Labels {
		if strings.HasPrefix(label, "node-role.kubernetes.io/") {
			role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
			if role != "" {
				roles = append(roles, role)
			}
		}
	}
	if len(roles) == 0 {
		roles = append(roles, "worker")
	}
	return roles
}

// formatMemoryQuantity 格式化内存为易读格式
// 注意：Kubernetes metrics API 有时会用 "m" (milli) 后缀返回内存值（如 "7615829333m"），
// 但对于内存来说，这实际上就是字节数（7615829333 字节 ≈ 7.09 GB），不是真正的毫字节。
// 这里使用 MilliValue() 获取原始数值作为字节数来处理这种情况。
func formatMemoryQuantity(q *resource.Quantity) string {
	if q == nil {
		return "0"
	}

	// 检查是否是 milli 格式（如 "7615829333m"）
	// 如果是，直接用 MilliValue() 作为字节数
	// 否则，用 Value() 作为字节数
	str := q.String()
	var bytes float64
	if len(str) > 0 && str[len(str)-1] == 'm' {
		// milli 格式，直接用 MilliValue 作为字节数
		bytes = float64(q.MilliValue())
	} else {
		// 正常格式，用 Value 作为字节数
		bytes = float64(q.Value())
	}

	if bytes <= 0 {
		return "0"
	}

	// 转换为 Gi
	if bytes >= 1024*1024*1024 {
		gi := bytes / (1024 * 1024 * 1024)
		if gi >= 10 {
			return fmt.Sprintf("%.1fGi", gi)
		}
		return fmt.Sprintf("%.2fGi", gi)
	}
	// 转换为 Mi
	if bytes >= 1024*1024 {
		mi := bytes / (1024 * 1024)
		if mi >= 10 {
			return fmt.Sprintf("%.1fMi", mi)
		}
		return fmt.Sprintf("%.2fMi", mi)
	}
	// 转换为 Ki
	if bytes >= 1024 {
		ki := bytes / 1024
		return fmt.Sprintf("%.1fKi", ki)
	}
	return fmt.Sprintf("%.0fB", bytes)
}

// formatCPU 格式化 CPU 为毫核(m)格式
func formatCPU(q *resource.Quantity) string {
	if q == nil {
		return "0"
	}
	// 转换为毫核
	milliValue := q.MilliValue()
	if milliValue >= 1000 {
		// 大于等于 1 核，显示为核数
		cores := float64(milliValue) / 1000
		if cores == float64(int64(cores)) {
			return fmt.Sprintf("%d", int64(cores))
		}
		return fmt.Sprintf("%.2f", cores)
	}
	// 小于 1 核，显示为毫核
	return fmt.Sprintf("%dm", milliValue)
}

// roundFloat 四舍五入到指定小数位
func roundFloat(val float64, precision int) float64 {
	p := float64(1)
	for i := 0; i < precision; i++ {
		p *= 10
	}
	return float64(int(val*p+0.5)) / p
}
