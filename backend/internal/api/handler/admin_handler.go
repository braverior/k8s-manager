package handler

import (
	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/service"
)

type AdminHandler struct {
	userSvc *service.UserService
}

func NewAdminHandler(userSvc *service.UserService) *AdminHandler {
	return &AdminHandler{userSvc: userSvc}
}

// ListUsers 获取用户列表
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	users, total, err := h.userSvc.List(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.SuccessWithPage(c, total, users)
}

// GetUser 获取用户详情
func (h *AdminHandler) GetUser(c *gin.Context) {
	userID := c.Param("user_id")

	user, err := h.userSvc.GetByID(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, user)
}

// UpdateUserRole 更新用户角色
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID := c.Param("user_id")

	var req dto.UserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters: role must be 'admin' or 'user'")
		return
	}

	// 获取当前操作者 ID
	currentUserID, _ := c.Get("userID")

	if err := h.userSvc.UpdateRole(c.Request.Context(), userID, currentUserID.(string), req.Role); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateUserStatus 更新用户状态
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("user_id")

	var req dto.UserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters: status must be 'active' or 'disabled'")
		return
	}

	if err := h.userSvc.UpdateStatus(c.Request.Context(), userID, req.Status); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetUserPermissions 获取用户权限
func (h *AdminHandler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("user_id")

	permissions, err := h.userSvc.GetPermissions(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, permissions)
}

// SetUserPermissions 设置用户权限
func (h *AdminHandler) SetUserPermissions(c *gin.Context) {
	userID := c.Param("user_id")

	var req dto.UserPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	if err := h.userSvc.SetPermissions(c.Request.Context(), userID, req.Permissions); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil)
}

// AddClusterPermission 添加集群权限
func (h *AdminHandler) AddClusterPermission(c *gin.Context) {
	userID := c.Param("user_id")

	var req dto.AddClusterPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	if err := h.userSvc.AddClusterPermission(c.Request.Context(), userID, req.Cluster, req.Namespaces); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil)
}

// RemoveClusterPermission 移除集群权限
func (h *AdminHandler) RemoveClusterPermission(c *gin.Context) {
	userID := c.Param("user_id")
	cluster := c.Param("cluster")

	if err := h.userSvc.RemoveClusterPermission(c.Request.Context(), userID, cluster); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil)
}

// UpdateNamespaces 更新命名空间权限
func (h *AdminHandler) UpdateNamespaces(c *gin.Context) {
	userID := c.Param("user_id")
	cluster := c.Param("cluster")

	var req dto.UpdateNamespacesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	if err := h.userSvc.UpdateNamespaces(c.Request.Context(), userID, cluster, req.Namespaces); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, nil)
}

// BatchSetPermissions 批量设置权限
func (h *AdminHandler) BatchSetPermissions(c *gin.Context) {
	var req dto.BatchPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	result := h.userSvc.BatchSetPermissions(c.Request.Context(), &req)
	response.Success(c, result)
}
