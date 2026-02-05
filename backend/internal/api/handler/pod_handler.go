package handler

import (
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
