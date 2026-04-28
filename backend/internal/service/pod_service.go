package service

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s_api_server/internal/k8s"
	"k8s_api_server/internal/k8s/operator"
	"k8s_api_server/internal/model/dto"
	apperrors "k8s_api_server/internal/pkg/errors"
)

type PodService struct {
	clientManager *k8s.ClientManager
}

func NewPodService(clientManager *k8s.ClientManager) *PodService {
	return &PodService{
		clientManager: clientManager,
	}
}

// List 列出命名空间下的所有 Pod（支持分页和搜索）
func (s *PodService) List(ctx context.Context, clusterName, namespace string, query *dto.ResourceQuery) ([]dto.PodResponse, int64, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	metricsClient, _ := s.clientManager.GetMetricsClient(clusterName)
	op := operator.NewPodOperator(client, metricsClient)

	pods, err := op.List(ctx, namespace)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, 500, 500, "获取 Pod 列表失败")
	}

	// 获取所有 Pod 的指标
	metricsMap := make(map[string]*operator.PodMetrics)
	if metricsClient != nil {
		if metrics, err := op.ListMetrics(ctx, namespace); err == nil {
			for i := range metrics {
				metricsMap[metrics[i].Name] = &metrics[i]
			}
		}
	}

	responses := s.toResponses(pods, metricsMap)

	// 搜索过滤
	if query != nil && query.Search != "" {
		search := strings.ToLower(query.Search)
		filtered := make([]dto.PodResponse, 0)
		for _, r := range responses {
			if strings.Contains(strings.ToLower(r.Name), search) {
				filtered = append(filtered, r)
			}
		}
		responses = filtered
	}

	// 状态分类过滤
	if query != nil && query.Status != "" {
		filtered := make([]dto.PodResponse, 0)
		for _, r := range responses {
			if classifyPodStatus(r.Phase) == query.Status {
				filtered = append(filtered, r)
			}
		}
		responses = filtered
	}

	total := int64(len(responses))

	// 分页
	if query != nil && query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		if offset >= len(responses) {
			return []dto.PodResponse{}, total, nil
		}
		end := offset + query.PageSize
		if end > len(responses) {
			end = len(responses)
		}
		responses = responses[offset:end]
	}

	return responses, total, nil
}

// ListByDeployment 根据 Deployment 名称查询关联的 Pod
func (s *PodService) ListByDeployment(ctx context.Context, clusterName, namespace, deploymentName string) ([]dto.PodResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	// 通过 ownerReferences 查找属于该 Deployment 的 ReplicaSet
	rsList, err := client.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "获取 ReplicaSet 列表失败")
	}
	rsNames := make(map[string]bool)
	for _, rs := range rsList.Items {
		for _, owner := range rs.OwnerReferences {
			if owner.Kind == "Deployment" && owner.Name == deploymentName {
				rsNames[rs.Name] = true
				break
			}
		}
	}

	metricsClient, _ := s.clientManager.GetMetricsClient(clusterName)
	podOp := operator.NewPodOperator(client, metricsClient)

	// 获取所有 Pod，通过 ownerReferences 过滤属于目标 Deployment 的 Pod
	allPods, err := podOp.List(ctx, namespace)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "获取 Pod 列表失败")
	}

	var pods []corev1.Pod
	for _, pod := range allPods {
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "ReplicaSet" && rsNames[owner.Name] {
				pods = append(pods, pod)
				break
			}
		}
	}

	// 获取所有 Pod 的指标
	metricsMap := make(map[string]*operator.PodMetrics)
	if metricsClient != nil {
		if metrics, err := podOp.ListMetrics(ctx, namespace); err == nil {
			for i := range metrics {
				metricsMap[metrics[i].Name] = &metrics[i]
			}
		}
	}

	return s.toResponses(pods, metricsMap), nil
}

// Get 获取单个 Pod 详情
func (s *PodService) Get(ctx context.Context, clusterName, namespace, name string) (*dto.PodResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	metricsClient, _ := s.clientManager.GetMetricsClient(clusterName)
	op := operator.NewPodOperator(client, metricsClient)

	pod, err := op.Get(ctx, namespace, name)
	if err != nil {
		return nil, apperrors.Wrap(err, 404, 404, "Pod 不存在")
	}

	// 获取指标
	var metrics *operator.PodMetrics
	if metricsClient != nil {
		metrics, _ = op.GetMetrics(ctx, namespace, name)
	}

	return s.toResponse(pod, metrics), nil
}

// Delete 删除 Pod（用于重启）
func (s *PodService) Delete(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := operator.NewPodOperator(client, nil)
	if err := op.Delete(ctx, namespace, name); err != nil {
		return apperrors.Wrap(err, 500, 500, "删除 Pod 失败")
	}

	return nil
}

