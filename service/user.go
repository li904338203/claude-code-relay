package service

import (
	"claude-code-relay/common"
	"claude-code-relay/constant"
	"claude-code-relay/model"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

type LoginResult struct {
	Token string          `json:"token"`
	User  *model.UserInfo `json:"user"`
}

type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// Login 用户登录 (保留原方法以兼容已有代码)
func (s *UserService) Login(username, password string, c *gin.Context) (*LoginResult, error) {
	return s.LoginWithPassword(username, "", password, c)
}

// LoginWithPassword 使用密码登录（支持用户名或邮箱）
func (s *UserService) LoginWithPassword(username, email, password string, c *gin.Context) (*LoginResult, error) {
	var user *model.User
	var err error

	// 根据提供的参数确定查询方式
	if username != "" {
		user, err = model.GetUserByUsername(username)
		if err != nil || user.Password != common.HashPassword(password) {
			return nil, errors.New("用户名或密码错误")
		}
	} else if email != "" {
		user, err = model.GetUserByEmail(email)
		if err != nil || user.Password != common.HashPassword(password) {
			return nil, errors.New("邮箱或密码错误")
		}
	} else {
		return nil, errors.New("用户名或邮箱必须提供其中一个")
	}

	if user.Status != constant.UserStatusActive {
		return nil, errors.New("账户已被禁用")
	}

	return s.generateLoginResult(user)
}

// LoginWithVerificationCode 使用验证码登录
func (s *UserService) LoginWithVerificationCode(email, verificationCode string, c *gin.Context) (*LoginResult, error) {
	if email == "" || verificationCode == "" {
		return nil, errors.New("邮箱和验证码不能为空")
	}

	// 验证验证码（使用登录验证码类型）
	err := common.VerifyCode(email, verificationCode, common.LoginVerification)
	if err != nil {
		// 如果登录类型验证码不存在，尝试使用邮箱验证类型的验证码
		err = common.VerifyCode(email, verificationCode, common.EmailVerification)
		if err != nil {
			return nil, err
		}
	}

	// 根据邮箱查询用户
	user, err := model.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if user.Status != constant.UserStatusActive {
		return nil, errors.New("账户已被禁用")
	}

	return s.generateLoginResult(user)
}

// generateLoginResult 生成登录结果
func (s *UserService) generateLoginResult(user *model.User) (*LoginResult, error) {
	// 生成JWT token
	token, err := common.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, errors.New("生成token失败")
	}

	result := &LoginResult{
		Token: token,
		User: &model.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}

	return result, nil
}

