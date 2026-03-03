package service

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/client-go/tools/remotecommand"

	"k8s_api_server/internal/k8s"
	"k8s_api_server/internal/k8s/operator"
	"k8s_api_server/internal/model/dto"
	apperrors "k8s_api_server/internal/pkg/errors"
)

// TerminalSession 终端会话
type TerminalSession struct {
	ID            string
	UserID        string
	ClusterName   string
	Namespace     string
	PodName       string
	ContainerName string
	SizeChan      chan remotecommand.TerminalSize
	Done          chan struct{}
}

// Next 实现 remotecommand.TerminalSizeQueue 接口
func (s *TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-s.SizeChan:
		return &size
	case <-s.Done:
		return nil
	}
}

// Close 关闭会话
func (s *TerminalSession) Close() {
	select {
	case <-s.Done:
		// 已关闭
	default:
		close(s.Done)
	}
}

// TerminalService 终端服务
type TerminalService struct {
	clientManager *k8s.ClientManager
	sessions      map[string]*TerminalSession
	mu            sync.RWMutex
}

// NewTerminalService 创建终端服务
func NewTerminalService(clientManager *k8s.ClientManager) *TerminalService {
	return &TerminalService{
		clientManager: clientManager,
		sessions:      make(map[string]*TerminalSession),
	}
}

// GetContainers 获取 Pod 的容器列表
func (s *TerminalService) GetContainers(ctx context.Context, clusterName, namespace, podName string) (*dto.ContainerListResponse, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群客户端失败")
	}

	config, err := s.clientManager.GetConfig(clusterName)
	if err != nil {
		return nil, apperrors.Wrap(err, 400, 400, "获取集群配置失败")
	}

	execOp := operator.NewExecOperator(client, config)
	containers, statuses, err := execOp.GetContainers(ctx, namespace, podName)
	if err != nil {
		return nil, apperrors.Wrap(err, 404, 404, "Pod 不存在")
	}

	// 构建状态映射
	statusMap := make(map[string]string)
	for _, cs := range statuses {
		if cs.State.Running != nil {
			statusMap[cs.Name] = "Running"
		} else if cs.State.Waiting != nil {
			statusMap[cs.Name] = "Waiting"
		} else if cs.State.Terminated != nil {
			statusMap[cs.Name] = "Terminated"
		}
	}

	result := &dto.ContainerListResponse{
		Containers: make([]dto.ContainerInfo, 0, len(containers)),
	}

	for _, c := range containers {
		state := statusMap[c.Name]
		if state == "" {
			state = "Unknown"
		}
		result.Containers = append(result.Containers, dto.ContainerInfo{
			Name:  c.Name,
			Image: c.Image,
			State: state,
		})
	}

	return result, nil
}

// CreateSession 创建终端会话
func (s *TerminalService) CreateSession(sessionID, userID, clusterName, namespace, podName, containerName string) *TerminalSession {
	session := &TerminalSession{
		ID:            sessionID,
		UserID:        userID,
		ClusterName:   clusterName,
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,
		SizeChan:      make(chan remotecommand.TerminalSize, 1),
		Done:          make(chan struct{}),
	}

	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()

	return session
}

// GetSession 获取终端会话
func (s *TerminalService) GetSession(sessionID string) (*TerminalSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[sessionID]
	return session, ok
}

// RemoveSession 移除终端会话
func (s *TerminalService) RemoveSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if session, ok := s.sessions[sessionID]; ok {
		session.Close()
		delete(s.sessions, sessionID)
	}
}

// GetExecOperator 获取 exec 操作器
func (s *TerminalService) GetExecOperator(clusterName string) (*operator.ExecOperator, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取集群客户端失败: %w", err)
	}

	config, err := s.clientManager.GetConfig(clusterName)
	if err != nil {
		return nil, fmt.Errorf("获取集群配置失败: %w", err)
	}

	return operator.NewExecOperator(client, config), nil
}

// SessionCount 获取当前会话数
func (s *TerminalService) SessionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}
