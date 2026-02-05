package handler

import (
	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// GetFeishuConfig 获取飞书登录配置
func (h *AuthHandler) GetFeishuConfig(c *gin.Context) {
	config := h.authSvc.GetFeishuConfig()
	response.Success(c, config)
}

// FeishuLogin 飞书登录
func (h *AuthHandler) FeishuLogin(c *gin.Context) {
	var req dto.FeishuLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	result, err := h.authSvc.FeishuLogin(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, result)
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, 401, 401, "unauthorized")
		return
	}

	user, err := h.authSvc.GetCurrentUser(c.Request.Context(), userID.(string))
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, user)
}

// Logout 退出登录
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT 是无状态的，客户端删除 token 即可
	// 如果需要实现 token 黑名单，可以在这里添加逻辑
	response.Success(c, nil)
}
