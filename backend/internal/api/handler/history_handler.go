package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/service"
)

type HistoryHandler struct {
	svc *service.HistoryService
}

func NewHistoryHandler(svc *service.HistoryService) *HistoryHandler {
	return &HistoryHandler{svc: svc}
}

func (h *HistoryHandler) List(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")

	var query dto.HistoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	histories, total, err := h.svc.List(c.Request.Context(), clusterName, namespace, &query)
	if err != nil {
		handleError(c, err)
		return
	}
	response.SuccessWithPage(c, total, histories)
}

func (h *HistoryHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的历史记录 ID")
		return
	}

	history, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, history)
}

func (h *HistoryHandler) Diff(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")

	var query dto.DiffQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	diff, err := h.svc.Diff(c.Request.Context(), clusterName, namespace, &query)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, diff)
}

func (h *HistoryHandler) Rollback(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的历史记录 ID")
		return
	}

	var req dto.RollbackRequest
	_ = c.ShouldBindJSON(&req) // 可选参数

	// 获取操作者信息
	if req.Operator == "" {
		req.Operator = getOperator(c)
	}

	result, err := h.svc.Rollback(c.Request.Context(), id, &req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, result)
}
