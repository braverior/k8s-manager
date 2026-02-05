package repository

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/model/entity"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List 获取用户列表
func (r *UserRepository) List(ctx context.Context, req *dto.UserListRequest) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.User{})

	// 关键词搜索
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("name LIKE ? OR email LIKE ? OR employee_id LIKE ?", keyword, keyword, keyword)
	}

	// 部门筛选
	if req.DepartmentID != "" {
		query = query.Where("department_id = ?", req.DepartmentID)
	}

	// 角色筛选
	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateRole 更新用户角色
func (r *UserRepository) UpdateRole(ctx context.Context, id string, role string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("role", role).Error
}

// UpdateStatus 更新用户状态
func (r *UserRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("status", status).Error
}

// PermissionRepository 用户权限仓库
type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// GetByUserID 获取用户的所有权限
func (r *PermissionRepository) GetByUserID(ctx context.Context, userID string) ([]entity.UserPermission, error) {
	var permissions []entity.UserPermission
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&permissions).Error
	return permissions, err
}

// SetPermissions 设置用户权限（覆盖）
func (r *PermissionRepository) SetPermissions(ctx context.Context, userID string, permissions []dto.ClusterPermission) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除旧权限
		if err := tx.Where("user_id = ?", userID).Delete(&entity.UserPermission{}).Error; err != nil {
			return err
		}

		// 添加新权限
		for _, p := range permissions {
			namespacesJSON, _ := json.Marshal(p.Namespaces)
			perm := entity.UserPermission{
				UserID:     userID,
				Cluster:    p.Cluster,
				Namespaces: string(namespacesJSON),
			}
			if err := tx.Create(&perm).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// AddClusterPermission 添加集群权限
func (r *PermissionRepository) AddClusterPermission(ctx context.Context, userID string, cluster string, namespaces []string) error {
	// 先检查是否已存在
	var existing entity.UserPermission
	err := r.db.WithContext(ctx).Where("user_id = ? AND cluster = ?", userID, cluster).First(&existing).Error
	if err == nil {
		// 已存在，更新
		namespacesJSON, _ := json.Marshal(namespaces)
		return r.db.WithContext(ctx).Model(&existing).Update("namespaces", string(namespacesJSON)).Error
	}

	// 不存在，创建
	namespacesJSON, _ := json.Marshal(namespaces)
	perm := entity.UserPermission{
		UserID:     userID,
		Cluster:    cluster,
		Namespaces: string(namespacesJSON),
	}
	return r.db.WithContext(ctx).Create(&perm).Error
}

// RemoveClusterPermission 移除集群权限
func (r *PermissionRepository) RemoveClusterPermission(ctx context.Context, userID string, cluster string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND cluster = ?", userID, cluster).Delete(&entity.UserPermission{}).Error
}

// UpdateNamespaces 更新命名空间权限
func (r *PermissionRepository) UpdateNamespaces(ctx context.Context, userID string, cluster string, namespaces []string) error {
	namespacesJSON, _ := json.Marshal(namespaces)
	result := r.db.WithContext(ctx).Model(&entity.UserPermission{}).
		Where("user_id = ? AND cluster = ?", userID, cluster).
		Update("namespaces", string(namespacesJSON))
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// GetUserClusterPermission 获取用户对指定集群的权限
func (r *PermissionRepository) GetUserClusterPermission(ctx context.Context, userID string, cluster string) (*entity.UserPermission, error) {
	var perm entity.UserPermission
	err := r.db.WithContext(ctx).Where("user_id = ? AND cluster = ?", userID, cluster).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}
