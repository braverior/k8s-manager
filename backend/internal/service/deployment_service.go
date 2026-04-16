package service

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s_api_server/internal/k8s"
	k8soperator "k8s_api_server/internal/k8s/operator"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/model/entity"
	apperrors "k8s_api_server/internal/pkg/errors"
	"k8s_api_server/internal/pkg/utils"
	"k8s_api_server/internal/repository"
)

type DeploymentService struct {
	clientManager *k8s.ClientManager
	historyRepo   *repository.HistoryRepository
}

func NewDeploymentService(clientManager *k8s.ClientManager, historyRepo *repository.HistoryRepository) *DeploymentService {
	return &DeploymentService{
		clientManager: clientManager,
		historyRepo:   historyRepo,
	}
}

func (s *DeploymentService) List(ctx context.Context, clusterName, namespace string) ([]dto.DeploymentResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := k8soperator.NewDeploymentOperator(client)
	deployments, err := op.List(ctx, namespace)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "获取 Deployment 列表失败")
	}

	// 获取所有 ReplicaSet，建立 RS -> Deployment 的归属映射
	rsList, _ := client.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	rsToDeployment := make(map[string]string)
	if rsList != nil {
		for _, rs := range rsList.Items {
			for _, owner := range rs.OwnerReferences {
				if owner.Kind == "Deployment" {
					rsToDeployment[rs.Name] = owner.Name
					break
				}
			}
		}
	}

	// 获取所有 Pod，通过 RS ownerReferences 精确统计每个 Deployment 的 Pod 数量及状态
	pods, _ := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	podCountByDeployment := make(map[string]int)
	podStatusByDeployment := make(map[string]map[string]int)
	if pods != nil {
		for _, pod := range pods.Items {
			for _, owner := range pod.OwnerReferences {
				if owner.Kind == "ReplicaSet" {
					if deployName, ok := rsToDeployment[owner.Name]; ok {
						podCountByDeployment[deployName]++
						if podStatusByDeployment[deployName] == nil {
							podStatusByDeployment[deployName] = make(map[string]int)
						}
						podStatusByDeployment[deployName][string(pod.Status.Phase)]++
					}
					break
				}
			}
		}
	}

	var responses []dto.DeploymentResponse
	for _, d := range deployments {
		yamlContent, err := utils.EncodeToYAML(&d, "apps/v1", "Deployment")
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "转换 YAML 失败")
		}
		replicas := int32(0)
		if d.Spec.Replicas != nil {
			replicas = *d.Spec.Replicas
		}
		responses = append(responses, dto.DeploymentResponse{
			Name:              d.Name,
			Namespace:         d.Namespace,
			Replicas:          replicas,
			ReadyReplicas:     d.Status.ReadyReplicas,
			UpdatedReplicas:   d.Status.UpdatedReplicas,
			AvailableReplicas: d.Status.AvailableReplicas,
			PodCount:          podCountByDeployment[d.Name],
			PodStatusCounts:   podStatusByDeployment[d.Name],
			YAML:              yamlContent,
		})
	}
	return responses, nil
}

func (s *DeploymentService) Get(ctx context.Context, clusterName, namespace, name string) (*dto.DeploymentResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := k8soperator.NewDeploymentOperator(client)
	d, err := op.Get(ctx, namespace, name)
	if err != nil {
		return nil, apperrors.Wrap(err, 404, 404, "Deployment 不存在")
	}

	// 通过 ReplicaSet ownerReferences 精确统计 Pod 数量
	rsList, _ := client.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	rsNames := make(map[string]bool)
	if rsList != nil {
		for _, rs := range rsList.Items {
			for _, owner := range rs.OwnerReferences {
				if owner.Kind == "Deployment" && owner.Name == name {
					rsNames[rs.Name] = true
					break
				}
			}
		}
	}
	podCount := 0
	if pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{}); err == nil {
		for _, pod := range pods.Items {
			for _, owner := range pod.OwnerReferences {
				if owner.Kind == "ReplicaSet" && rsNames[owner.Name] {
					podCount++
					break
				}
			}
		}
	}

	yamlContent, err := utils.EncodeToYAML(d, "apps/v1", "Deployment")
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "转换 YAML 失败")
	}

	replicas := int32(0)
	if d.Spec.Replicas != nil {
		replicas = *d.Spec.Replicas
	}

	return &dto.DeploymentResponse{
		Name:              d.Name,
		Namespace:         d.Namespace,
		Replicas:          replicas,
		ReadyReplicas:     d.Status.ReadyReplicas,
		UpdatedReplicas:   d.Status.UpdatedReplicas,
		AvailableReplicas: d.Status.AvailableReplicas,
		PodCount:          podCount,
		YAML:              yamlContent,
	}, nil
}

