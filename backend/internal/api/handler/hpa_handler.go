package handler

import (
	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/service"
)

type HPAHandler struct {
	svc *service.HPAService
}

func NewHPAHandler(svc *service.HPAService) *HPAHandler {
	return &HPAHandler{svc: svc}
}

func (h *HPAHandler) List(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")

	hpas, err := h.svc.List(c.Request.Context(), clusterName, namespace)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, hpas)
}

func (h *HPAHandler) Get(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")

	hpa, err := h.svc.Get(c.Request.Context(), clusterName, namespace, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, hpa)
}

func (h *HPAHandler) Apply(c *gin.Context) {
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

	hpa, err := h.svc.Apply(c.Request.Context(), clusterName, namespace, &req, operator)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, hpa)
}

func (h *HPAHandler) Delete(c *gin.Context) {
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
