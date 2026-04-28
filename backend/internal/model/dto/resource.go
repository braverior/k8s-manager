package dto

// ResourceRequest 通用资源请求，接收 YAML 或 Base64 内容
type ResourceRequest struct {
	// YAML 内容（与 Content 二选一）
	YAML string `json:"yaml"`
	// Base64 编码的 YAML 内容（与 YAML 二选一）
	Content string `json:"content"`
	// K8s resourceVersion，用于乐观并发控制；更新时传入，后端比对不一致则返回 409
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

// ResourceQuery 分页和搜索查询参数
type ResourceQuery struct {
	Search   string `form:"search"`
	Status   string `form:"status"` // Pod 状态分类筛选: healthy / pending / error
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=50"`
}

// ResourceYAMLResponse 通用资源 YAML 响应
type ResourceYAMLResponse struct {
	// 资源名称
	Name string `json:"name"`
	// 命名空间
	Namespace string `json:"namespace"`
	// 完整的 YAML 内容
	YAML string `json:"yaml"`
	// K8s resourceVersion，前端编辑后回传用于冲突检测
	ResourceVersion string `json:"resourceVersion,omitempty"`
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
	// 资源请求与限制（来自 Pod.Spec.Containers[].Resources），空字符串表示未设置
	CPURequest    string `json:"cpu_request,omitempty"`
	CPULimit      string `json:"cpu_limit,omitempty"`
	MemoryRequest string `json:"memory_request,omitempty"`
	MemoryLimit   string `json:"memory_limit,omitempty"`
}

// PodMetricsResponse Pod 指标响应
type PodMetricsResponse struct {
	Containers []ContainerMetricsResponse `json:"containers"`
	// Pod 级聚合：便于前端直接用于进度条计算
	CPUMillis        int64 `json:"cpu_millis"`         // 当前 CPU 用量（毫核）
	MemoryBytes      int64 `json:"memory_bytes"`       // 当前内存用量（字节）
	CPULimitMillis   int64 `json:"cpu_limit_millis"`   // 所有容器 CPU limit 之和（毫核），仅 HasCPULimit=true 时有意义
	MemoryLimitBytes int64 `json:"memory_limit_bytes"` // 所有容器 Memory limit 之和（字节），仅 HasMemoryLimit=true 时有意义
	HasCPULimit      bool  `json:"has_cpu_limit"`      // 所有容器是否均设置了 CPU limit
	HasMemoryLimit   bool  `json:"has_memory_limit"`   // 所有容器是否均设置了 Memory limit
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
