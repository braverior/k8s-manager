package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"k8s_api_server/internal/config"
	"k8s_api_server/internal/model/dto"
	"k8s_api_server/internal/model/entity"
	"k8s_api_server/internal/pkg/errors"
	"k8s_api_server/internal/pkg/logger"
	"k8s_api_server/internal/repository"
)

type AuthService struct {
	cfg            *config.Config
	userRepo       *repository.UserRepository
	permissionRepo *repository.PermissionRepository
}

func NewAuthService(cfg *config.Config, userRepo *repository.UserRepository, permissionRepo *repository.PermissionRepository) *AuthService {
	return &AuthService{
		cfg:            cfg,
		userRepo:       userRepo,
		permissionRepo: permissionRepo,
	}
}

// GetFeishuConfig 获取飞书登录配置
func (s *AuthService) GetFeishuConfig() *dto.FeishuConfigResponse {
	return &dto.FeishuConfigResponse{
		AppID:        s.cfg.Feishu.AppID,
		RedirectURI:  s.cfg.Feishu.RedirectURI,
		AuthorizeURL: s.cfg.Feishu.AuthorizeURL,
	}
}

// FeishuLogin 飞书登录
func (s *AuthService) FeishuLogin(ctx context.Context, req *dto.FeishuLoginRequest) (*dto.LoginResponse, error) {
	// 1. 用 code 换取 user_access_token
	userToken, err := s.getFeishuUserToken(req.Code)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrFeishuTokenFailed.Code, errors.ErrFeishuTokenFailed.HTTPCode, "failed to get feishu token")
	}

	// 2. 获取用户信息
	feishuUser, err := s.getFeishuUserInfo(userToken.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrFeishuUserInfoFailed.Code, errors.ErrFeishuUserInfoFailed.HTTPCode, "failed to get feishu user info")
	}

	// 3. 获取部门信息
	department := s.getFeishuDepartment(userToken.AccessToken)

	// 4. 查找或创建用户
	user, err := s.findOrCreateUser(ctx, feishuUser, department)
	if err != nil {
		return nil, err
	}

	// 5. 检查用户状态
	if !user.IsActive() {
		return nil, errors.ErrUserDisabled
	}

	// 6. 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	_ = s.userRepo.Update(ctx, user)

	// 7. 获取用户权限
	permissions, err := s.getUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// 8. 生成 JWT Token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalServer.Code, errors.ErrInternalServer.HTTPCode, "failed to generate token")
	}

	return &dto.LoginResponse{
		Token:     token,
		ExpiresIn: s.cfg.JWT.ExpireTime,
		User:      s.buildUserResponse(user, permissions),
	}, nil
}

// GetCurrentUser 获取当前用户信息
func (s *AuthService) GetCurrentUser(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	permissions, _ := s.getUserPermissions(ctx, userID)
	return s.buildUserResponse(user, permissions), nil
}

// ValidateToken 验证 Token 并返回用户 ID
func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["sub"].(string); ok {
			return userID, nil
		}
	}

	return "", fmt.Errorf("invalid token")
}

// getFeishuUserToken 用 code 换取飞书 user_access_token
func (s *AuthService) getFeishuUserToken(code string) (*dto.FeishuUserTokenResponse, error) {
	url := "https://passport.feishu.cn/suite/passport/oauth/token"

	body := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     s.cfg.Feishu.AppID,
		"client_secret": s.cfg.Feishu.AppSecret,
		"code":          code,
		"redirect_uri":  s.cfg.Feishu.RedirectURI,
	}
	jsonBody, _ := json.Marshal(body)

	logger.Info(fmt.Sprintf("Feishu token request: url=%s, app_id=%s, redirect_uri=%s", url, s.cfg.Feishu.AppID, s.cfg.Feishu.RedirectURI))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Error(fmt.Sprintf("Feishu token request failed: %v", err))
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	logger.Info(fmt.Sprintf("Feishu token response: %s", string(respBody)))

	var result dto.FeishuUserTokenResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		logger.Error(fmt.Sprintf("Feishu token response unmarshal failed: %v", err))
		return nil, err
	}

	if result.Error != "" {
		logger.Error(fmt.Sprintf("Feishu token error: %s - %s", result.Error, result.ErrorDescription))
		return nil, fmt.Errorf("feishu api error: %s", result.ErrorDescription)
	}

	return &result, nil
}

