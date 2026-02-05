package handler

import (
	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/service"
)

type NodeHandler struct {
	svc *service.NodeService
}

func NewNodeHandler(svc *service.NodeService) *NodeHandler {
	return &NodeHandler{svc: svc}
}

// List 获取节点列表
func (h *NodeHandler) List(c *gin.Context) {
	clusterName := c.Param("cluster")

	nodes, err := h.svc.List(c.Request.Context(), clusterName)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nodes)
}

// Get 获取单个节点详情
func (h *NodeHandler) Get(c *gin.Context) {
	clusterName := c.Param("cluster")
	nodeName := c.Param("name")

	node, err := h.svc.Get(c.Request.Context(), clusterName, nodeName)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, node)
}
