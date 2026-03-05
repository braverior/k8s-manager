package repository

import (
	"context"

	"gorm.io/gorm"

	"k8s_api_server/internal/model/entity"
)

type ClusterRepository struct {
	db *gorm.DB
}

func NewClusterRepository(db *gorm.DB) *ClusterRepository {
	return &ClusterRepository{db: db}
}

// Create 创建集群
func (r *ClusterRepository) Create(ctx context.Context, cluster *entity.Cluster) error {
	return r.db.WithContext(ctx).Create(cluster).Error
}

// Update 更新集群
func (r *ClusterRepository) Update(ctx context.Context, cluster *entity.Cluster) error {
	return r.db.WithContext(ctx).Save(cluster).Error
}

// Delete 删除集群
func (r *ClusterRepository) Delete(ctx context.Context, name string) error {
	return r.db.WithContext(ctx).Where("name = ?", name).Delete(&entity.Cluster{}).Error
}

// GetByName 根据名称获取集群
func (r *ClusterRepository) GetByName(ctx context.Context, name string) (*entity.Cluster, error) {
	var cluster entity.Cluster
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

// List 获取所有集群
func (r *ClusterRepository) List(ctx context.Context) ([]entity.Cluster, error) {
	var clusters []entity.Cluster
	err := r.db.WithContext(ctx).Order("created_at ASC").Find(&clusters).Error
	return clusters, err
}

// UpdateStatus 更新集群状态
func (r *ClusterRepository) UpdateStatus(ctx context.Context, name string, status string) error {
	return r.db.WithContext(ctx).Model(&entity.Cluster{}).Where("name = ?", name).Update("status", status).Error
}

// ExistsByName 检查集群是否存在
func (r *ClusterRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Cluster{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// DeletePermissionsByCluster 删除集群关联的用户权限
func (r *ClusterRepository) DeletePermissionsByCluster(ctx context.Context, clusterName string) error {
	return r.db.WithContext(ctx).Where("cluster = ?", clusterName).Delete(&entity.UserPermission{}).Error
}
