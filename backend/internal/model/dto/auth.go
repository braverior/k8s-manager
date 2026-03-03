package dto

import "time"

// FeishuConfigResponse 飞书登录配置响应
type FeishuConfigResponse struct {
	AppID        string `json:"app_id"`
	RedirectURI  string `json:"redirect_uri"`
	AuthorizeURL string `json:"authorize_url"`
}

// FeishuLoginRequest 飞书登录请求
type FeishuLoginRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string        `json:"token"`
	ExpiresIn int           `json:"expires_in"`
	User      *UserResponse `json:"user"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Email        string              `json:"email"`
	AvatarURL    string              `json:"avatar_url"`
	Mobile       string              `json:"mobile"`
	EmployeeID   string              `json:"employee_id"`
	Department   *DepartmentInfo     `json:"department"`
	Role         string              `json:"role"`
	IsAdmin      bool                `json:"is_admin"`
	Permissions  []ClusterPermission `json:"permissions"`
	Status       string              `json:"status,omitempty"`
	LastLoginAt  *time.Time          `json:"last_login_at,omitempty"`
	CreatedAt    *time.Time          `json:"created_at,omitempty"`
}

// DepartmentInfo 部门信息
type DepartmentInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// ClusterPermission 集群权限
type ClusterPermission struct {
	Cluster    string   `json:"cluster"`
	Namespaces []string `json:"namespaces"`
}

// FeishuUserTokenResponse 飞书用户 Token 响应（OAuth 扁平格式）
type FeishuUserTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	Scope            string `json:"scope"`
	// 错误时返回
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// FeishuUserInfoResponse 飞书用户信息响应（OAuth 扁平格式）
type FeishuUserInfoResponse struct {
	Sub        string `json:"sub"`         // 用户唯一标识（open_id）
	Name       string `json:"name"`        // 用户姓名
	Picture    string `json:"picture"`     // 头像 URL
	OpenID     string `json:"open_id"`     // open_id
	UnionID    string `json:"union_id"`    // union_id
	EnName     string `json:"en_name"`     // 英文名
	TenantKey  string `json:"tenant_key"`  // 企业标识
	AvatarURL  string `json:"avatar_url"`  // 头像（兼容）
	Email      string `json:"email"`       // 邮箱（需要权限）
	Mobile     string `json:"mobile"`      // 手机号（需要权限）
	EmployeeNo string `json:"employee_no"` // 员工工号（需要权限）
}

// FeishuDepartmentResponse 飞书部门信息响应
type FeishuDepartmentResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		User struct {
			DepartmentIDs []string `json:"department_ids"`
		} `json:"user"`
	} `json:"data"`
}

// FeishuDepartmentDetailResponse 飞书部门详情响应
type FeishuDepartmentDetailResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Department struct {
			Name               string `json:"name"`
			ParentDepartmentID string `json:"parent_department_id"`
			OpenDepartmentID   string `json:"open_department_id"`
		} `json:"department"`
	} `json:"data"`
}

// FeishuAppTokenResponse 飞书应用 Token 响应
type FeishuAppTokenResponse struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int    `json:"expire"`
}
