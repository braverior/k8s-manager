package service

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

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

// List 列出命名空间下的所有 Pod
func (s *PodService) List(ctx context.Context, clusterName, namespace string) ([]dto.PodResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	metricsClient, _ := s.clientManager.GetMetricsClient(clusterName)
	op := operator.NewPodOperator(client, metricsClient)

	pods, err := op.List(ctx, namespace)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "获取 Pod 列表失败")
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

	return s.toResponses(pods, metricsMap), nil
}

// ListByDeployment 根据 Deployment 名称查询关联的 Pod
func (s *PodService) ListByDeployment(ctx context.Context, clusterName, namespace, deploymentName string) ([]dto.PodResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	// 先获取 Deployment 的 selector
	deployOp := operator.NewDeploymentOperator(client)
	deploy, err := deployOp.Get(ctx, namespace, deploymentName)
	if err != nil {
		return nil, apperrors.Wrap(err, 404, 404, "Deployment 不存在")
	}

	// 构建标签选择器
	labelSelector := ""
	for k, v := range deploy.Spec.Selector.MatchLabels {
		if labelSelector != "" {
			labelSelector += ","
		}
		labelSelector += fmt.Sprintf("%s=%s", k, v)
	}

	metricsClient, _ := s.clientManager.GetMetricsClient(clusterName)
	podOp := operator.NewPodOperator(client, metricsClient)

	pods, err := podOp.ListByLabels(ctx, namespace, labelSelector)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "获取 Pod 列表失败")
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

func (s *PodService) toResponses(pods []corev1.Pod, metricsMap map[string]*operator.PodMetrics) []dto.PodResponse {
	var responses []dto.PodResponse
	for _, pod := range pods {
		responses = append(responses, *s.toResponse(&pod, metricsMap[pod.Name]))
	}
	return responses
}

func (s *PodService) toResponse(pod *corev1.Pod, metrics *operator.PodMetrics) *dto.PodResponse {
	resp := &dto.PodResponse{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Phase:     string(pod.Status.Phase),
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
		} else if cs.State.Terminated != nil {
			container.State = "Terminated"
			container.Reason = cs.State.Terminated.Reason
		}

		resp.Containers = append(resp.Containers, container)
	}

	// 添加资源指标
	if metrics != nil {
		resp.Metrics = &dto.PodMetricsResponse{
			Containers: make([]dto.ContainerMetricsResponse, 0, len(metrics.Containers)),
		}
		for _, cm := range metrics.Containers {
			resp.Metrics.Containers = append(resp.Metrics.Containers, dto.ContainerMetricsResponse{
				Name:   cm.Name,
				CPU:    cm.CPU,
				Memory: cm.Memory,
			})
		}
	}

	return resp
}
