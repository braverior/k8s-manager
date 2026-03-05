package handler

import (
	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/service"
)

type ClusterManageHandler struct {
	svc *service.ClusterManageService
}

func NewClusterManageHandler(svc *service.ClusterManageService) *ClusterManageHandler {
	return &ClusterManageHandler{svc: svc}
}

// ListClusters 列出所有集群
func (h *ClusterManageHandler) ListClusters(c *gin.Context) {
	clusters, err := h.svc.ListClusters(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, clusters)
}

// GetCluster 获取集群详情
func (h *ClusterManageHandler) GetCluster(c *gin.Context) {
	name := c.Param("name")
	cluster, err := h.svc.GetCluster(c.Request.Context(), name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, cluster)
}

// AddCluster 添加集群
func (h *ClusterManageHandler) AddCluster(c *gin.Context) {
	var req dto.AddClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: name and kubeconfig are required")
		return
	}

	createdBy := ""
	if userName, exists := c.Get("userName"); exists {
		createdBy = userName.(string)
	} else if userID, exists := c.Get("userID"); exists {
		createdBy = userID.(string)
	}

	cluster, err := h.svc.AddCluster(c.Request.Context(), &req, createdBy)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, cluster)
}

// UpdateCluster 更新集群
func (h *ClusterManageHandler) UpdateCluster(c *gin.Context) {
	name := c.Param("name")

	var req dto.UpdateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	cluster, err := h.svc.UpdateCluster(c.Request.Context(), name, &req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, cluster)
}

// DeleteCluster 删除集群
func (h *ClusterManageHandler) DeleteCluster(c *gin.Context) {
	name := c.Param("name")

	if err := h.svc.DeleteCluster(c.Request.Context(), name); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil)
}

// TestNewConnection 测试新连接
func (h *ClusterManageHandler) TestNewConnection(c *gin.Context) {
	var req dto.TestNewConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: kubeconfig is required")
		return
	}

	result, err := h.svc.TestNewConnection(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, result)
}
