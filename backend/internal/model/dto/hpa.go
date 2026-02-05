package dto

// HPAResponse HPA 详细响应
type HPAResponse struct {
	// 资源名称
	Name string `json:"name"`
	// 命名空间
	Namespace string `json:"namespace"`
	// 目标资源
	ScaleTargetRef HPAScaleTargetRef `json:"scale_target_ref"`
	// 最小副本数
	MinReplicas int32 `json:"min_replicas"`
	// 最大副本数
	MaxReplicas int32 `json:"max_replicas"`
	// 当前副本数
	CurrentReplicas int32 `json:"current_replicas"`
	// 期望副本数
	DesiredReplicas int32 `json:"desired_replicas"`
	// CPU 统计
	CPU *HPAResourceMetric `json:"cpu,omitempty"`
	// 内存统计
	Memory *HPAResourceMetric `json:"memory,omitempty"`
	// 指标配置和当前值（包含所有类型指标）
	Metrics []HPAMetricStatus `json:"metrics,omitempty"`
	// 条件状态
	Conditions []HPACondition `json:"conditions,omitempty"`
	// 完整的 YAML 内容
	YAML string `json:"yaml"`
}

// HPAResourceMetric HPA 资源指标（CPU/内存）
type HPAResourceMetric struct {
	// 目标类型: Utilization, AverageValue
	TargetType string `json:"target_type"`
	// 目标利用率百分比（当 TargetType 为 Utilization 时）
	TargetUtilization *int32 `json:"target_utilization,omitempty"`
	// 目标值（当 TargetType 为 AverageValue 时，如 "200Mi", "500m"）
	TargetAverageValue string `json:"target_average_value,omitempty"`
	// 当前利用率百分比
	CurrentUtilization *int32 `json:"current_utilization,omitempty"`
	// 当前平均值（如 "150Mi", "300m"）
	CurrentAverageValue string `json:"current_average_value,omitempty"`
}

// HPAScaleTargetRef HPA 目标资源引用
type HPAScaleTargetRef struct {
	// API 版本
	APIVersion string `json:"api_version"`
	// 资源类型
	Kind string `json:"kind"`
	// 资源名称
	Name string `json:"name"`
}

// HPAMetricStatus HPA 指标状态
type HPAMetricStatus struct {
	// 指标类型: Resource, Pods, Object, External
	Type string `json:"type"`
	// 资源名称（cpu, memory 等）
	Name string `json:"name,omitempty"`
	// 目标类型: Utilization, AverageValue, Value
	TargetType string `json:"target_type,omitempty"`
	// 目标值
	TargetValue string `json:"target_value,omitempty"`
	// 目标利用率百分比
	TargetUtilization *int32 `json:"target_utilization,omitempty"`
	// 当前值
	CurrentValue string `json:"current_value,omitempty"`
	// 当前利用率百分比
	CurrentUtilization *int32 `json:"current_utilization,omitempty"`
}

// HPACondition HPA 条件状态
type HPACondition struct {
	// 条件类型
	Type string `json:"type"`
	// 状态
	Status string `json:"status"`
	// 原因
	Reason string `json:"reason,omitempty"`
	// 消息
	Message string `json:"message,omitempty"`
	// 最后转换时间
	LastTransitionTime string `json:"last_transition_time,omitempty"`
}

// HPASummary HPA 概要信息（用于 Dashboard）
type HPASummary struct {
	// 资源名称
	Name string `json:"name"`
	// 命名空间
	Namespace string `json:"namespace"`
	// 目标资源类型
	TargetKind string `json:"target_kind"`
	// 目标资源名称
	TargetName string `json:"target_name"`
	// 最小副本数
	MinReplicas int32 `json:"min_replicas"`
	// 最大副本数
	MaxReplicas int32 `json:"max_replicas"`
	// 当前副本数
	CurrentReplicas int32 `json:"current_replicas"`
	// 期望副本数
	DesiredReplicas int32 `json:"desired_replicas"`
	// CPU 目标利用率
	CPUTargetUtilization *int32 `json:"cpu_target_utilization,omitempty"`
	// CPU 当前利用率
	CPUCurrentUtilization *int32 `json:"cpu_current_utilization,omitempty"`
	// 内存目标利用率
	MemoryTargetUtilization *int32 `json:"memory_target_utilization,omitempty"`
	// 内存当前利用率
	MemoryCurrentUtilization *int32 `json:"memory_current_utilization,omitempty"`
}
