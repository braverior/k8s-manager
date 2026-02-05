package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"

	"k8s_api_server/internal/k8s"
	k8soperator "k8s_api_server/internal/k8s/operator"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/model/entity"
	apperrors "k8s_api_server/internal/pkg/errors"
	"k8s_api_server/internal/pkg/utils"
	"k8s_api_server/internal/repository"
)

type HistoryService struct {
	historyRepo   *repository.HistoryRepository
	clientManager *k8s.ClientManager
}

func NewHistoryService(historyRepo *repository.HistoryRepository, clientManager *k8s.ClientManager) *HistoryService {
	return &HistoryService{
		historyRepo:   historyRepo,
		clientManager: clientManager,
	}
}

func (s *HistoryService) List(ctx context.Context, clusterName, namespace string, query *dto.HistoryQuery) ([]dto.HistoryResponse, int64, error) {
	offset := (query.Page - 1) * query.PageSize
	if offset < 0 {
		offset = 0
	}

	histories, total, err := s.historyRepo.List(ctx, clusterName, namespace, query.ResourceType, query.ResourceName, offset, query.PageSize)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, 500, 500, "获取历史记录失败")
	}

	var responses []dto.HistoryResponse
	for _, h := range histories {
		responses = append(responses, dto.HistoryResponse{
			ID:           h.ID,
			ClusterName:  h.ClusterName,
			Namespace:    h.Namespace,
			ResourceType: h.ResourceType,
			ResourceName: h.ResourceName,
			Version:      h.Version,
			Operation:    h.Operation,
			Operator:     h.Operator,
			CreatedAt:    h.CreatedAt.Format(time.RFC3339),
		})
	}

	return responses, total, nil
}

func (s *HistoryService) GetByID(ctx context.Context, id uint64) (*dto.HistoryDetailResponse, error) {
	history, err := s.historyRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrHistoryNotFound
		}
		return nil, apperrors.Wrap(err, 500, 500, "获取历史记录失败")
	}

	return &dto.HistoryDetailResponse{
		HistoryResponse: dto.HistoryResponse{
			ID:           history.ID,
			ClusterName:  history.ClusterName,
			Namespace:    history.Namespace,
			ResourceType: history.ResourceType,
			ResourceName: history.ResourceName,
			Version:      history.Version,
			Operation:    history.Operation,
			Operator:     history.Operator,
			CreatedAt:    history.CreatedAt.Format(time.RFC3339),
		},
		Content: history.Content,
	}, nil
}

func (s *HistoryService) Diff(ctx context.Context, clusterName, namespace string, query *dto.DiffQuery) (*dto.DiffResponse, error) {
	sourceHistory, err := s.historyRepo.GetByID(ctx, query.SourceVersion)
	if err != nil {
		return nil, apperrors.Wrap(err, 404, 404, "源版本不存在")
	}

	targetHistory, err := s.historyRepo.GetByID(ctx, query.TargetVersion)
	if err != nil {
		return nil, apperrors.Wrap(err, 404, 404, "目标版本不存在")
	}

	diff := s.generateDiff(sourceHistory.Content, targetHistory.Content)

	return &dto.DiffResponse{
		SourceVersion: query.SourceVersion,
		TargetVersion: query.TargetVersion,
		SourceContent: sourceHistory.Content,
		TargetContent: targetHistory.Content,
		Diff:          diff,
	}, nil
}

