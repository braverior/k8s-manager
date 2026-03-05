package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	HTTPCode int    `json:"-"`
	Err      error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code int, httpCode int, message string) *AppError {
	return &AppError{
		Code:     code,
		HTTPCode: httpCode,
		Message:  message,
	}
}

func Wrap(err error, code int, httpCode int, message string) *AppError {
	return &AppError{
		Code:     code,
		HTTPCode: httpCode,
		Message:  message,
		Err:      err,
	}
}

// Common errors
var (
	ErrInvalidRequest   = New(400, http.StatusBadRequest, "invalid request")
	ErrUnauthorized     = New(401, http.StatusUnauthorized, "unauthorized")
	ErrForbidden        = New(403, http.StatusForbidden, "forbidden")
	ErrNotFound         = New(404, http.StatusNotFound, "resource not found")
	ErrConflict         = New(409, http.StatusConflict, "resource conflict")
	ErrInternalServer   = New(500, http.StatusInternalServerError, "internal server error")
	ErrClusterNotFound  = New(4001, http.StatusNotFound, "cluster not found")
	ErrClusterConnect   = New(4002, http.StatusBadRequest, "failed to connect to cluster")
	ErrResourceNotFound = New(4003, http.StatusNotFound, "kubernetes resource not found")
	ErrHistoryNotFound  = New(4004, http.StatusNotFound, "history not found")

	// 权限相关错误
	ErrClusterAccessDenied   = New(4005, http.StatusForbidden, "no permission to access this cluster")
	ErrNamespaceAccessDenied = New(4006, http.StatusForbidden, "no permission to access this namespace")

	// 飞书认证相关错误
	ErrFeishuCodeInvalid     = New(4010, http.StatusUnauthorized, "authorization code is invalid or expired")
	ErrFeishuTokenFailed     = New(4011, http.StatusUnauthorized, "failed to get feishu token")
	ErrFeishuUserInfoFailed  = New(4012, http.StatusUnauthorized, "failed to get feishu user info")
	ErrUserAccessDenied      = New(4013, http.StatusForbidden, "user is not allowed to access this system")
	ErrUserDisabled          = New(4014, http.StatusForbidden, "user is disabled")

	// 用户管理相关错误
	ErrUserNotFound          = New(4020, http.StatusNotFound, "user not found")
	ErrAdminRequired         = New(4021, http.StatusForbidden, "admin permission required")
	ErrCannotModifySelfRole  = New(4022, http.StatusBadRequest, "cannot modify your own admin role")

	// 集群管理相关错误
	ErrClusterAlreadyExists  = New(4030, http.StatusConflict, "cluster already exists")
	ErrClusterIsConfigSource = New(4031, http.StatusBadRequest, "cannot modify config-sourced cluster")
	ErrKubeconfigRequired    = New(4032, http.StatusBadRequest, "kubeconfig is required")
	ErrKubeconfigInvalid     = New(4033, http.StatusBadRequest, "kubeconfig is invalid")
)

func IsNotFound(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPCode == http.StatusNotFound
	}
	return false
}
