package dto

type ClusterResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	APIServer   string `json:"api_server"`
	Status      string `json:"status"`
}

type NamespaceResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Version string `json:"version,omitempty"`
}

// 集群管理相关 DTO

type AddClusterRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Kubeconfig  string `json:"kubeconfig" binding:"required"`
}

type UpdateClusterRequest struct {
	Description *string `json:"description"`
	Kubeconfig  *string `json:"kubeconfig"`
}

type TestNewConnectionRequest struct {
	Kubeconfig string `json:"kubeconfig" binding:"required"`
}

type ClusterDetailResponse struct {
	ID             uint64 `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ClusterType    string `json:"cluster_type"`
	APIServer      string `json:"api_server"`
	Status         string `json:"status"`
	Source         string `json:"source"`
	HasKubeconfig  bool   `json:"has_kubeconfig"`
	CreatedBy      string `json:"created_by"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}
