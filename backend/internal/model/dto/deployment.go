package dto

// DeploymentResponse Deployment 响应
type DeploymentResponse struct {
	// 名称
	Name string `json:"name"`
	// 命名空间
	Namespace string `json:"namespace"`
	// 期望副本数
	Replicas int32 `json:"replicas"`
	// 就绪副本数
	ReadyReplicas int32 `json:"ready_replicas"`
	// 已更新副本数（滚动更新进度）
	UpdatedReplicas int32 `json:"updated_replicas"`
	// 可用副本数
	AvailableReplicas int32 `json:"available_replicas"`
	// Pod 数量
	PodCount int `json:"pod_count"`
	// 各状态 Pod 数量 (Running, Pending, Succeeded, Failed, Unknown)
	PodStatusCounts map[string]int `json:"pod_status_counts"`
	// YAML 内容
	YAML string `json:"yaml"`
	// K8s resourceVersion，前端编辑后回传用于冲突检测
	ResourceVersion string `json:"resourceVersion,omitempty"`
}
