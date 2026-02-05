package dto

// ResourceRequest 通用资源请求，接收 YAML 或 Base64 内容
type ResourceRequest struct {
	// YAML 内容（与 Content 二选一）
	YAML string `json:"yaml"`
	// Base64 编码的 YAML 内容（与 YAML 二选一）
	Content string `json:"content"`
}

// ResourceYAMLResponse 通用资源 YAML 响应
type ResourceYAMLResponse struct {
	// 资源名称
	Name string `json:"name"`
	// 命名空间
	Namespace string `json:"namespace"`
	// 完整的 YAML 内容
	YAML string `json:"yaml"`
}

// PodResponse Pod 响应
type PodResponse struct {
	Name            string                 `json:"name"`
	Namespace       string                 `json:"namespace"`
	Phase           string                 `json:"phase"`
	PodIP           string                 `json:"pod_ip,omitempty"`
	HostIP          string                 `json:"host_ip,omitempty"`
	NodeName        string                 `json:"node_name,omitempty"`
	ReadyContainers int                    `json:"ready_containers"`
	TotalContainers int                    `json:"total_containers"`
	RestartCount    int32                  `json:"restart_count"`
	CreatedAt       string                 `json:"created_at"`
	Containers      []ContainerStatus      `json:"containers"`
	Metrics         *PodMetricsResponse    `json:"metrics,omitempty"`
}

// ContainerStatus 容器状态
type ContainerStatus struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restart_count"`
	State        string `json:"state"`
	Reason       string `json:"reason,omitempty"`
	Image        string `json:"image"`
	StartedAt    string `json:"started_at,omitempty"`
}

// PodMetricsResponse Pod 指标响应
type PodMetricsResponse struct {
	Containers []ContainerMetricsResponse `json:"containers"`
}

// ContainerMetricsResponse 容器指标响应
type ContainerMetricsResponse struct {
	Name   string `json:"name"`
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}
