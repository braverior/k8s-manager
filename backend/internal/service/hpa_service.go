package service

import (
	"context"

	autoscalingv2 "k8s.io/api/autoscaling/v2"

	"k8s_api_server/internal/k8s"
	k8soperator "k8s_api_server/internal/k8s/operator"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/model/entity"
	apperrors "k8s_api_server/internal/pkg/errors"
	"k8s_api_server/internal/pkg/utils"
	"k8s_api_server/internal/repository"
)

type HPAService struct {
	clientManager *k8s.ClientManager
	historyRepo   *repository.HistoryRepository
}

func NewHPAService(clientManager *k8s.ClientManager, historyRepo *repository.HistoryRepository) *HPAService {
	return &HPAService{
		clientManager: clientManager,
		historyRepo:   historyRepo,
	}
}

func (s *HPAService) List(ctx context.Context, clusterName, namespace string) ([]dto.HPAResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := k8soperator.NewHPAOperator(client)
	hpas, err := op.List(ctx, namespace)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "获取 HPA 列表失败: "+err.Error())
	}

	responses := make([]dto.HPAResponse, 0, len(hpas))
	for _, hpa := range hpas {
		yamlContent, err := utils.EncodeToYAML(&hpa, "autoscaling/v2", "HorizontalPodAutoscaler")
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "转换 YAML 失败")
		}
		responses = append(responses, s.buildHPAResponse(&hpa, yamlContent))
	}
	return responses, nil
}

func (s *HPAService) Get(ctx context.Context, clusterName, namespace, name string) (*dto.HPAResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := k8soperator.NewHPAOperator(client)
	hpa, err := op.Get(ctx, namespace, name)
	if err != nil {
		return nil, apperrors.Wrap(err, 404, 404, "HPA 不存在")
	}

	yamlContent, err := utils.EncodeToYAML(hpa, "autoscaling/v2", "HorizontalPodAutoscaler")
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "转换 YAML 失败")
	}

	resp := s.buildHPAResponse(hpa, yamlContent)
	return &resp, nil
}

func (s *HPAService) Apply(ctx context.Context, clusterName, namespace string, req *dto.ResourceRequest, operator string) (*dto.ResourceYAMLResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	content, err := utils.ParseYAMLOrBase64(req.YAML, req.Content)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "解析请求内容失败")
	}

	obj, err := utils.DecodeYAML(content)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "解析 YAML 失败")
	}

	hpa, ok := obj.(*autoscalingv2.HorizontalPodAutoscaler)
	if !ok {
		return nil, apperrors.Wrap(nil, 400, 400, "YAML 内容不是有效的 HPA")
	}

	hpa.Namespace = namespace
	op := k8soperator.NewHPAOperator(client)

	var result *autoscalingv2.HorizontalPodAutoscaler
	var operation string
	exists, _ := op.Exists(ctx, namespace, hpa.Name)
	if exists {
		existing, err := op.Get(ctx, namespace, hpa.Name)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "获取现有 HPA 失败")
		}
		if req.ResourceVersion != "" && req.ResourceVersion != existing.ResourceVersion {
			return nil, apperrors.New(409, 409, "资源已被其他用户修改，请刷新后重试")
		}
		hpa.ResourceVersion = existing.ResourceVersion
		result, err = op.Update(ctx, namespace, hpa)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "更新 HPA 失败")
		}
		operation = "update"
	} else {
		result, err = op.Create(ctx, namespace, hpa)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "创建 HPA 失败")
		}
		operation = "create"
	}

	_ = s.saveHistory(ctx, clusterName, namespace, result.Name, string(content), operation, operator)

	yamlContent, err := utils.EncodeToYAML(result, "autoscaling/v2", "HorizontalPodAutoscaler")
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "转换 YAML 失败")
	}

	return &dto.ResourceYAMLResponse{
		Name:            result.Name,
		Namespace:       result.Namespace,
		YAML:            yamlContent,
		ResourceVersion: result.ResourceVersion,
	}, nil
}

func (s *HPAService) Delete(ctx context.Context, clusterName, namespace, name, operator string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := k8soperator.NewHPAOperator(client)

	// 获取当前状态用于保存历史（转换为 YAML 存储，便于回滚）
	hpa, err := op.Get(ctx, namespace, name)
	if err == nil {
		yamlContent, _ := utils.EncodeToYAML(hpa, "autoscaling/v2", "HorizontalPodAutoscaler")
		_ = s.saveHistory(ctx, clusterName, namespace, name, yamlContent, "delete", operator)
	}

	if err := op.Delete(ctx, namespace, name); err != nil {
		return apperrors.Wrap(err, 500, 500, "删除 HPA 失败")
	}

	return nil
}