func (s *DeploymentService) Apply(ctx context.Context, clusterName, namespace string, req *dto.ResourceRequest, operator string) (*dto.ResourceYAMLResponse, error) {
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

	deploy, ok := obj.(*appsv1.Deployment)
	if !ok {
		return nil, apperrors.Wrap(nil, 400, 400, "YAML 内容不是有效的 Deployment")
	}

	deploy.Namespace = namespace
	op := k8soperator.NewDeploymentOperator(client)

	var result *appsv1.Deployment
	var operation string
	exists, _ := op.Exists(ctx, namespace, deploy.Name)
	if exists {
		existing, err := op.Get(ctx, namespace, deploy.Name)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "获取现有 Deployment 失败")
		}
		deploy.ResourceVersion = existing.ResourceVersion
		result, err = op.Update(ctx, namespace, deploy)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "更新 Deployment 失败")
		}
		operation = "update"
	} else {
		result, err = op.Create(ctx, namespace, deploy)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "创建 Deployment 失败")
		}
		operation = "create"
	}

	_ = s.saveHistory(ctx, clusterName, namespace, result.Name, string(content), operation, operator)

	yamlContent, err := utils.EncodeToYAML(result, "apps/v1", "Deployment")
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "转换 YAML 失败")
	}

	return &dto.ResourceYAMLResponse{
		Name:      result.Name,
		Namespace: result.Namespace,
		YAML:      yamlContent,
	}, nil
}

func (s *DeploymentService) Delete(ctx context.Context, clusterName, namespace, name, operator string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := k8soperator.NewDeploymentOperator(client)

	// 获取当前状态用于保存历史（转换为 YAML 存储，便于回滚）
	deploy, err := op.Get(ctx, namespace, name)
	if err == nil {
		yamlContent, _ := utils.EncodeToYAML(deploy, "apps/v1", "Deployment")
		_ = s.saveHistory(ctx, clusterName, namespace, name, yamlContent, "delete", operator)
	}

	if err := op.Delete(ctx, namespace, name); err != nil {
		return apperrors.Wrap(err, 500, 500, "删除 Deployment 失败")
	}

	return nil
}

func (s *DeploymentService) Restart(ctx context.Context, clusterName, namespace, name, operator string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	op := k8soperator.NewDeploymentOperator(client)
	deploy, err := op.Get(ctx, namespace, name)
	if err != nil {
		return apperrors.Wrap(err, 404, 404, "Deployment 不存在")
	}

	// 等价于 kubectl rollout restart：在 pod template 上添加/更新 restart 注解
	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = make(map[string]string)
	}
	deploy.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	result, err := op.Update(ctx, namespace, deploy)
	if err != nil {
		return apperrors.Wrap(err, 500, 500, "重启 Deployment 失败")
	}

	yamlContent, _ := utils.EncodeToYAML(result, "apps/v1", "Deployment")
	_ = s.saveHistory(ctx, clusterName, namespace, name, yamlContent, "restart", operator)
	return nil
}

func (s *DeploymentService) saveHistory(ctx context.Context, clusterName, namespace, resourceName, yamlContent, operation, operator string) error {
	latestVersion, _ := s.historyRepo.GetLatestVersionByClusterName(ctx, clusterName, namespace, "Deployment", resourceName)

	history := &entity.ResourceHistory{
		ClusterName:  clusterName,
		Namespace:    namespace,
		ResourceType: "Deployment",
		ResourceName: resourceName,
		Version:      latestVersion + 1,
		Content:      yamlContent,
		Operation:    operation,
		Operator:     operator,
	}

	return s.historyRepo.Create(ctx, history)
}
