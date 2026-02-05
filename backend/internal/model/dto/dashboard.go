package dto

// DashboardResponse 集群 Dashboard 概览响应
type DashboardResponse struct {
	// 集群名称
	ClusterName string `json:"cluster_name"`
	// 集群状态
	Status string `json:"status"`
	// Kubernetes 版本
	Version string `json:"version,omitempty"`
	// 资源统计
	Resources ResourceStats `json:"resources"`
	// 集群资源容量（总量）
	Capacity *ClusterCapacity `json:"capacity,omitempty"`
	// 集群资源使用（已用量）
	Usage *ClusterUsage `json:"usage,omitempty"`
	// HPA 概要列表
	HPASummaries []HPASummary `json:"hpa_summaries,omitempty"`
}

// ResourceStats 资源数量统计
type ResourceStats struct {
	// 节点数量
	Nodes int `json:"nodes"`
	// 命名空间数量
	Namespaces int `json:"namespaces"`
	// Pod 数量
	Pods PodStats `json:"pods"`
	// Deployment 数量
	Deployments int `json:"deployments"`
	// Service 数量
	Services int `json:"services"`
	// ConfigMap 数量
	ConfigMaps int `json:"configmaps"`
	// HPA 数量
	HPAs int `json:"hpas"`
}

// PodStats Pod 状态统计
type PodStats struct {
	// 总数
	Total int `json:"total"`
	// 运行中
	Running int `json:"running"`
	// 等待中
	Pending int `json:"pending"`
	// 成功完成
	Succeeded int `json:"succeeded"`
	// 失败
	Failed int `json:"failed"`
}

// ClusterCapacity 集群资源容量（总量）
type ClusterCapacity struct {
	// CPU 总容量（如 "8000m" 或 "8"）
	CPU string `json:"cpu"`
	// 内存总容量（如 "32Gi"）
	Memory string `json:"memory"`
	// 可调度 Pod 总数
	Pods int64 `json:"pods"`
}

// ClusterUsage 集群资源使用情况（已用量）
type ClusterUsage struct {
	// CPU 已使用量（如 "2500m"）
	CPU string `json:"cpu"`
	// CPU 使用百分比
	CPUPercentage float64 `json:"cpu_percentage"`
	// 内存已使用量（如 "8Gi"）
	Memory string `json:"memory"`
	// 内存使用百分比
	MemoryPercentage float64 `json:"memory_percentage"`
}
