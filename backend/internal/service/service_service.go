package service

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"k8s_api_server/internal/k8s"
	k8soperator "k8s_api_server/internal/k8s/operator"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/model/entity"
	apperrors "k8s_api_server/internal/pkg/errors"
	"k8s_api_server/internal/pkg/utils"
	"k8s_api_server/internal/repository"
)

type ServiceService struct {
	clientManager *k8s.ClientManager
	historyRepo   *repository.HistoryRepository
}

func NewServiceService(clientManager *k8s.ClientManager, historyRepo *repository.HistoryRepository) *ServiceService {
	return &ServiceService{
		clientManager: clientManager,
		historyRepo:   historyRepo,
	}
}

func (s *ServiceService) List(ctx context.Context, clusterName, namespace string) ([]dto.ResourceYAMLResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := k8soperator.NewServiceOperator(client)
	services, err := op.List(ctx, namespace)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "获取 Service 列表失败")
	}

	var responses []dto.ResourceYAMLResponse
	for _, svc := range services {
		yamlContent, err := utils.EncodeToYAML(&svc, "v1", "Service")
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "转换 YAML 失败")
		}
		responses = append(responses, dto.ResourceYAMLResponse{
			Name:            svc.Name,
			Namespace:       svc.Namespace,
			YAML:            yamlContent,
			ResourceVersion: svc.ResourceVersion,
		})
	}
	return responses, nil
}

func (s *ServiceService) Get(ctx context.Context, clusterName, namespace, name string) (*dto.ResourceYAMLResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := k8soperator.NewServiceOperator(client)
	svc, err := op.Get(ctx, namespace, name)
	if err != nil {
		return nil, apperrors.Wrap(err, 404, 404, "Service 不存在")
	}

	yamlContent, err := utils.EncodeToYAML(svc, "v1", "Service")
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "转换 YAML 失败")
	}

	return &dto.ResourceYAMLResponse{
		Name:            svc.Name,
		Namespace:       svc.Namespace,
		YAML:            yamlContent,
		ResourceVersion: svc.ResourceVersion,
	}, nil
}

func (s *ServiceService) Apply(ctx context.Context, clusterName, namespace string, req *dto.ResourceRequest, operator string) (*dto.ResourceYAMLResponse, error) {
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

	svc, ok := obj.(*corev1.Service)
	if !ok {
		return nil, apperrors.Wrap(nil, 400, 400, "YAML 内容不是有效的 Service")
	}

	svc.Namespace = namespace
	op := k8soperator.NewServiceOperator(client)

	var result *corev1.Service
	var operation string
	exists, _ := op.Exists(ctx, namespace, svc.Name)
	if exists {
		existing, err := op.Get(ctx, namespace, svc.Name)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "获取现有 Service 失败")
		}
		if req.ResourceVersion != "" && req.ResourceVersion != existing.ResourceVersion {
			return nil, apperrors.New(409, 409, "资源已被其他用户修改，请刷新后重试")
		}
		svc.ResourceVersion = existing.ResourceVersion
		svc.Spec.ClusterIP = existing.Spec.ClusterIP // 保留 ClusterIP
		result, err = op.Update(ctx, namespace, svc)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "更新 Service 失败")
		}
		operation = "update"
	} else {
		result, err = op.Create(ctx, namespace, svc)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "创建 Service 失败")
		}
		operation = "create"
	}

	_ = s.saveHistory(ctx, clusterName, namespace, result.Name, string(content), operation, operator)

	yamlContent, err := utils.EncodeToYAML(result, "v1", "Service")
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

func (s *ServiceService) Delete(ctx context.Context, clusterName, namespace, name, operator string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := k8soperator.NewServiceOperator(client)

	// 获取当前状态用于保存历史（转换为 YAML 存储，便于回滚）
	svc, err := op.Get(ctx, namespace, name)
	if err == nil {
		yamlContent, _ := utils.EncodeToYAML(svc, "v1", "Service")
		_ = s.saveHistory(ctx, clusterName, namespace, name, yamlContent, "delete", operator)
	}

	if err := op.Delete(ctx, namespace, name); err != nil {
		return apperrors.Wrap(err, 500, 500, "删除 Service 失败")
	}

	return nil
}

func (s *ServiceService) saveHistory(ctx context.Context, clusterName, namespace, resourceName, yamlContent, operation, operator string) error {
	latestVersion, _ := s.historyRepo.GetLatestVersionByClusterName(ctx, clusterName, namespace, "Service", resourceName)

	history := &entity.ResourceHistory{
		ClusterName:  clusterName,
		Namespace:    namespace,
		ResourceType: "Service",
		ResourceName: resourceName,
		Version:      latestVersion + 1,
		Content:      yamlContent,
		Operation:    operation,
		Operator:     operator,
	}

	return s.historyRepo.Create(ctx, history)
}
