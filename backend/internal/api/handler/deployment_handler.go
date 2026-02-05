package handler

import (
	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/service"
)

type DeploymentHandler struct {
	svc *service.DeploymentService
}

func NewDeploymentHandler(svc *service.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{svc: svc}
}

func (h *DeploymentHandler) List(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")

	deployments, err := h.svc.List(c.Request.Context(), clusterName, namespace)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, deployments)
}

func (h *DeploymentHandler) Get(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")

	deployment, err := h.svc.Get(c.Request.Context(), clusterName, namespace, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, deployment)
}

func (h *DeploymentHandler) Apply(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")

	var req dto.ResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.YAML == "" && req.Content == "" {
		response.BadRequest(c, "yaml 或 content 字段必须提供其一")
		return
	}

	// 获取操作者信息
	operator := getOperator(c)

	deployment, err := h.svc.Apply(c.Request.Context(), clusterName, namespace, &req, operator)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, deployment)
}

func (h *DeploymentHandler) Delete(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")

	// 获取操作者信息
	operator := getOperator(c)

	if err := h.svc.Delete(c.Request.Context(), clusterName, namespace, name, operator); err != nil {
		handleError(c, err)
		return
	}
	response.NoContent(c)
}
