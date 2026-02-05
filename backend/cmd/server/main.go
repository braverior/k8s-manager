package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s_api_server/internal/api/handler"
	"k8s_api_server/internal/api/router"
	"k8s_api_server/internal/config"
	"k8s_api_server/internal/k8s"
	"k8s_api_server/internal/pkg/logger"
	"k8s_api_server/internal/repository"
	"k8s_api_server/internal/service"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(&cfg.Log); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("正在启动 K8s API Server...")

	// 初始化数据库
	db, err := repository.NewDB(&cfg.Database)
	if err != nil {
		logger.Fatal(fmt.Sprintf("连接数据库失败: %v", err))
	}
	logger.Info("数据库已连接")

	// 初始化 Repository
	historyRepo := repository.NewHistoryRepository(db)
	userRepo := repository.NewUserRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)

	// 初始化 K8s 客户端管理器
	clientManager := k8s.NewClientManager()

	// 从配置文件加载集群
	if err := clientManager.LoadFromConfig(cfg.Clusters); err != nil {
		logger.Error(fmt.Sprintf("加载集群失败: %v", err))
	}
	logger.Info(fmt.Sprintf("已加载 %d 个集群", len(cfg.Clusters)))

	// 初始化 Services
	clusterService := service.NewClusterService(clientManager)
	configMapService := service.NewConfigMapService(clientManager, historyRepo)
	deploymentService := service.NewDeploymentService(clientManager, historyRepo)
	serviceService := service.NewServiceService(clientManager, historyRepo)
	hpaService := service.NewHPAService(clientManager, historyRepo)
	historyService := service.NewHistoryService(historyRepo, clientManager)
	podService := service.NewPodService(clientManager)
	dashboardService := service.NewDashboardService(clientManager)
	nodeService := service.NewNodeService(clientManager)
	authService := service.NewAuthService(cfg, userRepo, permissionRepo)
	userService := service.NewUserService(userRepo, permissionRepo)

	// 初始化 Handlers
	clusterHandler := handler.NewClusterHandler(clusterService)
	configMapHandler := handler.NewConfigMapHandler(configMapService)
	deploymentHandler := handler.NewDeploymentHandler(deploymentService)
	serviceHandler := handler.NewServiceHandler(serviceService)
	hpaHandler := handler.NewHPAHandler(hpaService)
	historyHandler := handler.NewHistoryHandler(historyService)
	podHandler := handler.NewPodHandler(podService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	nodeHandler := handler.NewNodeHandler(nodeService)
	authHandler := handler.NewAuthHandler(authService)
	adminHandler := handler.NewAdminHandler(userService)

	// 设置路由
	r := router.NewRouter(
		clusterHandler,
		configMapHandler,
		deploymentHandler,
		serviceHandler,
		hpaHandler,
		historyHandler,
		podHandler,
		dashboardHandler,
		nodeHandler,
		authHandler,
		adminHandler,
		authService,
		userService,
	)
	engine := r.Setup(cfg.Server.Mode)

	// 创建服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: engine,
	}

	// 在 goroutine 中启动服务器
	go func() {
		logger.Info(fmt.Sprintf("服务器监听端口 %d", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(fmt.Sprintf("启动服务器失败: %v", err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal(fmt.Sprintf("服务器强制关闭: %v", err))
	}

	logger.Info("服务器已退出")
}
