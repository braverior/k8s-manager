package handler

import (
	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/service"
)

type DashboardHandler struct {
	svc *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// GetOverview 获取集群概览信息
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Query("namespace")

	overview, err := h.svc.GetOverview(c.Request.Context(), clusterName, namespace)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, overview)
}
