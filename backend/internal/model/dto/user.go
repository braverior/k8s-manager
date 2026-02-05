package dto

// UserListRequest 用户列表请求
type UserListRequest struct {
	Keyword      string `form:"keyword"`
	DepartmentID string `form:"department_id"`
	Role         string `form:"role"`
	Page         int    `form:"page,default=1"`
	PageSize     int    `form:"page_size,default=20"`
}

// UserRoleRequest 更新用户角色请求
type UserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin user"`
}

// UserStatusRequest 更新用户状态请求
type UserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active disabled"`
}

// UserPermissionsRequest 设置用户权限请求
type UserPermissionsRequest struct {
	Permissions []ClusterPermission `json:"permissions" binding:"required"`
}

// AddClusterPermissionRequest 添加集群权限请求
type AddClusterPermissionRequest struct {
	Cluster    string   `json:"cluster" binding:"required"`
	Namespaces []string `json:"namespaces" binding:"required"`
}

// UpdateNamespacesRequest 更新命名空间权限请求
type UpdateNamespacesRequest struct {
	Namespaces []string `json:"namespaces" binding:"required"`
}

// BatchPermissionsRequest 批量设置权限请求
type BatchPermissionsRequest struct {
	UserIDs     []string            `json:"user_ids" binding:"required,min=1"`
	Permissions []ClusterPermission `json:"permissions" binding:"required"`
}

// BatchPermissionsResponse 批量设置权限响应
type BatchPermissionsResponse struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	FailedUsers  []string `json:"failed_users"`
}

// UserPermissionsResponse 用户权限响应
type UserPermissionsResponse struct {
	UserID      string              `json:"user_id"`
	UserName    string              `json:"user_name"`
	IsAdmin     bool                `json:"is_admin"`
	Permissions []ClusterPermission `json:"permissions"`
}