// GetLogs 获取 Pod 容器日志
func (s *PodService) GetLogs(ctx context.Context, clusterName, namespace, name, container string, tailLines int64, previous bool, timestamps bool) (*dto.PodLogResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := operator.NewPodOperator(client, nil)

	// 如果未指定容器，获取第一个容器名
	if container == "" {
		pod, err := op.Get(ctx, namespace, name)
		if err != nil {
			return nil, apperrors.Wrap(err, 404, 404, "Pod 不存在")
		}
		if len(pod.Spec.Containers) > 0 {
			container = pod.Spec.Containers[0].Name
		}
	}

	logs, err := op.GetLogs(ctx, namespace, name, container, tailLines, previous, timestamps)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "获取 Pod 日志失败")
	}

	return &dto.PodLogResponse{
		PodName:       name,
		ContainerName: container,
		Logs:          logs,
	}, nil
}

// GetEvents 获取 Pod 相关事件
func (s *PodService) GetEvents(ctx context.Context, clusterName, namespace, name string) ([]dto.PodEvent, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := operator.NewPodOperator(client, nil)
	events, err := op.GetEvents(ctx, namespace, name)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "获取 Pod 事件失败")
	}

	result := make([]dto.PodEvent, 0, len(events))
	for _, e := range events {
		source := e.Source.Component
		if e.Source.Host != "" {
			source += "/" + e.Source.Host
		}

		firstTime := ""
		if !e.FirstTimestamp.IsZero() {
			firstTime = e.FirstTimestamp.Format("2006-01-02T15:04:05Z")
		}
		lastTime := ""
		if !e.LastTimestamp.IsZero() {
			lastTime = e.LastTimestamp.Format("2006-01-02T15:04:05Z")
		}

		result = append(result, dto.PodEvent{
			Type:      e.Type,
			Reason:    e.Reason,
			Message:   e.Message,
			Source:    source,
			Count:     e.Count,
			FirstTime: firstTime,
			LastTime:  lastTime,
		})
	}

	return result, nil
}

func (s *PodService) toResponses(pods []corev1.Pod, metricsMap map[string]*operator.PodMetrics) []dto.PodResponse {
	var responses []dto.PodResponse
	for _, pod := range pods {
		responses = append(responses, *s.toResponse(&pod, metricsMap[pod.Name]))
	}
	return responses
}

// classifyPodStatus 将展示状态归类为 healthy / pending / error
func classifyPodStatus(phase string) string {
	switch phase {
	case "Running", "Succeeded":
		return "healthy"
	case "Pending", "ContainerCreating", "PodInitializing", "Terminating":
		return "pending"
	case "Failed", "Error", "CrashLoopBackOff", "OOMKilled",
		"ImagePullBackOff", "ErrImagePull",
		"CreateContainerError", "CreateContainerConfigError",
		"InvalidImageName":
		return "error"
	}
	if strings.HasPrefix(phase, "Init:") {
		return "pending"
	}
	return "unknown"
}

// getPodDisplayStatus 推导 Pod 的展示状态，参考 kubectl 的逻辑：
// 优先级 Terminating > Init 容器 Waiting/Terminated Reason > 容器 Waiting/Terminated Reason > Phase
func getPodDisplayStatus(pod *corev1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}

	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
			return "Init:" + cs.State.Waiting.Reason
		}
		if cs.State.Terminated != nil && cs.State.Terminated.Reason != "" && cs.State.Terminated.ExitCode != 0 {
			return "Init:" + cs.State.Terminated.Reason
		}
	}

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
			return cs.State.Waiting.Reason
		}
		if cs.State.Terminated != nil && cs.State.Terminated.Reason != "" {
			return cs.State.Terminated.Reason
		}
	}

	return string(pod.Status.Phase)
}

