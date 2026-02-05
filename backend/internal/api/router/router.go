package router

import (
	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/handler"
	"k8s_api_server/internal/api/middleware"
	"k8s_api_server/internal/service"
)

type Router struct {
	clusterHandler    *handler.ClusterHandler
	configMapHandler  *handler.ConfigMapHandler
	deploymentHandler *handler.DeploymentHandler
	serviceHandler    *handler.ServiceHandler
	hpaHandler        *handler.HPAHandler
	historyHandler    *handler.HistoryHandler
	podHandler        *handler.PodHandler
	dashboardHandler  *handler.DashboardHandler
	nodeHandler       *handler.NodeHandler
	authHandler       *handler.AuthHandler
	adminHandler      *handler.AdminHandler
	authSvc           *service.AuthService
	userSvc           *service.UserService
}

func NewRouter(
	clusterHandler *handler.ClusterHandler,
	configMapHandler *handler.ConfigMapHandler,
	deploymentHandler *handler.DeploymentHandler,
	serviceHandler *handler.ServiceHandler,
	hpaHandler *handler.HPAHandler,
	historyHandler *handler.HistoryHandler,
	podHandler *handler.PodHandler,
	dashboardHandler *handler.DashboardHandler,
	nodeHandler *handler.NodeHandler,
	authHandler *handler.AuthHandler,
	adminHandler *handler.AdminHandler,
	authSvc *service.AuthService,
	userSvc *service.UserService,
) *Router {
	return &Router{
		clusterHandler:    clusterHandler,
		configMapHandler:  configMapHandler,
		deploymentHandler: deploymentHandler,
		serviceHandler:    serviceHandler,
		hpaHandler:        hpaHandler,
		historyHandler:    historyHandler,
		podHandler:        podHandler,
		dashboardHandler:  dashboardHandler,
		nodeHandler:       nodeHandler,
		authHandler:       authHandler,
		adminHandler:      adminHandler,
		authSvc:           authSvc,
		userSvc:           userSvc,
	}
}

func (r *Router) Setup(mode string) *gin.Engine {
	gin.SetMode(mode)
	engine := gin.New()

	// Middleware
	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.CORS())

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// 认证相关（无需登录）
		auth := v1.Group("/auth")
		{
			auth.GET("/feishu/config", r.authHandler.GetFeishuConfig)
			auth.POST("/feishu/login", r.authHandler.FeishuLogin)
		}

		// 需要认证的认证接口
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.Auth(r.authSvc, r.userSvc))
		{
			authProtected.GET("/me", r.authHandler.GetCurrentUser)
			authProtected.POST("/logout", r.authHandler.Logout)
		}

		// 管理员接口
		admin := v1.Group("/admin")
		admin.Use(middleware.Auth(r.authSvc, r.userSvc))
		admin.Use(middleware.AdminRequired())
		{
			// 用户管理
			users := admin.Group("/users")
			{
				users.GET("", r.adminHandler.ListUsers)
				users.GET("/:user_id", r.adminHandler.GetUser)
				users.PUT("/:user_id/role", r.adminHandler.UpdateUserRole)
				users.PUT("/:user_id/status", r.adminHandler.UpdateUserStatus)

				// 权限管理
				users.GET("/:user_id/permissions", r.adminHandler.GetUserPermissions)
				users.PUT("/:user_id/permissions", r.adminHandler.SetUserPermissions)
				users.POST("/:user_id/permissions/clusters", r.adminHandler.AddClusterPermission)
				users.DELETE("/:user_id/permissions/clusters/:cluster", r.adminHandler.RemoveClusterPermission)
				users.PUT("/:user_id/permissions/clusters/:cluster/namespaces", r.adminHandler.UpdateNamespaces)
			}

			// 批量权限管理
			admin.POST("/permissions/batch", r.adminHandler.BatchSetPermissions)
		}

		// 需要认证和资源权限的接口
		protected := v1.Group("")
		protected.Use(middleware.Auth(r.authSvc, r.userSvc))
		{
			// Clusters (只读，从配置文件加载)
			clusters := protected.Group("/clusters")
			{
				clusters.GET("", r.clusterHandler.List)
				clusters.GET("/:cluster", middleware.ResourcePermission(r.userSvc), r.clusterHandler.Get)
				clusters.POST("/:cluster/test-connection", middleware.ResourcePermission(r.userSvc), r.clusterHandler.TestConnection)
				clusters.GET("/:cluster/namespaces", middleware.ResourcePermission(r.userSvc), r.clusterHandler.GetNamespaces)
				// Dashboard - 集群概览
				clusters.GET("/:cluster/dashboard", middleware.ResourcePermission(r.userSvc), r.dashboardHandler.GetOverview)
				// Nodes - 节点管理
				clusters.GET("/:cluster/nodes", middleware.ResourcePermission(r.userSvc), r.nodeHandler.List)
				clusters.GET("/:cluster/nodes/:name", middleware.ResourcePermission(r.userSvc), r.nodeHandler.Get)
			}

			// Resources within clusters
			clusterResources := protected.Group("/clusters/:cluster/namespaces/:namespace")
			clusterResources.Use(middleware.ResourcePermission(r.userSvc))
			{
				// ConfigMaps
				configmaps := clusterResources.Group("/configmaps")
				{
					configmaps.GET("", r.configMapHandler.List)
					configmaps.POST("", r.configMapHandler.Apply)          // Apply (创建或更新)
					configmaps.GET("/:name", r.configMapHandler.Get)
					configmaps.PUT("/:name", r.configMapHandler.Apply)     // Apply (更新)
					configmaps.DELETE("/:name", r.configMapHandler.Delete)
				}

				// Deployments
				deployments := clusterResources.Group("/deployments")
				{
					deployments.GET("", r.deploymentHandler.List)
					deployments.POST("", r.deploymentHandler.Apply)
					deployments.GET("/:name", r.deploymentHandler.Get)
					deployments.PUT("/:name", r.deploymentHandler.Apply)
					deployments.DELETE("/:name", r.deploymentHandler.Delete)
					// 获取 Deployment 关联的 Pod
					deployments.GET("/:name/pods", r.podHandler.ListByDeployment)
				}

				// Pods
				pods := clusterResources.Group("/pods")
				{
					pods.GET("", r.podHandler.List)
					pods.GET("/:name", r.podHandler.Get)
					pods.DELETE("/:name", r.podHandler.Delete) // 删除 Pod（用于重启）
				}

				// Services
				services := clusterResources.Group("/services")
				{
					services.GET("", r.serviceHandler.List)
					services.POST("", r.serviceHandler.Apply)
					services.GET("/:name", r.serviceHandler.Get)
					services.PUT("/:name", r.serviceHandler.Apply)
					services.DELETE("/:name", r.serviceHandler.Delete)
				}

				// HPAs
				hpas := clusterResources.Group("/hpas")
				{
					hpas.GET("", r.hpaHandler.List)
					hpas.POST("", r.hpaHandler.Apply)
					hpas.GET("/:name", r.hpaHandler.Get)
					hpas.PUT("/:name", r.hpaHandler.Apply)
					hpas.DELETE("/:name", r.hpaHandler.Delete)
				}

				// Histories
				histories := clusterResources.Group("/histories")
				{
					histories.GET("", r.historyHandler.List)
					histories.GET("/diff", r.historyHandler.Diff)
					histories.GET("/:id", r.historyHandler.Get)
					histories.POST("/:id/rollback", r.historyHandler.Rollback)
				}
			}
		}
	}

	return engine
}