func (s *HPAService) saveHistory(ctx context.Context, clusterName, namespace, resourceName, yamlContent, operation, operator string) error {
	latestVersion, _ := s.historyRepo.GetLatestVersionByClusterName(ctx, clusterName, namespace, "HPA", resourceName)

	history := &entity.ResourceHistory{
		ClusterName:  clusterName,
		Namespace:    namespace,
		ResourceType: "HPA",
		ResourceName: resourceName,
		Version:      latestVersion + 1,
		Content:      yamlContent,
		Operation:    operation,
		Operator:     operator,
	}

	return s.historyRepo.Create(ctx, history)
}

// buildHPAResponse 构建 HPA 详细响应
func (s *HPAService) buildHPAResponse(hpa *autoscalingv2.HorizontalPodAutoscaler, yamlContent string) dto.HPAResponse {
	resp := dto.HPAResponse{
		Name:      hpa.Name,
		Namespace: hpa.Namespace,
		ScaleTargetRef: dto.HPAScaleTargetRef{
			APIVersion: hpa.Spec.ScaleTargetRef.APIVersion,
			Kind:       hpa.Spec.ScaleTargetRef.Kind,
			Name:       hpa.Spec.ScaleTargetRef.Name,
		},
		MaxReplicas:     hpa.Spec.MaxReplicas,
		CurrentReplicas: hpa.Status.CurrentReplicas,
		DesiredReplicas: hpa.Status.DesiredReplicas,
		YAML:            yamlContent,
		ResourceVersion: hpa.ResourceVersion,
	}

	// MinReplicas 默认为 1
	if hpa.Spec.MinReplicas != nil {
		resp.MinReplicas = *hpa.Spec.MinReplicas
	} else {
		resp.MinReplicas = 1
	}

	// 构建 CPU 和内存统计
	resp.CPU, resp.Memory = s.buildResourceMetrics(hpa)

	// 构建指标状态
	resp.Metrics = s.buildMetricStatuses(hpa)

	// 构建条件状态
	for _, cond := range hpa.Status.Conditions {
		resp.Conditions = append(resp.Conditions, dto.HPACondition{
			Type:               string(cond.Type),
			Status:             string(cond.Status),
			Reason:             cond.Reason,
			Message:            cond.Message,
			LastTransitionTime: cond.LastTransitionTime.Format("2006-01-02T15:04:05Z"),
		})
	}

	return resp
}

// buildResourceMetrics 构建 CPU 和内存资源指标
func (s *HPAService) buildResourceMetrics(hpa *autoscalingv2.HorizontalPodAutoscaler) (*dto.HPAResourceMetric, *dto.HPAResourceMetric) {
	var cpuMetric, memoryMetric *dto.HPAResourceMetric

	// 解析目标指标配置
	for _, metric := range hpa.Spec.Metrics {
		if metric.Type == autoscalingv2.ResourceMetricSourceType && metric.Resource != nil {
			rm := &dto.HPAResourceMetric{
				TargetType: string(metric.Resource.Target.Type),
			}

			if metric.Resource.Target.AverageUtilization != nil {
				rm.TargetUtilization = metric.Resource.Target.AverageUtilization
			}
			if metric.Resource.Target.AverageValue != nil {
				if metric.Resource.Name == "memory" {
					rm.TargetAverageValue = formatMemoryQuantity(metric.Resource.Target.AverageValue)
				} else {
					rm.TargetAverageValue = metric.Resource.Target.AverageValue.String()
				}
			}

			switch metric.Resource.Name {
			case "cpu":
				cpuMetric = rm
			case "memory":
				memoryMetric = rm
			}
		}
	}

	// 解析当前指标值
	for _, metric := range hpa.Status.CurrentMetrics {
		if metric.Type == autoscalingv2.ResourceMetricSourceType && metric.Resource != nil {
			switch metric.Resource.Name {
			case "cpu":
				if cpuMetric != nil {
					if metric.Resource.Current.AverageUtilization != nil {
						cpuMetric.CurrentUtilization = metric.Resource.Current.AverageUtilization
					}
					if metric.Resource.Current.AverageValue != nil {
						cpuMetric.CurrentAverageValue = metric.Resource.Current.AverageValue.String()
					}
				}
			case "memory":
				if memoryMetric != nil {
					if metric.Resource.Current.AverageUtilization != nil {
						memoryMetric.CurrentUtilization = metric.Resource.Current.AverageUtilization
					}
					if metric.Resource.Current.AverageValue != nil {
						memoryMetric.CurrentAverageValue = formatMemoryQuantity(metric.Resource.Current.AverageValue)
					}
				}
			}
		}
	}

	return cpuMetric, memoryMetric
}