func (s *PodService) toResponse(pod *corev1.Pod, metrics *operator.PodMetrics) *dto.PodResponse {
	resp := &dto.PodResponse{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Phase:     getPodDisplayStatus(pod),
		PodIP:     pod.Status.PodIP,
		HostIP:    pod.Status.HostIP,
		NodeName:  pod.Spec.NodeName,
		CreatedAt: pod.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}

	// 计算就绪容器数
	readyCount := 0
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			readyCount++
		}
	}
	resp.ReadyContainers = readyCount
	resp.TotalContainers = len(pod.Spec.Containers)

	// 计算重启次数
	restartCount := int32(0)
	for _, cs := range pod.Status.ContainerStatuses {
		restartCount += cs.RestartCount
	}
	resp.RestartCount = restartCount

	// 按容器名索引 Spec 中的资源配置（requests/limits 来自 Spec）
	specByName := make(map[string]corev1.Container, len(pod.Spec.Containers))
	for _, c := range pod.Spec.Containers {
		specByName[c.Name] = c
	}

	// Pod 级 limit 聚合：所有容器都设置了 CPU/Mem limit 才算 "有完整 limit"
	var (
		cpuLimitSum    int64
		memLimitSum    int64
		hasCPULimit    = len(pod.Spec.Containers) > 0
		hasMemoryLimit = len(pod.Spec.Containers) > 0
	)
	for _, c := range pod.Spec.Containers {
		cpuLim := c.Resources.Limits.Cpu()
		if cpuLim == nil || cpuLim.IsZero() {
			hasCPULimit = false
		} else {
			cpuLimitSum += cpuLim.MilliValue()
		}
		memLim := c.Resources.Limits.Memory()
		if memLim == nil || memLim.IsZero() {
			hasMemoryLimit = false
		} else {
			memLimitSum += memLim.Value()
		}
	}

	// 容器信息
	resp.Containers = make([]dto.ContainerStatus, 0, len(pod.Status.ContainerStatuses))
	for _, cs := range pod.Status.ContainerStatuses {
		container := dto.ContainerStatus{
			Name:         cs.Name,
			Ready:        cs.Ready,
			RestartCount: cs.RestartCount,
			Image:        cs.Image,
		}

		// 容器状态
		if cs.State.Running != nil {
			container.State = "Running"
			container.StartedAt = cs.State.Running.StartedAt.Format("2006-01-02T15:04:05Z")
		} else if cs.State.Waiting != nil {
			container.State = "Waiting"
			container.Reason = cs.State.Waiting.Reason
			container.Message = cs.State.Waiting.Message
		} else if cs.State.Terminated != nil {
			container.State = "Terminated"
			container.Reason = cs.State.Terminated.Reason
			container.Message = cs.State.Terminated.Message
			exitCode := cs.State.Terminated.ExitCode
			container.ExitCode = &exitCode
		}

		// 上一次容器状态
		if cs.LastTerminationState.Running != nil {
			container.LastState = "Running"
		} else if cs.LastTerminationState.Waiting != nil {
			container.LastState = "Waiting"
			container.LastReason = cs.LastTerminationState.Waiting.Reason
			container.LastMessage = cs.LastTerminationState.Waiting.Message
		} else if cs.LastTerminationState.Terminated != nil {
			container.LastState = "Terminated"
			container.LastReason = cs.LastTerminationState.Terminated.Reason
			container.LastMessage = cs.LastTerminationState.Terminated.Message
		}

		// 注入 requests/limits（来自 Spec）
		if spec, ok := specByName[cs.Name]; ok {
			if q := spec.Resources.Requests.Cpu(); q != nil && !q.IsZero() {
				container.CPURequest = q.String()
			}
			if q := spec.Resources.Limits.Cpu(); q != nil && !q.IsZero() {
				container.CPULimit = q.String()
			}
			if q := spec.Resources.Requests.Memory(); q != nil && !q.IsZero() {
				container.MemoryRequest = q.String()
			}
			if q := spec.Resources.Limits.Memory(); q != nil && !q.IsZero() {
				container.MemoryLimit = q.String()
			}
		}

		resp.Containers = append(resp.Containers, container)
	}

	// Pod 条件
	if len(pod.Status.Conditions) > 0 {
		resp.Conditions = make([]dto.PodCondition, 0, len(pod.Status.Conditions))
		for _, cond := range pod.Status.Conditions {
			pc := dto.PodCondition{
				Type:    string(cond.Type),
				Status:  string(cond.Status),
				Reason:  cond.Reason,
				Message: cond.Message,
			}
			if !cond.LastTransitionTime.IsZero() {
				pc.LastTransitionTime = cond.LastTransitionTime.Format("2006-01-02T15:04:05Z")
			}
			resp.Conditions = append(resp.Conditions, pc)
		}
	}

	// 添加资源指标
	if metrics != nil {
		var cpuSum, memSum int64
		containers := make([]dto.ContainerMetricsResponse, 0, len(metrics.Containers))
		for _, cm := range metrics.Containers {
			containers = append(containers, dto.ContainerMetricsResponse{
				Name:   cm.Name,
				CPU:    cm.CPU,
				Memory: cm.Memory,
			})
			cpuSum += cm.CPUMillis
			memSum += cm.MemoryBytes
		}
		resp.Metrics = &dto.PodMetricsResponse{
			Containers:       containers,
			CPUMillis:        cpuSum,
			MemoryBytes:      memSum,
			HasCPULimit:      hasCPULimit,
			HasMemoryLimit:   hasMemoryLimit,
			CPULimitMillis:   cpuLimitSum,
			MemoryLimitBytes: memLimitSum,
		}
	}

	return resp
}
