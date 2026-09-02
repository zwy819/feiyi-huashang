package service

import (
	"errors"
	"strings"

	"nonlegacy-garment-3d/server/internal/middleware"
	"nonlegacy-garment-3d/server/internal/model"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userStore  *model.UserStore
	jwtSecret  string
	jwtExpires string
}

func NewAuthService(userStore *model.UserStore, jwtSecret, jwtExpires string) *AuthService {
	return &AuthService{userStore: userStore, jwtSecret: jwtSecret, jwtExpires: jwtExpires}
}

type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResult struct {
	Token    string      `json:"token"`
	ID       int64       `json:"id"`
	Username string      `json:"username"`
	Email    string      `json:"email"`
	Role     string      `json:"role"`
}

func (s *AuthService) Register(input RegisterInput) (*AuthResult, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(input.Email)

	if input.Username == "" || input.Email == "" || input.Password == "" {
		return nil, errors.New("用户名、邮箱和密码不能为空")
	}
	if len(input.Password) < 6 {
		return nil, errors.New("密码长度不能少于6位")
	}

	// 检查用户名是否已存在
	if _, err := s.userStore.GetByUsername(input.Username); err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if _, err := s.userStore.GetByEmail(input.Email); err == nil {
		return nil, errors.New("邮箱已被注册")
	}

	// 加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	user, err := s.userStore.Create(input.Username, input.Email, string(hash))
	if err != nil {
		return nil, errors.New("创建用户失败")
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role, s.jwtSecret, s.jwtExpires)
	if err != nil {
		return nil, errors.New("生成令牌失败")
	}

	return &AuthResult{
		Token:    token,
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}, nil
}

func (s *AuthService) Login(input LoginInput) (*AuthResult, error) {
	input.Username = strings.TrimSpace(input.Username)

	if input.Username == "" || input.Password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}

	user, err := s.userStore.GetByUsername(input.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role, s.jwtSecret, s.jwtExpires)
	if err != nil {
		return nil, errors.New("生成令牌失败")
	}

	return &AuthResult{
		Token:    token,
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}, nil
}

func (s *AuthService) GetProfile(userID int64) (*model.User, error) {
	return s.userStore.GetByID(userID)
}