// buildMetricStatuses 构建指标状态列表
func (s *HPAService) buildMetricStatuses(hpa *autoscalingv2.HorizontalPodAutoscaler) []dto.HPAMetricStatus {
	var metrics []dto.HPAMetricStatus

	// 创建目标指标 map，用于关联当前值
	targetMetrics := make(map[string]*dto.HPAMetricStatus)

	// 解析目标指标配置
	for _, metric := range hpa.Spec.Metrics {
		ms := dto.HPAMetricStatus{
			Type: string(metric.Type),
		}

		switch metric.Type {
		case autoscalingv2.ResourceMetricSourceType:
			if metric.Resource != nil {
				ms.Name = string(metric.Resource.Name)
				ms.TargetType = string(metric.Resource.Target.Type)
				if metric.Resource.Target.AverageUtilization != nil {
					ms.TargetUtilization = metric.Resource.Target.AverageUtilization
				}
				if metric.Resource.Target.AverageValue != nil {
					if metric.Resource.Name == "memory" {
						ms.TargetValue = formatMemoryQuantity(metric.Resource.Target.AverageValue)
					} else {
						ms.TargetValue = metric.Resource.Target.AverageValue.String()
					}
				}
				targetMetrics[ms.Name] = &ms
			}
		case autoscalingv2.PodsMetricSourceType:
			if metric.Pods != nil {
				ms.Name = metric.Pods.Metric.Name
				ms.TargetType = string(metric.Pods.Target.Type)
				if metric.Pods.Target.AverageValue != nil {
					ms.TargetValue = metric.Pods.Target.AverageValue.String()
				}
				targetMetrics[ms.Name] = &ms
			}
		case autoscalingv2.ObjectMetricSourceType:
			if metric.Object != nil {
				ms.Name = metric.Object.Metric.Name
				ms.TargetType = string(metric.Object.Target.Type)
				if metric.Object.Target.Value != nil {
					ms.TargetValue = metric.Object.Target.Value.String()
				}
				if metric.Object.Target.AverageValue != nil {
					ms.TargetValue = metric.Object.Target.AverageValue.String()
				}
				targetMetrics[ms.Name] = &ms
			}
		case autoscalingv2.ExternalMetricSourceType:
			if metric.External != nil {
				ms.Name = metric.External.Metric.Name
				ms.TargetType = string(metric.External.Target.Type)
				if metric.External.Target.Value != nil {
					ms.TargetValue = metric.External.Target.Value.String()
				}
				if metric.External.Target.AverageValue != nil {
					ms.TargetValue = metric.External.Target.AverageValue.String()
				}
				targetMetrics[ms.Name] = &ms
			}
		}
	}

	// 解析当前指标值
	for _, metric := range hpa.Status.CurrentMetrics {
		var name string

		switch metric.Type {
		case autoscalingv2.ResourceMetricSourceType:
			if metric.Resource != nil {
				name = string(metric.Resource.Name)
				if ms, ok := targetMetrics[name]; ok {
					if metric.Resource.Current.AverageUtilization != nil {
						ms.CurrentUtilization = metric.Resource.Current.AverageUtilization
					}
					if metric.Resource.Current.AverageValue != nil {
						if metric.Resource.Name == "memory" {
							ms.CurrentValue = formatMemoryQuantity(metric.Resource.Current.AverageValue)
						} else {
							ms.CurrentValue = metric.Resource.Current.AverageValue.String()
						}
					}
				}
			}
		case autoscalingv2.PodsMetricSourceType:
			if metric.Pods != nil {
				name = metric.Pods.Metric.Name
				if ms, ok := targetMetrics[name]; ok {
					if metric.Pods.Current.AverageValue != nil {
						ms.CurrentValue = metric.Pods.Current.AverageValue.String()
					}
				}
			}
		case autoscalingv2.ObjectMetricSourceType:
			if metric.Object != nil {
				name = metric.Object.Metric.Name
				if ms, ok := targetMetrics[name]; ok {
					if metric.Object.Current.Value != nil {
						ms.CurrentValue = metric.Object.Current.Value.String()
					}
					if metric.Object.Current.AverageValue != nil {
						ms.CurrentValue = metric.Object.Current.AverageValue.String()
					}
				}
			}
		case autoscalingv2.ExternalMetricSourceType:
			if metric.External != nil {
				name = metric.External.Metric.Name
				if ms, ok := targetMetrics[name]; ok {
					if metric.External.Current.Value != nil {
						ms.CurrentValue = metric.External.Current.Value.String()
					}
					if metric.External.Current.AverageValue != nil {
						ms.CurrentValue = metric.External.Current.AverageValue.String()
					}
				}
			}
		}
	}

	// 将 map 转为列表
	for _, ms := range targetMetrics {
		metrics = append(metrics, *ms)
	}

	return metrics
}