// getFeishuUserInfo 获取飞书用户信息
func (s *AuthService) getFeishuUserInfo(accessToken string) (*dto.FeishuUserInfoResponse, error) {
	url := "https://passport.feishu.cn/suite/passport/oauth/userinfo"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("Feishu userinfo request failed: %v", err))
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	logger.Info(fmt.Sprintf("Feishu userinfo response: %s", string(respBody)))

	var result dto.FeishuUserInfoResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		logger.Error(fmt.Sprintf("Feishu userinfo unmarshal failed: %v", err))
		return nil, err
	}

	// 检查是否获取到用户标识
	if result.OpenID == "" && result.Sub == "" {
		return nil, fmt.Errorf("failed to get user info from feishu")
	}

	return &result, nil
}

// getFeishuDepartment 获取飞书部门信息（简化版本，返回空信息如果获取失败）
func (s *AuthService) getFeishuDepartment(accessToken string) *dto.DepartmentInfo {
	// 部门信息获取比较复杂，需要 tenant_access_token，这里简化处理
	// 实际生产环境可以通过飞书开放平台 API 获取完整的部门信息
	return &dto.DepartmentInfo{
		ID:   "",
		Name: "",
		Path: "",
	}
}

// findOrCreateUser 查找或创建用户
func (s *AuthService) findOrCreateUser(ctx context.Context, feishuUser *dto.FeishuUserInfoResponse, department *dto.DepartmentInfo) (*entity.User, error) {
	// 获取用户唯一标识（优先使用 open_id，否则使用 sub）
	userID := feishuUser.OpenID
	if userID == "" {
		userID = feishuUser.Sub
	}

	// 获取头像 URL（优先使用 picture，否则使用 avatar_url）
	avatarURL := feishuUser.Picture
	if avatarURL == "" {
		avatarURL = feishuUser.AvatarURL
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err == nil {
		// 更新用户信息
		user.Name = feishuUser.Name
		user.Email = feishuUser.Email
		user.Mobile = feishuUser.Mobile
		user.AvatarURL = avatarURL
		user.EmployeeID = feishuUser.EmployeeNo
		if department != nil {
			user.DepartmentID = department.ID
			user.DepartmentName = department.Name
			user.DepartmentPath = department.Path
		}
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, err
		}
		return user, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 创建新用户
	user = &entity.User{
		ID:         userID,
		Name:       feishuUser.Name,
		Email:      feishuUser.Email,
		Mobile:     feishuUser.Mobile,
		AvatarURL:  avatarURL,
		EmployeeID: feishuUser.EmployeeNo,
		Role:       "user",
		Status:     "active",
	}
	if department != nil {
		user.DepartmentID = department.ID
		user.DepartmentName = department.Name
		user.DepartmentPath = department.Path
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// getUserPermissions 获取用户权限
func (s *AuthService) getUserPermissions(ctx context.Context, userID string) ([]dto.ClusterPermission, error) {
	permissions, err := s.permissionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ClusterPermission, 0, len(permissions))
	for _, p := range permissions {
		var namespaces []string
		_ = json.Unmarshal([]byte(p.Namespaces), &namespaces)
		result = append(result, dto.ClusterPermission{
			Cluster:    p.Cluster,
			Namespaces: namespaces,
		})
	}
	return result, nil
}

// generateToken 生成 JWT Token
func (s *AuthService) generateToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"name":     user.Name,
		"role":     user.Role,
		"iss":      s.cfg.JWT.Issuer,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Duration(s.cfg.JWT.ExpireTime) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

// buildUserResponse 构建用户响应
func (s *AuthService) buildUserResponse(user *entity.User, permissions []dto.ClusterPermission) *dto.UserResponse {
	resp := &dto.UserResponse{
		ID:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		AvatarURL:  user.AvatarURL,
		Mobile:     user.Mobile,
		EmployeeID: user.EmployeeID,
		Department: &dto.DepartmentInfo{
			ID:   user.DepartmentID,
			Name: user.DepartmentName,
			Path: user.DepartmentPath,
		},
		Role:        user.Role,
		IsAdmin:     user.IsAdmin(),
		Permissions: permissions,
	}
	return resp
}
