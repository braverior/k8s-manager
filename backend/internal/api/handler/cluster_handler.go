package handler

import (
	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/service"
)

type ClusterHandler struct {
	svc *service.ClusterService
}

func NewClusterHandler(svc *service.ClusterService) *ClusterHandler {
	return &ClusterHandler{svc: svc}
}

func (h *ClusterHandler) List(c *gin.Context) {
	clusters, err := h.svc.List(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, clusters)
}

func (h *ClusterHandler) Get(c *gin.Context) {
	clusterName := c.Param("cluster")

	cluster, err := h.svc.Get(c.Request.Context(), clusterName)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, cluster)
}

func (h *ClusterHandler) TestConnection(c *gin.Context) {
	clusterName := c.Param("cluster")

	result, err := h.svc.TestConnection(c.Request.Context(), clusterName)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ClusterHandler) GetNamespaces(c *gin.Context) {
	clusterName := c.Param("cluster")

	namespaces, err := h.svc.GetNamespaces(c.Request.Context(), clusterName)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, namespaces)
}
