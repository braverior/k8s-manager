package service

import (
	"context"
	"fmt"
	"regexp"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"k8s_api_server/internal/k8s"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/model/entity"
	"k8s_api_server/internal/pkg/crypto"
	apperrors "k8s_api_server/internal/pkg/errors"
	"k8s_api_server/internal/pkg/logger"
	"k8s_api_server/internal/repository"
)

var clusterNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_]*$`)

type ClusterManageService struct {
	clusterRepo   *repository.ClusterRepository
	clientManager *k8s.ClientManager
	encryptionKey []byte
}

func NewClusterManageService(
	clusterRepo *repository.ClusterRepository,
	clientManager *k8s.ClientManager,
	encryptionKey string,
) *ClusterManageService {
	return &ClusterManageService{
		clusterRepo:   clusterRepo,
		clientManager: clientManager,
		encryptionKey: []byte(encryptionKey),
	}
}

// AddCluster 添加集群
func (s *ClusterManageService) AddCluster(ctx context.Context, req *dto.AddClusterRequest, createdBy string) (*dto.ClusterDetailResponse, error) {
	// 校验集群名
	if !clusterNameRegex.MatchString(req.Name) {
		return nil, apperrors.New(400, 400, "cluster name must contain only letters, numbers, hyphens, and underscores")
	}

	// 检查是否已存在
	exists, err := s.clusterRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "failed to check cluster existence")
	}
	if exists {
		return nil, apperrors.ErrClusterAlreadyExists
	}

	// 校验 kubeconfig
	kubeconfigData := []byte(req.Kubeconfig)
	restCfg, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigData)
	if err != nil {
		return nil, apperrors.ErrKubeconfigInvalid
	}

	// 测试连接
	if err := s.clientManager.AddClusterFromKubeconfig(req.Name, req.Description, kubeconfigData); err != nil {
		return nil, apperrors.Wrap(err, 4002, 400, "failed to connect to cluster")
	}

	// 测试连接是否可用
	version, err := s.clientManager.TestConnection(ctx, req.Name)
	if err != nil {
		s.clientManager.RemoveCluster(req.Name)
		return nil, apperrors.Wrap(err, 4002, 400, "failed to connect to cluster")
	}

	// 加密 kubeconfig
	encrypted, err := crypto.Encrypt(kubeconfigData, s.encryptionKey)
	if err != nil {
		s.clientManager.RemoveCluster(req.Name)
		return nil, apperrors.Wrap(err, 500, 500, "failed to encrypt kubeconfig")
	}

	// 存入数据库
	cluster := &entity.Cluster{
		Name:           req.Name,
		Description:    req.Description,
		ClusterType:    "kubeconfig",
		KubeconfigData: encrypted,
		APIServer:      restCfg.Host,
		Status:         "connected",
		Source:         "database",
		CreatedBy:      createdBy,
	}

	if err := s.clusterRepo.Create(ctx, cluster); err != nil {
		s.clientManager.RemoveCluster(req.Name)
		return nil, apperrors.Wrap(err, 500, 500, "failed to save cluster")
	}

	logger.Info(fmt.Sprintf("Cluster %s added by %s, version: %s", req.Name, createdBy, version))

	return s.toDetailResponse(cluster), nil
}

// UpdateCluster 更新集群
func (s *ClusterManageService) UpdateCluster(ctx context.Context, name string, req *dto.UpdateClusterRequest) (*dto.ClusterDetailResponse, error) {
	cluster, err := s.clusterRepo.GetByName(ctx, name)
	if err != nil {
		return nil, apperrors.ErrClusterNotFound
	}

	if req.Description != nil {
		cluster.Description = *req.Description
	}

	if req.Kubeconfig != nil && *req.Kubeconfig != "" {
		kubeconfigData := []byte(*req.Kubeconfig)

		// 校验 kubeconfig
		restCfg, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigData)
		if err != nil {
			return nil, apperrors.ErrKubeconfigInvalid
		}

		// 重新加载客户端
		s.clientManager.RemoveCluster(name)
		if err := s.clientManager.AddClusterFromKubeconfig(name, cluster.Description, kubeconfigData); err != nil {
			return nil, apperrors.Wrap(err, 4002, 400, "failed to connect to cluster with new kubeconfig")
		}

		// 测试连接
		if _, err := s.clientManager.TestConnection(ctx, name); err != nil {
			return nil, apperrors.Wrap(err, 4002, 400, "failed to connect to cluster with new kubeconfig")
		}

		// 加密新的 kubeconfig
		encrypted, err := crypto.Encrypt(kubeconfigData, s.encryptionKey)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, 500, "failed to encrypt kubeconfig")
		}

		cluster.KubeconfigData = encrypted
		cluster.APIServer = restCfg.Host
		cluster.Status = "connected"
		cluster.ClusterType = "kubeconfig"
	}

	if err := s.clusterRepo.Update(ctx, cluster); err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "failed to update cluster")
	}

	return s.toDetailResponse(cluster), nil
}

// DeleteCluster 删除集群
func (s *ClusterManageService) DeleteCluster(ctx context.Context, name string) error {
	_, err := s.clusterRepo.GetByName(ctx, name)
	if err != nil {
		return apperrors.ErrClusterNotFound
	}

	// 从数据库删除
	if err := s.clusterRepo.Delete(ctx, name); err != nil {
		return apperrors.Wrap(err, 500, 500, "failed to delete cluster")
	}

	// 清理关联权限
	if err := s.clusterRepo.DeletePermissionsByCluster(ctx, name); err != nil {
		logger.Error(fmt.Sprintf("Failed to clean up permissions for cluster %s: %v", name, err))
	}

	// 从 ClientManager 移除
	s.clientManager.RemoveCluster(name)

	logger.Info(fmt.Sprintf("Cluster %s deleted", name))
	return nil
}

// ListClusters 列出所有集群
func (s *ClusterManageService) ListClusters(ctx context.Context) ([]dto.ClusterDetailResponse, error) {
	clusters, err := s.clusterRepo.List(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, 500, "failed to list clusters")
	}

	result := make([]dto.ClusterDetailResponse, 0, len(clusters))
	for _, c := range clusters {
		result = append(result, *s.toDetailResponse(&c))
	}
	return result, nil
}

// GetCluster 获取集群详情
func (s *ClusterManageService) GetCluster(ctx context.Context, name string) (*dto.ClusterDetailResponse, error) {
	cluster, err := s.clusterRepo.GetByName(ctx, name)
	if err != nil {
		return nil, apperrors.ErrClusterNotFound
	}
	return s.toDetailResponse(cluster), nil
}

// TestNewConnection 测试新连接（不保存）
func (s *ClusterManageService) TestNewConnection(ctx context.Context, req *dto.TestNewConnectionRequest) (*dto.TestConnectionResponse, error) {
	kubeconfigData := []byte(req.Kubeconfig)

	// 校验 kubeconfig
	restCfg, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigData)
	if err != nil {
		return &dto.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("invalid kubeconfig: %v", err),
		}, nil
	}

	// 创建临时客户端测试连接
	tempClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return &dto.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("failed to create client: %v", err),
		}, nil
	}

	version, err := tempClient.Discovery().ServerVersion()
	if err != nil {
		return &dto.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("failed to connect: %v", err),
		}, nil
	}

	return &dto.TestConnectionResponse{
		Success: true,
		Message: "connection successful",
		Version: version.GitVersion,
	}, nil
}

// MigrateConfigClusters 启动时将遗留的 source=config 集群转为 source=database
// 这是一个迁移操作：旧版本中集群可以通过 config.yaml 导入（source=config），
// 新版本仅支持 kubeconfig 方式管理。此方法将所有 config 来源的集群转为 database 来源，
// 使其可以通过管理页面进行完整管理（上传 kubeconfig、删除等）。
// 历史数据（resource_histories、user_permissions）通过集群名字关联，不受影响。
func (s *ClusterManageService) MigrateConfigClusters(ctx context.Context) error {
	clusters, err := s.clusterRepo.List(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to list clusters for migration: %v", err))
		return nil
	}
	for _, cluster := range clusters {
		if cluster.Source == "config" {
			cluster.Source = "database"
			if err := s.clusterRepo.Update(ctx, &cluster); err != nil {
				logger.Error(fmt.Sprintf("Failed to migrate cluster %s to database source: %v", cluster.Name, err))
			} else {
				logger.Info(fmt.Sprintf("Migrated cluster %s from config source to database source (data preserved)", cluster.Name))
			}
		}
	}
	return nil
}

// LoadAllClusters 启动时从 DB 加载所有集群到 ClientManager
func (s *ClusterManageService) LoadAllClusters(ctx context.Context) error {
	clusters, err := s.clusterRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list clusters from DB: %w", err)
	}

	for _, cluster := range clusters {
		if cluster.KubeconfigData == "" {
			continue
		}
		if s.clientManager.HasCluster(cluster.Name) {
			continue
		}

		// 解密 kubeconfig
		kubeconfigData, err := crypto.Decrypt(cluster.KubeconfigData, s.encryptionKey)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to decrypt kubeconfig for cluster %s: %v", cluster.Name, err))
			_ = s.clusterRepo.UpdateStatus(ctx, cluster.Name, "disconnected")
			continue
		}

		// 加载到 ClientManager
		if err := s.clientManager.AddClusterFromKubeconfig(cluster.Name, cluster.Description, kubeconfigData); err != nil {
			logger.Error(fmt.Sprintf("Failed to load cluster %s from DB: %v", cluster.Name, err))
			_ = s.clusterRepo.UpdateStatus(ctx, cluster.Name, "disconnected")
			continue
		}

		_ = s.clusterRepo.UpdateStatus(ctx, cluster.Name, "connected")
		logger.Info(fmt.Sprintf("Loaded cluster %s from DB", cluster.Name))
	}

	return nil
}

func (s *ClusterManageService) toDetailResponse(cluster *entity.Cluster) *dto.ClusterDetailResponse {
	return &dto.ClusterDetailResponse{
		ID:            cluster.ID,
		Name:          cluster.Name,
		Description:   cluster.Description,
		ClusterType:   cluster.ClusterType,
		APIServer:     cluster.APIServer,
		Status:        cluster.Status,
		Source:        cluster.Source,
		HasKubeconfig: cluster.KubeconfigData != "",
		CreatedBy:     cluster.CreatedBy,
		CreatedAt:     cluster.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     cluster.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
