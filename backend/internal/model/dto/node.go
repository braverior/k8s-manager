package dto

// NodeResponse 节点详情响应
type NodeResponse struct {
	// 节点名称
	Name string `json:"name"`
	// 节点状态
	Status string `json:"status"`
	// 节点角色（master/worker）
	Roles []string `json:"roles"`
	// 内部 IP
	InternalIP string `json:"internal_ip,omitempty"`
	// 外部 IP
	ExternalIP string `json:"external_ip,omitempty"`
	// Kubelet 版本
	KubeletVersion string `json:"kubelet_version"`
	// 内核版本
	KernelVersion string `json:"kernel_version"`
	// 操作系统镜像
	OSImage string `json:"os_image"`
	// 操作系统
	OS string `json:"os"`
	// 架构
	Architecture string `json:"architecture"`
	// 容器运行时
	ContainerRuntime string `json:"container_runtime"`
	// 创建时间
	CreatedAt string `json:"created_at"`
	// 资源容量
	Capacity NodeResourceInfo `json:"capacity"`
	// 可分配资源
	Allocatable NodeResourceInfo `json:"allocatable"`
	// 资源使用情况（需要 metrics-server）
	Usage *NodeUsageInfo `json:"usage,omitempty"`
	// 节点条件
	Conditions []NodeCondition `json:"conditions"`
	// 节点标签
	Labels map[string]string `json:"labels,omitempty"`
	// 节点污点
	Taints []NodeTaint `json:"taints,omitempty"`
	// Pod 数量
	PodCount int `json:"pod_count"`
}

// NodeResourceInfo 节点资源信息
type NodeResourceInfo struct {
	// CPU
	CPU string `json:"cpu"`
	// 内存
	Memory string `json:"memory"`
	// 可调度 Pod 数
	Pods string `json:"pods"`
	// 临时存储
	EphemeralStorage string `json:"ephemeral_storage,omitempty"`
}

// NodeUsageInfo 节点资源使用情况
type NodeUsageInfo struct {
	// CPU 使用量
	CPU string `json:"cpu"`
	// CPU 使用百分比
	CPUPercentage float64 `json:"cpu_percentage"`
	// 内存使用量
	Memory string `json:"memory"`
	// 内存使用百分比
	MemoryPercentage float64 `json:"memory_percentage"`
}

// NodeCondition 节点条件
type NodeCondition struct {
	// 条件类型
	Type string `json:"type"`
	// 条件状态
	Status string `json:"status"`
	// 原因
	Reason string `json:"reason,omitempty"`
	// 消息
	Message string `json:"message,omitempty"`
	// 最后心跳时间
	LastHeartbeatTime string `json:"last_heartbeat_time,omitempty"`
}

// NodeTaint 节点污点
type NodeTaint struct {
	// 键
	Key string `json:"key"`
	// 值
	Value string `json:"value,omitempty"`
	// 效果
	Effect string `json:"effect"`
}

// NodeListItem 节点列表项（简化版）
type NodeListItem struct {
	// 节点名称
	Name string `json:"name"`
	// 节点状态
	Status string `json:"status"`
	// 节点角色
	Roles []string `json:"roles"`
	// 内部 IP
	InternalIP string `json:"internal_ip,omitempty"`
	// Kubelet 版本
	KubeletVersion string `json:"kubelet_version"`
	// CPU 容量
	CPUCapacity string `json:"cpu_capacity"`
	// 内存容量
	MemoryCapacity string `json:"memory_capacity"`
	// CPU 使用量
	CPUUsage string `json:"cpu_usage,omitempty"`
	// CPU 使用百分比
	CPUPercentage float64 `json:"cpu_percentage,omitempty"`
	// 内存使用量
	MemoryUsage string `json:"memory_usage,omitempty"`
	// 内存使用百分比
	MemoryPercentage float64 `json:"memory_percentage,omitempty"`
	// Pod 数量
	PodCount int `json:"pod_count"`
	// 创建时间
	CreatedAt string `json:"created_at"`
}