// Register 注册新用户
func (s *UserService) Register(username, email, password string) error {
	// 检查用户名是否已存在
	if _, err := model.GetUserByUsername(username); err == nil {
		return errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if _, err := model.GetUserByEmail(email); err == nil {
		return errors.New("邮箱已存在")
	}

	user := &model.User{
		Username: username,
		Email:    email,
		Password: common.HashPassword(password), // 实际项目中应该加密
		Role:     constant.RoleUser,
		Status:   constant.UserStatusActive,
	}

	if err := model.CreateUser(user); err != nil {
		return errors.New("注册失败")
	}

	// 为新用户创建余额记录
	userBalance := &model.UserBalance{
		UserID:         user.ID,
		Balance:        0,
		FrozenBalance:  0,
		TotalRecharged: 0,
		TotalConsumed:  0,
	}
	if err := model.DB.Create(userBalance).Error; err != nil {
		common.SysError(fmt.Sprintf("Failed to create user balance for user %d: %v", user.ID, err))
		// 不影响注册流程，只记录错误
	}

	return nil
}

func (s *UserService) GetProfile(user *model.User) *model.UserProfile {
	return &model.UserProfile{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.String(),
	}
}

// UpdateProfile 更新用户信息
func (s *UserService) UpdateProfile(currentUser *model.User, username, email, password string) error {
	// 如果要更新用户名，检查是否已存在
	if username != "" && username != currentUser.Username {
		if _, err := model.GetUserByUsername(username); err == nil {
			return errors.New("用户名已存在")
		}
		currentUser.Username = username
	}

	// 如果要更新邮箱，检查是否已存在
	if email != "" && email != currentUser.Email {
		if _, err := model.GetUserByEmail(email); err == nil {
			return errors.New("邮箱已存在")
		}
		currentUser.Email = email
	}

	// 如果要更新密码，进行加密
	if password != "" {
		currentUser.Password = common.HashPassword(password)
	}

	// 保存更新
	if err := model.UpdateUser(currentUser); err != nil {
		return errors.New("更新失败")
	}

	return nil
}

// ChangeEmail 更改用户邮箱
func (s *UserService) ChangeEmail(currentUser *model.User, newEmail, password, verificationCode string) error {
	// 验证当前密码
	if currentUser.Password != common.HashPassword(password) {
		return errors.New("当前密码错误")
	}

	// 检查新邮箱是否与当前邮箱相同
	if newEmail == currentUser.Email {
		return errors.New("新邮箱不能与当前邮箱相同")
	}

	// 验证新邮箱的验证码
	err := common.VerifyCode(newEmail, verificationCode, common.EmailChange)
	if err != nil {
		return err
	}

	// 检查新邮箱是否已被其他用户使用
	if existingUser, err := model.GetUserByEmail(newEmail); err == nil && existingUser.ID != currentUser.ID {
		return errors.New("邮箱已被其他用户使用")
	}

	// 更新邮箱
	currentUser.Email = newEmail
	if err := model.UpdateUser(currentUser); err != nil {
		return errors.New("更新邮箱失败")
	}

	// 记录日志
	common.SysLog(fmt.Sprintf("用户 %s (ID: %d) 更改邮箱为: %s", currentUser.Username, currentUser.ID, newEmail))

	return nil
}

func (s *UserService) GetUsers(page, limit int) (*model.UserListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := model.GetUsers(page, limit)
	if err != nil {
		return nil, errors.New("获取用户列表失败")
	}

	result := &model.UserListResult{
		Users: users,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	return result, nil
}

// AdminCreateUser 管理员创建用户
func (s *UserService) AdminCreateUser(username, email, password, role string) error {
	// 检查用户名是否已存在
	if _, err := model.GetUserByUsername(username); err == nil {
		return errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if _, err := model.GetUserByEmail(email); err == nil {
		return errors.New("邮箱已存在")
	}

	// 验证角色是否有效
	if role != constant.RoleUser && role != constant.RoleAdmin {
		return errors.New("角色参数无效")
	}

	user := &model.User{
		Username: username,
		Email:    email,
		Password: common.HashPassword(password),
		Role:     role,
		Status:   constant.UserStatusActive,
	}

	if err := model.CreateUser(user); err != nil {
		return errors.New("创建用户失败")
	}

	// 为新用户创建余额记录
	userBalance := &model.UserBalance{
		UserID:         user.ID,
		Balance:        0,
		FrozenBalance:  0,
		TotalRecharged: 0,
		TotalConsumed:  0,
	}
	if err := model.DB.Create(userBalance).Error; err != nil {
		common.SysError(fmt.Sprintf("Failed to create user balance for user %d: %v", user.ID, err))
		// 不影响创建流程，只记录错误
	}

	return nil
}

// AdminUpdateUserStatus 管理员更新用户状态
func (s *UserService) AdminUpdateUserStatus(userID uint, status int) error {
	// 获取用户
	user, err := model.GetUserById(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证状态参数
	if status != constant.UserStatusActive && status != constant.UserStatusInactive {
		return errors.New("状态参数无效")
	}

	// 更新状态
	user.Status = status
	if err := model.UpdateUser(user); err != nil {
		return errors.New("更新用户状态失败")
	}

	return nil
}
