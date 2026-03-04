package repository

import (
	"context"

	"gorm.io/gorm"

	"k8s_api_server/internal/model/entity"
)

type HistoryRepository struct {
	db *gorm.DB
}

func NewHistoryRepository(db *gorm.DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

func (r *HistoryRepository) Create(ctx context.Context, history *entity.ResourceHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *HistoryRepository) GetByID(ctx context.Context, id uint64) (*entity.ResourceHistory, error) {
	var history entity.ResourceHistory
	err := r.db.WithContext(ctx).First(&history, id).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *HistoryRepository) List(ctx context.Context, clusterName, namespace string, resourceType, resourceName string, offset, limit int) ([]entity.ResourceHistory, int64, error) {
	var histories []entity.ResourceHistory
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.ResourceHistory{}).
		Where("cluster_name = ? AND namespace = ?", clusterName, namespace)

	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	if resourceName != "" {
		query = query.Where("resource_name = ?", resourceName)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&histories).Error
	return histories, total, err
}

func (r *HistoryRepository) GetLatestVersionByClusterName(ctx context.Context, clusterName, namespace, resourceType, resourceName string) (uint, error) {
	var history entity.ResourceHistory
	err := r.db.WithContext(ctx).
		Where("cluster_name = ? AND namespace = ? AND resource_type = ? AND resource_name = ?",
			clusterName, namespace, resourceType, resourceName).
		Order("version DESC").
		First(&history).Error

	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return history.Version, nil
}

func (r *HistoryRepository) GetPreviousVersion(ctx context.Context, clusterName, namespace, resourceType, resourceName string, currentVersion uint) (*entity.ResourceHistory, error) {
	var history entity.ResourceHistory
	err := r.db.WithContext(ctx).
		Where("cluster_name = ? AND namespace = ? AND resource_type = ? AND resource_name = ? AND version < ?",
			clusterName, namespace, resourceType, resourceName, currentVersion).
		Order("version DESC").
		First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *HistoryRepository) ListByResource(ctx context.Context, clusterName, namespace, resourceType, resourceName string) ([]entity.ResourceHistory, error) {
	var histories []entity.ResourceHistory
	err := r.db.WithContext(ctx).
		Where("cluster_name = ? AND namespace = ? AND resource_type = ? AND resource_name = ?",
			clusterName, namespace, resourceType, resourceName).
		Order("version DESC").
		Find(&histories).Error
	return histories, err
}
