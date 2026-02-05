package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/service"
)

// Auth 认证中间件
func Auth(authSvc *service.AuthService, userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 401, 401, "missing authorization header")
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, 401, 401, "invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证 token
		userID, err := authSvc.ValidateToken(tokenString)
		if err != nil {
			response.Error(c, 401, 401, "invalid or expired token")
			c.Abort()
			return
		}

		// 获取用户信息
		user, err := userSvc.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			response.Error(c, 401, 401, "user not found")
			c.Abort()
			return
		}

		// 检查用户状态
		if !user.IsActive() {
			response.Error(c, 403, 4014, "user is disabled")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", user.ID)
		c.Set("userName", user.Name)
		c.Set("userRole", user.Role)
		c.Set("isAdmin", user.IsAdmin())

		c.Next()
	}
}

// AdminRequired 管理员权限中间件
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			response.Error(c, 403, 4021, "admin permission required")
			c.Abort()
			return
		}
		c.Next()
	}
}

// ResourcePermission 资源权限中间件
func ResourcePermission(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			response.Error(c, 401, 401, "unauthorized")
			c.Abort()
			return
		}

		// 管理员跳过权限检查
		isAdmin, _ := c.Get("isAdmin")
		if isAdmin.(bool) {
			c.Next()
			return
		}

		cluster := c.Param("cluster")
		namespace := c.Param("namespace")

		// 检查集群权限
		if cluster != "" {
			if !userSvc.CheckClusterPermission(c.Request.Context(), userID.(string), cluster) {
				response.Error(c, 403, 4005, "no permission to access this cluster")
				c.Abort()
				return
			}

			// 检查命名空间权限
			if namespace != "" {
				if !userSvc.CheckNamespacePermission(c.Request.Context(), userID.(string), cluster, namespace) {
					response.Error(c, 403, 4006, "no permission to access this namespace")
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}
