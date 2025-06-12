package services

import (
	"errors"
	"support-app-backend/internal/models"
	"support-app-backend/internal/repositories"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrInvalidToken       = errors.New("invalid token")
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Login(req *models.LoginRequest) (*models.LoginResponse, error)
	CreateUser(req *models.CreateUserRequest) (*models.UserInfo, error)
	GetUserByID(id uint) (*models.UserInfo, error)
	GetAllUsers(page, pageSize int) ([]*models.UserInfo, int64, error)
	UpdateUser(id uint, req *models.UpdateUserRequest) (*models.UserInfo, error)
	ChangePassword(userID uint, req *models.ChangePasswordRequest) error
	DeleteUser(id uint) error
	ValidateToken(tokenString string) (*models.User, error)
	CreateDefaultAdmin() error
}

// authService implements AuthService
type authService struct {
	userRepo  repositories.UserRepository
	jwtSecret string
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repositories.UserRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Login authenticates a user and returns a JWT token
func (s *authService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	if req == nil {
		return nil, ErrInvalidRequest
	}

	// Get user by username
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		return nil, ErrInvalidCredentials
	}

	// Update last login
	s.userRepo.UpdateLastLogin(user.ID)

	// Generate JWT token
	token, expiresAt, err := s.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.ToUserInfo(),
	}, nil
}

// CreateUser creates a new user
func (s *authService) CreateUser(req *models.CreateUserRequest) (*models.UserInfo, error) {
	if req == nil {
		return nil, ErrInvalidRequest
	}

	// Check if user already exists
	exists, err := s.userRepo.UserExists(req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// Create user
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     req.Role,
		IsActive: true,
	}

	// Hash password
	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}

	// Save user
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	userInfo := user.ToUserInfo()
	return &userInfo, nil
}

// GetUserByID retrieves a user by ID
func (s *authService) GetUserByID(id uint) (*models.UserInfo, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	userInfo := user.ToUserInfo()
	return &userInfo, nil
}

// GetAllUsers retrieves all users with pagination
func (s *authService) GetAllUsers(page, pageSize int) ([]*models.UserInfo, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	users, total, err := s.userRepo.GetAll(offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	userInfos := make([]*models.UserInfo, len(users))
	for i, user := range users {
		userInfo := user.ToUserInfo()
		userInfos[i] = &userInfo
	}

	return userInfos, total, nil
}

// UpdateUser updates a user
func (s *authService) UpdateUser(id uint, req *models.UpdateUserRequest) (*models.UserInfo, error) {
	if req == nil {
		return nil, ErrInvalidRequest
	}

	// Get existing user
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Update fields if provided
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Save updated user
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	userInfo := user.ToUserInfo()
	return &userInfo, nil
}

// ChangePassword changes a user's password
func (s *authService) ChangePassword(userID uint, req *models.ChangePasswordRequest) error {
	if req == nil {
		return ErrInvalidRequest
	}

	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Verify current password
	if !user.CheckPassword(req.CurrentPassword) {
		return ErrInvalidCredentials
	}

	// Set new password
	if err := user.SetPassword(req.NewPassword); err != nil {
		return err
	}

	// Save user
	return s.userRepo.Update(user)
}

// DeleteUser deletes a user
func (s *authService) DeleteUser(id uint) error {
	// Check if user exists
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return s.userRepo.Delete(id)
}

// ValidateToken validates a JWT token and returns the user
func (s *authService) ValidateToken(tokenString string) (*models.User, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Get user from database
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	return user, nil
}

// CreateDefaultAdmin creates the default admin account if it doesn't exist
func (s *authService) CreateDefaultAdmin() error {
	// Check if admin already exists
	_, err := s.userRepo.GetByUsername("admin")
	if err == nil {
		// Admin already exists
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Some other error occurred
		return err
	}

	// Create default admin
	adminUser := &models.User{
		Username: "admin",
		Email:    "admin@supportapp.local",
		Role:     models.UserRoleAdmin,
		IsActive: true,
	}

	// Set password
	if err := adminUser.SetPassword("securePassword@123"); err != nil {
		return err
	}

	// Save admin user
	return s.userRepo.Create(adminUser)
}

// generateJWT generates a JWT token for the user
func (s *authService) generateJWT(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour)

	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// JWTClaims represents the JWT claims (moved from middleware to be shared)
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
