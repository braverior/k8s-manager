package entity

import (
	"time"
)

// ResourceHistory 资源版本历史
type ResourceHistory struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ClusterName  string    `gorm:"column:cluster_name;size:255;not null;index" json:"cluster_name"`
	Namespace    string    `gorm:"size:255;not null" json:"namespace"`
	ResourceType string    `gorm:"column:resource_type;size:50;not null" json:"resource_type"`
	ResourceName string    `gorm:"column:resource_name;size:255;not null" json:"resource_name"`
	Version      uint      `gorm:"not null" json:"version"`
	Content      string    `gorm:"type:longtext;not null" json:"content"`
	Operation    string    `gorm:"size:50;not null" json:"operation"`
	Operator     string    `gorm:"size:255" json:"operator"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ResourceHistory) TableName() string {
	return "resource_histories"
}

// User 用户信息
type User struct {
	ID             string     `gorm:"primaryKey;size:255" json:"id"`                                   // 飞书 open_id
	Name           string     `gorm:"size:255;not null" json:"name"`                                   // 用户姓名
	Email          string     `gorm:"size:255;index" json:"email"`                                     // 邮箱
	Mobile         string     `gorm:"size:50" json:"mobile"`                                           // 手机号
	AvatarURL      string     `gorm:"column:avatar_url;size:1024" json:"avatar_url"`                   // 头像 URL
	EmployeeID     string     `gorm:"column:employee_id;size:255;index" json:"employee_id"`            // 企业员工 ID
	DepartmentID   string     `gorm:"column:department_id;size:255;index" json:"department_id"`        // 部门 ID
	DepartmentName string     `gorm:"column:department_name;size:255" json:"department_name"`          // 部门名称
	DepartmentPath string     `gorm:"column:department_path;size:1024" json:"department_path"`         // 部门完整路径
	Role           string     `gorm:"size:50;not null;default:user" json:"role"`                       // 角色: admin/user
	Status         string     `gorm:"size:50;not null;default:active" json:"status"`                   // 状态: active/disabled
	LastLoginAt    *time.Time `gorm:"column:last_login_at" json:"last_login_at"`                       // 最后登录时间（可为空）
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`                                // 创建时间
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                                // 更新时间
}

func (User) TableName() string {
	return "users"
}

// IsAdmin 判断是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// IsActive 判断用户是否激活
func (u *User) IsActive() bool {
	return u.Status == "active"
}

// UserPermission 用户权限
type UserPermission struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     string    `gorm:"column:user_id;size:255;not null;index" json:"user_id"`          // 用户 ID
	Cluster    string    `gorm:"size:255;not null;index" json:"cluster"`                         // 集群名称
	Namespaces string    `gorm:"type:text;not null" json:"namespaces"`                           // 命名空间列表，JSON 数组格式，["*"] 表示所有
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserPermission) TableName() string {
	return "user_permissions"
}
