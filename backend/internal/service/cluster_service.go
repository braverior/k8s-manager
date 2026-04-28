package service

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s_api_server/internal/k8s"
	"k8s_api_server/internal/model/dto"
	apperrors "k8s_api_server/internal/pkg/errors"
)

type ClusterService struct {
	clientManager *k8s.ClientManager
}

func NewClusterService(clientManager *k8s.ClientManager) *ClusterService {
	return &ClusterService{
		clientManager: clientManager,
	}
}

func (s *ClusterService) List(ctx context.Context) ([]dto.ClusterResponse, error) {
	clusters := s.clientManager.ListClusters()

	responses := make([]dto.ClusterResponse, 0, len(clusters))
	for _, c := range clusters {
		responses = append(responses, dto.ClusterResponse{
			Name:        c.Name,
			Description: c.Description,
			APIServer:   c.APIServer,
			Status:      c.Status,
		})
	}
	return responses, nil
}

func (s *ClusterService) Get(ctx context.Context, clusterName string) (*dto.ClusterResponse, error) {
	info, err := s.clientManager.GetClusterInfo(clusterName)
	if err != nil {
		return nil, apperrors.ErrClusterNotFound
	}

	return &dto.ClusterResponse{
		Name:        info.Name,
		Description: info.Description,
		APIServer:   info.APIServer,
		Status:      info.Status,
	}, nil
}

func (s *ClusterService) TestConnection(ctx context.Context, clusterName string) (*dto.TestConnectionResponse, error) {
	version, err := s.clientManager.TestConnection(ctx, clusterName)
	if err != nil {
		return &dto.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("连接失败: %v", err),
		}, nil
	}

	return &dto.TestConnectionResponse{
		Success: true,
		Message: "连接成功",
		Version: version,
	}, nil
}

func (s *ClusterService) GetNamespaces(ctx context.Context, clusterName string) ([]dto.NamespaceResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	list, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "获取命名空间列表失败")
	}

	var namespaces []dto.NamespaceResponse
	for _, ns := range list.Items {
		namespaces = append(namespaces, dto.NamespaceResponse{
			Name:   ns.Name,
			Status: string(ns.Status.Phase),
		})
	}
	return namespaces, nil
}
