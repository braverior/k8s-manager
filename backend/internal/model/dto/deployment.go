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
	// 可用副本数
	AvailableReplicas int32 `json:"available_replicas"`
	// Pod 数量
	PodCount int `json:"pod_count"`
	// YAML 内容
	YAML string `json:"yaml"`
}
