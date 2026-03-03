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
	Conditions      []PodCondition         `json:"conditions,omitempty"`
	Metrics         *PodMetricsResponse    `json:"metrics,omitempty"`
}

// ContainerStatus 容器状态
type ContainerStatus struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restart_count"`
	State        string `json:"state"`
	Reason       string `json:"reason,omitempty"`
	Message      string `json:"message,omitempty"`
	Image        string `json:"image"`
	StartedAt    string `json:"started_at,omitempty"`
	ExitCode     *int32 `json:"exit_code,omitempty"`
	LastState    string `json:"last_state,omitempty"`
	LastReason   string `json:"last_reason,omitempty"`
	LastMessage  string `json:"last_message,omitempty"`
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

// PodCondition Pod 条件
type PodCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	Reason             string `json:"reason,omitempty"`
	Message            string `json:"message,omitempty"`
	LastTransitionTime string `json:"last_transition_time,omitempty"`
}

// PodLogResponse Pod 日志响应
type PodLogResponse struct {
	PodName       string `json:"pod_name"`
	ContainerName string `json:"container_name"`
	Logs          string `json:"logs"`
}

// PodEvent Pod 事件
type PodEvent struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Source    string `json:"source"`
	Count     int32  `json:"count"`
	FirstTime string `json:"first_time"`
	LastTime  string `json:"last_time"`
}
