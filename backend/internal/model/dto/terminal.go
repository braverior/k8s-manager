package dto

// TerminalMessage WebSocket 消息格式
type TerminalMessage struct {
	// 消息类型: input, output, resize, ping, pong, error
	Type string `json:"type"`
	// 数据内容
	Data string `json:"data,omitempty"`
	// 终端列数（resize 类型使用）
	Cols uint16 `json:"cols,omitempty"`
	// 终端行数（resize 类型使用）
	Rows uint16 `json:"rows,omitempty"`
}

// TerminalMessage 类型常量
const (
	TerminalMessageTypeInput  = "input"
	TerminalMessageTypeOutput = "output"
	TerminalMessageTypeResize = "resize"
	TerminalMessageTypePing   = "ping"
	TerminalMessageTypePong   = "pong"
	TerminalMessageTypeError  = "error"
)

// ContainerInfo 容器信息
type ContainerInfo struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	State string `json:"state"`
}

// ContainerListResponse 容器列表响应
type ContainerListResponse struct {
	Containers []ContainerInfo `json:"containers"`
}
