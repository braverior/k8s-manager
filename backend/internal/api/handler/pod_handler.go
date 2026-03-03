package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/service"
)

type PodHandler struct {
	svc *service.PodService
}

func NewPodHandler(svc *service.PodService) *PodHandler {
	return &PodHandler{svc: svc}
}

// List 列出命名空间下的所有 Pod
// 支持 ?deployment=xxx 参数按 Deployment 筛选
func (h *PodHandler) List(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Query("deployment")

	var pods []dto.PodResponse
	var err error

	if deploymentName != "" {
		pods, err = h.svc.ListByDeployment(c.Request.Context(), clusterName, namespace, deploymentName)
	} else {
		pods, err = h.svc.List(c.Request.Context(), clusterName, namespace)
	}

	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, pods)
}

// ListByDeployment 根据 Deployment 名称查询关联的 Pod
func (h *PodHandler) ListByDeployment(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	deploymentName := c.Param("name")

	pods, err := h.svc.ListByDeployment(c.Request.Context(), clusterName, namespace, deploymentName)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, pods)
}

// Get 获取单个 Pod 详情
func (h *PodHandler) Get(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")

	pod, err := h.svc.Get(c.Request.Context(), clusterName, namespace, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, pod)
}

// Delete 删除 Pod（用于重启）
func (h *PodHandler) Delete(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")

	if err := h.svc.Delete(c.Request.Context(), clusterName, namespace, name); err != nil {
		handleError(c, err)
		return
	}
	response.NoContent(c)
}

// GetLogs 获取 Pod 容器日志
func (h *PodHandler) GetLogs(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")
	container := c.Query("container")

	tailLines := int64(200)
	if tl := c.Query("tail_lines"); tl != "" {
		if v, err := strconv.ParseInt(tl, 10, 64); err == nil && v > 0 {
			tailLines = v
		}
	}

	previous := c.Query("previous") == "true"
	timestamps := c.Query("timestamps") == "true"

	logs, err := h.svc.GetLogs(c.Request.Context(), clusterName, namespace, name, container, tailLines, previous, timestamps)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, logs)
}

// GetEvents 获取 Pod 事件
func (h *PodHandler) GetEvents(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")

	events, err := h.svc.GetEvents(c.Request.Context(), clusterName, namespace, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, events)
}
