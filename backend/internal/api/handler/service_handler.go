package handler

import (
	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/service"
)

type ServiceHandler struct {
	svc *service.ServiceService
}

func NewServiceHandler(svc *service.ServiceService) *ServiceHandler {
	return &ServiceHandler{svc: svc}
}

func (h *ServiceHandler) List(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")

	services, err := h.svc.List(c.Request.Context(), clusterName, namespace)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, services)
}

func (h *ServiceHandler) Get(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")

	svc, err := h.svc.Get(c.Request.Context(), clusterName, namespace, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, svc)
}

func (h *ServiceHandler) Apply(c *gin.Context) {
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

	svc, err := h.svc.Apply(c.Request.Context(), clusterName, namespace, &req, operator)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, svc)
}

func (h *ServiceHandler) Delete(c *gin.Context) {
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
