package dto

type HistoryQuery struct {
	ResourceType string `form:"resource_type"`
	ResourceName string `form:"resource_name"`
	Page         int    `form:"page,default=1"`
	PageSize     int    `form:"page_size,default=20"`
}

type HistoryResponse struct {
	ID           uint64 `json:"id"`
	ClusterName  string `json:"cluster_name"`
	Namespace    string `json:"namespace"`
	ResourceType string `json:"resource_type"`
	ResourceName string `json:"resource_name"`
	Version      uint   `json:"version"`
	Operation    string `json:"operation"`
	Operator     string `json:"operator"`
	CreatedAt    string `json:"created_at"`
}

type HistoryDetailResponse struct {
	HistoryResponse
	Content string `json:"content"`
}

type DiffQuery struct {
	SourceVersion uint64 `form:"source_version" binding:"required"`
	TargetVersion uint64 `form:"target_version" binding:"required"`
}

type DiffResponse struct {
	SourceVersion uint64 `json:"source_version"`
	TargetVersion uint64 `json:"target_version"`
	SourceContent string `json:"source_content"`
	TargetContent string `json:"target_content"`
	Diff          string `json:"diff"`
}

type RollbackRequest struct {
	Operator string `json:"operator"`
}

type RollbackResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	RestoredVersion uint   `json:"restored_version"`
	NewVersion      uint   `json:"new_version"`
}
