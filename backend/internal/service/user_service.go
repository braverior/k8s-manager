package service

import (
	"context"
	"encoding/json"

	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/model/entity"
	"k8s_api_server/internal/pkg/errors"
	"k8s_api_server/internal/repository"
)

type UserService struct {
	userRepo       *repository.UserRepository
	permissionRepo *repository.PermissionRepository
}

func NewUserService(userRepo *repository.UserRepository, permissionRepo *repository.PermissionRepository) *UserService {
	return &UserService{
		userRepo:       userRepo,
		permissionRepo: permissionRepo,
	}
}

// List 获取用户列表
func (s *UserService) List(ctx context.Context, req *dto.UserListRequest) ([]dto.UserResponse, int64, error) {
	users, total, err := s.userRepo.List(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, *s.buildUserResponse(&user, nil))
	}

	return result, total, nil
}

// GetByID 获取用户详情
func (s *UserService) GetByID(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	permissions, _ := s.getUserPermissions(ctx, userID)
	resp := s.buildUserResponse(user, permissions)
	resp.Status = user.Status
	if user.LastLoginAt != nil {
		resp.LastLoginAt = user.LastLoginAt
	}
	resp.CreatedAt = &user.CreatedAt

	return resp, nil
}

// UpdateRole 更新用户角色
func (s *UserService) UpdateRole(ctx context.Context, userID string, currentUserID string, role string) error {
	// 不能修改自己的角色
	if userID == currentUserID {
		return errors.ErrCannotModifySelfRole
	}

	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	return s.userRepo.UpdateRole(ctx, userID, role)
}

// UpdateStatus 更新用户状态
func (s *UserService) UpdateStatus(ctx context.Context, userID string, status string) error {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	return s.userRepo.UpdateStatus(ctx, userID, status)
}

// GetPermissions 获取用户权限
func (s *UserService) GetPermissions(ctx context.Context, userID string) (*dto.UserPermissionsResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	permissions, _ := s.getUserPermissions(ctx, userID)

	return &dto.UserPermissionsResponse{
		UserID:      user.ID,
		UserName:    user.Name,
		IsAdmin:     user.IsAdmin(),
		Permissions: permissions,
	}, nil
}

// SetPermissions 设置用户权限（覆盖）
func (s *UserService) SetPermissions(ctx context.Context, userID string, permissions []dto.ClusterPermission) error {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	return s.permissionRepo.SetPermissions(ctx, userID, permissions)
}

// AddClusterPermission 添加集群权限
func (s *UserService) AddClusterPermission(ctx context.Context, userID string, cluster string, namespaces []string) error {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	return s.permissionRepo.AddClusterPermission(ctx, userID, cluster, namespaces)
}

// RemoveClusterPermission 移除集群权限
func (s *UserService) RemoveClusterPermission(ctx context.Context, userID string, cluster string) error {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	return s.permissionRepo.RemoveClusterPermission(ctx, userID, cluster)
}

// UpdateNamespaces 更新命名空间权限
func (s *UserService) UpdateNamespaces(ctx context.Context, userID string, cluster string, namespaces []string) error {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	return s.permissionRepo.UpdateNamespaces(ctx, userID, cluster, namespaces)
}

// BatchSetPermissions 批量设置权限
func (s *UserService) BatchSetPermissions(ctx context.Context, req *dto.BatchPermissionsRequest) *dto.BatchPermissionsResponse {
	resp := &dto.BatchPermissionsResponse{
		FailedUsers: make([]string, 0),
	}

	for _, userID := range req.UserIDs {
		if err := s.SetPermissions(ctx, userID, req.Permissions); err != nil {
			resp.FailedCount++
			resp.FailedUsers = append(resp.FailedUsers, userID)
		} else {
			resp.SuccessCount++
		}
	}

	return resp
}

// CheckClusterPermission 检查用户是否有集群权限
func (s *UserService) CheckClusterPermission(ctx context.Context, userID string, cluster string) bool {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false
	}

	// 管理员拥有所有权限
	if user.IsAdmin() {
		return true
	}

	perm, err := s.permissionRepo.GetUserClusterPermission(ctx, userID, cluster)
	return err == nil && perm != nil
}

// CheckNamespacePermission 检查用户是否有命名空间权限
func (s *UserService) CheckNamespacePermission(ctx context.Context, userID string, cluster string, namespace string) bool {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false
	}

	// 管理员拥有所有权限
	if user.IsAdmin() {
		return true
	}

	perm, err := s.permissionRepo.GetUserClusterPermission(ctx, userID, cluster)
	if err != nil {
		return false
	}

	var namespaces []string
	if err := json.Unmarshal([]byte(perm.Namespaces), &namespaces); err != nil {
		return false
	}

	// 检查是否有所有命名空间的权限
	for _, ns := range namespaces {
		if ns == "*" || ns == namespace {
			return true
		}
	}

	return false
}

// getUserPermissions 获取用户权限
func (s *UserService) getUserPermissions(ctx context.Context, userID string) ([]dto.ClusterPermission, error) {
	permissions, err := s.permissionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ClusterPermission, 0, len(permissions))
	for _, p := range permissions {
		var namespaces []string
		_ = json.Unmarshal([]byte(p.Namespaces), &namespaces)
		result = append(result, dto.ClusterPermission{
			Cluster:    p.Cluster,
			Namespaces: namespaces,
		})
	}
	return result, nil
}

// buildUserResponse 构建用户响应
func (s *UserService) buildUserResponse(user *entity.User, permissions []dto.ClusterPermission) *dto.UserResponse {
	resp := &dto.UserResponse{
		ID:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		AvatarURL:  user.AvatarURL,
		Mobile:     user.Mobile,
		EmployeeID: user.EmployeeID,
		Department: &dto.DepartmentInfo{
			ID:   user.DepartmentID,
			Name: user.DepartmentName,
			Path: user.DepartmentPath,
		},
		Role:        user.Role,
		IsAdmin:     user.IsAdmin(),
		Permissions: permissions,
		Status:      user.Status,
	}

	if user.LastLoginAt != nil {
		resp.LastLoginAt = user.LastLoginAt
	}

	createdAt := user.CreatedAt
	resp.CreatedAt = &createdAt

	return resp
}

// GetUserByID 获取用户实体（供中间件使用）
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*entity.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
