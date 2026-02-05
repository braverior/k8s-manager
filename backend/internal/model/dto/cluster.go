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
