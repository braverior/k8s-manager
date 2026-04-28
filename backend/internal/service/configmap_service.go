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

type ConfigMapService struct {
	clientManager *k8s.ClientManager
	historyRepo   *repository.HistoryRepository
}

func NewConfigMapService(clientManager *k8s.ClientManager, historyRepo *repository.HistoryRepository) *ConfigMapService {
	return &ConfigMapService{
		clientManager: clientManager,
		historyRepo:   historyRepo,
	}
}

func (s *ConfigMapService) List(ctx context.Context, clusterName, namespace string) ([]dto.ResourceYAMLResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := k8soperator.NewConfigMapOperator(client)
	cms, err := op.List(ctx, namespace)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "获取 ConfigMap 列表失败")
	}

	var responses []dto.ResourceYAMLResponse
	for _, cm := range cms {
		yamlContent, err := utils.EncodeToYAML(&cm, "v1", "ConfigMap")
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "转换 YAML 失败")
		}
		responses = append(responses, dto.ResourceYAMLResponse{
			Name:            cm.Name,
			Namespace:       cm.Namespace,
			YAML:            yamlContent,
			ResourceVersion: cm.ResourceVersion,
		})
	}
	return responses, nil
}

func (s *ConfigMapService) Get(ctx context.Context, clusterName, namespace, name string) (*dto.ResourceYAMLResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := k8soperator.NewConfigMapOperator(client)
	cm, err := op.Get(ctx, namespace, name)
	if err != nil {
		return nil, apperrors.Wrap(err, 404, 404, "ConfigMap 不存在")
	}

	yamlContent, err := utils.EncodeToYAML(cm, "v1", "ConfigMap")
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "转换 YAML 失败")
	}

	return &dto.ResourceYAMLResponse{
		Name:            cm.Name,
		Namespace:       cm.Namespace,
		YAML:            yamlContent,
		ResourceVersion: cm.ResourceVersion,
	}, nil
}

func (s *ConfigMapService) Apply(ctx context.Context, clusterName, namespace string, req *dto.ResourceRequest, operator string) (*dto.ResourceYAMLResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	// 解析 YAML 或 Base64 内容
	content, err := utils.ParseYAMLOrBase64(req.YAML, req.Content)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "解析请求内容失败")
	}

	obj, err := utils.DecodeYAML(content)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "解析 YAML 失败")
	}

	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil, apperrors.Wrap(nil, 400, 400, "YAML 内容不是有效的 ConfigMap")
	}

	// 设置 namespace
	cm.Namespace = namespace

	op := k8soperator.NewConfigMapOperator(client)

	// 检查是否存在，决定创建还是更新
	var result *corev1.ConfigMap
	var operation string
	exists, _ := op.Exists(ctx, namespace, cm.Name)
	if exists {
		// 获取现有资源以保留 resourceVersion
		existing, err := op.Get(ctx, namespace, cm.Name)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "获取现有 ConfigMap 失败")
		}
		if req.ResourceVersion != "" && req.ResourceVersion != existing.ResourceVersion {
			return nil, apperrors.New(409, 409, "资源已被其他用户修改，请刷新后重试")
		}
		cm.ResourceVersion = existing.ResourceVersion
		result, err = op.Update(ctx, namespace, cm)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "更新 ConfigMap 失败")
		}
		operation = "update"
	} else {
		result, err = op.Create(ctx, namespace, cm)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "创建 ConfigMap 失败")
		}
		operation = "create"
	}

	// 保存历史版本（存储原始 YAML）
	_ = s.saveHistory(ctx, clusterName, namespace, result.Name, string(content), operation, operator)

	yamlContent, err := utils.EncodeToYAML(result, "v1", "ConfigMap")
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

func (s *ConfigMapService) Delete(ctx context.Context, clusterName, namespace, name, operator string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := k8soperator.NewConfigMapOperator(client)

	// 获取当前状态用于保存历史（转换为 YAML 存储，便于回滚）
	cm, err := op.Get(ctx, namespace, name)
	if err == nil {
		yamlContent, _ := utils.EncodeToYAML(cm, "v1", "ConfigMap")
		_ = s.saveHistory(ctx, clusterName, namespace, name, yamlContent, "delete", operator)
	}

	if err := op.Delete(ctx, namespace, name); err != nil {
		return apperrors.Wrap(err, 500, 500, "删除 ConfigMap 失败")
	}

	return nil
}

func (s *ConfigMapService) saveHistory(ctx context.Context, clusterName, namespace, resourceName, yamlContent, operation, operator string) error {
	latestVersion, _ := s.historyRepo.GetLatestVersionByClusterName(ctx, clusterName, namespace, "ConfigMap", resourceName)

	history := &entity.ResourceHistory{
		ClusterName:  clusterName,
		Namespace:    namespace,
		ResourceType: "ConfigMap",
		ResourceName: resourceName,
		Version:      latestVersion + 1,
		Content:      yamlContent,
		Operation:    operation,
		Operator:     operator,
	}

	return s.historyRepo.Create(ctx, history)
}
