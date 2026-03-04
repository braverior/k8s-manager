package handler

import (
	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/service"
)

type NodeHandler struct {
	svc *service.NodeService
}

func NewNodeHandler(svc *service.NodeService) *NodeHandler {
	return &NodeHandler{svc: svc}
}

// List 获取节点列表（支持分页和搜索）
func (h *NodeHandler) List(c *gin.Context) {
	clusterName := c.Param("cluster")

	var query dto.ResourceQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		query = dto.ResourceQuery{Page: 1, PageSize: 50}
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 50
	}

	nodes, total, err := h.svc.List(c.Request.Context(), clusterName, &query)
	if err != nil {
		handleError(c, err)
		return
	}
	response.SuccessWithPage(c, total, nodes)
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
