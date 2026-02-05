package handler

import (
	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/model/dto"
	apperrors "k8s_api_server/internal/pkg/errors"
	"k8s_api_server/internal/service"
)

type ConfigMapHandler struct {
	svc *service.ConfigMapService
}

func NewConfigMapHandler(svc *service.ConfigMapService) *ConfigMapHandler {
	return &ConfigMapHandler{svc: svc}
}

func (h *ConfigMapHandler) List(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")

	cms, err := h.svc.List(c.Request.Context(), clusterName, namespace)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, cms)
}

func (h *ConfigMapHandler) Get(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	name := c.Param("name")

	cm, err := h.svc.Get(c.Request.Context(), clusterName, namespace, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, cm)
}

func (h *ConfigMapHandler) Apply(c *gin.Context) {
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

	cm, err := h.svc.Apply(c.Request.Context(), clusterName, namespace, &req, operator)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, cm)
}

func (h *ConfigMapHandler) Delete(c *gin.Context) {
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

func handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		response.Error(c, appErr.HTTPCode, appErr.Code, appErr.Message)
		return
	}
	response.InternalError(c, err.Error())
}

// getOperator 从上下文获取操作者信息
func getOperator(c *gin.Context) string {
	// 优先使用用户名，其次使用用户 ID
	if userName, exists := c.Get("userName"); exists {
		return userName.(string)
	}
	if userID, exists := c.Get("userID"); exists {
		return userID.(string)
	}
	return ""
}