func (s *HistoryService) Rollback(ctx context.Context, historyID uint64, req *dto.RollbackRequest) (*dto.RollbackResponse, error) {
	history, err := s.historyRepo.GetByID(ctx, historyID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrHistoryNotFound
		}
		return nil, apperrors.Wrap(err, 500, 500, "获取历史记录失败")
	}

	client, err := s.clientManager.GetClient(history.ClusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	// 解析存储的 YAML 内容
	obj, err := utils.DecodeYAML([]byte(history.Content))
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "无效的 YAML 内容")
	}

	switch history.ResourceType {
	case "ConfigMap":
		cm, ok := obj.(*corev1.ConfigMap)
		if !ok {
			return nil, apperrors.Wrap(nil, 400, 400, "内容不是有效的 ConfigMap")
		}
		cm.ResourceVersion = ""
		op := k8soperator.NewConfigMapOperator(client)
		exists, _ := op.Exists(ctx, history.Namespace, cm.Name)
		if exists {
			_, err = op.Update(ctx, history.Namespace, cm)
		} else {
			_, err = op.Create(ctx, history.Namespace, cm)
		}
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "回滚 ConfigMap 失败")
		}

	case "Deployment":
		deploy, ok := obj.(*appsv1.Deployment)
		if !ok {
			return nil, apperrors.Wrap(nil, 400, 400, "内容不是有效的 Deployment")
		}
		deploy.ResourceVersion = ""
		op := k8soperator.NewDeploymentOperator(client)
		exists, _ := op.Exists(ctx, history.Namespace, deploy.Name)
		if exists {
			_, err = op.Update(ctx, history.Namespace, deploy)
		} else {
			_, err = op.Create(ctx, history.Namespace, deploy)
		}
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "回滚 Deployment 失败")
		}

	case "Service":
		svc, ok := obj.(*corev1.Service)
		if !ok {
			return nil, apperrors.Wrap(nil, 400, 400, "内容不是有效的 Service")
		}
		svc.ResourceVersion = ""
		op := k8soperator.NewServiceOperator(client)
		exists, _ := op.Exists(ctx, history.Namespace, svc.Name)
		if exists {
			existing, _ := op.Get(ctx, history.Namespace, svc.Name)
			if existing != nil {
				svc.Spec.ClusterIP = existing.Spec.ClusterIP
			}
			_, err = op.Update(ctx, history.Namespace, svc)
		} else {
			svc.Spec.ClusterIP = ""
			_, err = op.Create(ctx, history.Namespace, svc)
		}
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "回滚 Service 失败")
		}

	case "HPA":
		hpa, ok := obj.(*autoscalingv2.HorizontalPodAutoscaler)
		if !ok {
			return nil, apperrors.Wrap(nil, 400, 400, "内容不是有效的 HPA")
		}
		hpa.ResourceVersion = ""
		op := k8soperator.NewHPAOperator(client)
		exists, _ := op.Exists(ctx, history.Namespace, hpa.Name)
		if exists {
			_, err = op.Update(ctx, history.Namespace, hpa)
		} else {
			_, err = op.Create(ctx, history.Namespace, hpa)
		}
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "回滚 HPA 失败")
		}

	default:
		return nil, apperrors.Wrap(nil, 400, 400, fmt.Sprintf("不支持的资源类型: %s", history.ResourceType))
	}

	latestVersion, _ := s.historyRepo.GetLatestVersionByClusterName(ctx, history.ClusterName, history.Namespace, history.ResourceType, history.ResourceName)
	newHistory := &entity.ResourceHistory{
		ClusterName:  history.ClusterName,
		Namespace:    history.Namespace,
		ResourceType: history.ResourceType,
		ResourceName: history.ResourceName,
		Version:      latestVersion + 1,
		Content:      history.Content,
		Operation:    fmt.Sprintf("rollback_to_v%d", history.Version),
		Operator:     req.Operator,
	}
	_ = s.historyRepo.Create(ctx, newHistory)

	return &dto.RollbackResponse{
		Success:         true,
		Message:         fmt.Sprintf("成功回滚到版本 %d", history.Version),
		RestoredVersion: history.Version,
		NewVersion:      latestVersion + 1,
	}, nil
}

func (s *HistoryService) generateDiff(source, target string) string {
	sourceLines := strings.Split(source, "\n")
	targetLines := strings.Split(target, "\n")

	var diff strings.Builder
	diff.WriteString("--- source\n")
	diff.WriteString("+++ target\n")

	maxLines := len(sourceLines)
	if len(targetLines) > maxLines {
		maxLines = len(targetLines)
	}

	for i := 0; i < maxLines; i++ {
		sourceLine := ""
		targetLine := ""
		if i < len(sourceLines) {
			sourceLine = sourceLines[i]
		}
		if i < len(targetLines) {
			targetLine = targetLines[i]
		}

		if sourceLine != targetLine {
			if sourceLine != "" {
				diff.WriteString(fmt.Sprintf("- %s\n", sourceLine))
			}
			if targetLine != "" {
				diff.WriteString(fmt.Sprintf("+ %s\n", targetLine))
			}
		}
	}

	return diff.String()
}
