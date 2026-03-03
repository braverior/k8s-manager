package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/remotecommand"

	"k8s_api_server/internal/api/response"
	"k8s_api_server/internal/k8s/operator"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/pkg/logger"
	"k8s_api_server/internal/service"
)

const (
	// WebSocket 配置
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应该限制
	},
}

// TerminalHandler 终端处理器
type TerminalHandler struct {
	svc     *service.TerminalService
	authSvc *service.AuthService
	userSvc *service.UserService
}

// NewTerminalHandler 创建终端处理器
func NewTerminalHandler(svc *service.TerminalService, authSvc *service.AuthService, userSvc *service.UserService) *TerminalHandler {
	return &TerminalHandler{
		svc:     svc,
		authSvc: authSvc,
		userSvc: userSvc,
	}
}

// GetContainers 获取 Pod 的容器列表
func (h *TerminalHandler) GetContainers(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("name")

	result, err := h.svc.GetContainers(c.Request.Context(), clusterName, namespace, podName)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, result)
}

// Exec WebSocket exec 处理
func (h *TerminalHandler) Exec(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("name")
	containerName := c.Query("container")
	command := c.Query("command")
	token := c.Query("token")

	// 验证 token
	userID, err := h.authSvc.ValidateToken(token)
	if err != nil {
		response.Error(c, 401, 401, "invalid or expired token")
		return
	}

	// 获取用户信息
	user, err := h.userSvc.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, 401, 401, "user not found")
		return
	}

	// 检查用户状态
	if !user.IsActive() {
		response.Error(c, 403, 4014, "user is disabled")
		return
	}

	// 检查权限（管理员跳过）
	if !user.IsAdmin() {
		if !h.userSvc.CheckClusterPermission(c.Request.Context(), userID, clusterName) {
			response.Error(c, 403, 4005, "no permission to access this cluster")
			return
		}
		if !h.userSvc.CheckNamespacePermission(c.Request.Context(), userID, clusterName, namespace) {
			response.Error(c, 403, 4006, "no permission to access this namespace")
			return
		}
	}

	// 默认命令
	if command == "" {
		command = "/bin/bash"
	}

	// 升级为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("WebSocket upgrade failed: %v", err))
		return
	}
	defer conn.Close()

	// 记录会话开始
	sessionID := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())
	logger.Info(fmt.Sprintf("Terminal session started: user=%s, cluster=%s, namespace=%s, pod=%s, container=%s, session=%s",
		userID, clusterName, namespace, podName, containerName, sessionID))

	// 创建会话
	session := h.svc.CreateSession(sessionID, userID, clusterName, namespace, podName, containerName)
	defer func() {
		h.svc.RemoveSession(sessionID)
		logger.Info(fmt.Sprintf("Terminal session ended: session=%s", sessionID))
	}()

	// 获取 exec 操作器
	execOp, err := h.svc.GetExecOperator(clusterName)
	if err != nil {
		h.sendError(conn, err.Error())
		return
	}

	// 创建管道
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	var wg sync.WaitGroup

	// 启动 exec
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer stdinWriter.Close()
		defer stdoutWriter.Close()

		execOpts := &operator.ExecOptions{
			Namespace:     namespace,
			PodName:       podName,
			ContainerName: containerName,
			Command:       []string{command},
			Stdin:         stdinReader,
			Stdout:        stdoutWriter,
			Stderr:        stdoutWriter,
			TTY:           true,
			SizeQueue:     session,
		}

		if err := execOp.Exec(ctx, execOpts); err != nil {
			logger.Error(fmt.Sprintf("Exec error: %v", err))
		}
	}()

	// 读取 stdout 并发送到 WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := stdoutReader.Read(buf)
			if err != nil {
				if err != io.EOF {
					logger.Error(fmt.Sprintf("Read stdout error: %v", err))
				}
				return
			}
			if n > 0 {
				msg := dto.TerminalMessage{
					Type: dto.TerminalMessageTypeOutput,
					Data: string(buf[:n]),
				}
				if err := h.sendMessage(conn, msg); err != nil {
					return
				}
			}
		}
	}()

	// 心跳
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-ctx.Done():
				return
			case <-session.Done:
				return
			}
		}
	}()

	// 设置 WebSocket 参数
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// 读取 WebSocket 消息
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error(fmt.Sprintf("WebSocket read error: %v", err))
			}
			break
		}

		var msg dto.TerminalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case dto.TerminalMessageTypeInput:
			if _, err := stdinWriter.Write([]byte(msg.Data)); err != nil {
				logger.Error(fmt.Sprintf("Write stdin error: %v", err))
			}
		case dto.TerminalMessageTypeResize:
			select {
			case session.SizeChan <- remotecommand.TerminalSize{
				Width:  msg.Cols,
				Height: msg.Rows,
			}:
			default:
			}
		case dto.TerminalMessageTypePing:
			h.sendMessage(conn, dto.TerminalMessage{Type: dto.TerminalMessageTypePong})
		}
	}

	cancel()
	stdinWriter.Close()
	wg.Wait()
}

// sendMessage 发送 WebSocket 消息
func (h *TerminalHandler) sendMessage(conn *websocket.Conn, msg dto.TerminalMessage) error {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

// sendError 发送错误消息
func (h *TerminalHandler) sendError(conn *websocket.Conn, errMsg string) {
	h.sendMessage(conn, dto.TerminalMessage{
		Type: dto.TerminalMessageTypeError,
		Data: errMsg,
	})
}
