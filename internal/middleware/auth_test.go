package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"support-app-backend/internal/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of AuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) CreateUser(req *models.CreateUserRequest) (*models.UserInfo, error) {
	args := m.Called(req)
	return args.Get(0).(*models.UserInfo), args.Error(1)
}

func (m *MockAuthService) GetUserByID(id uint) (*models.UserInfo, error) {
	args := m.Called(id)
	return args.Get(0).(*models.UserInfo), args.Error(1)
}

func (m *MockAuthService) GetAllUsers(page, pageSize int) ([]*models.UserInfo, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]*models.UserInfo), args.Get(1).(int64), args.Error(2)
}

func (m *MockAuthService) UpdateUser(id uint, req *models.UpdateUserRequest) (*models.UserInfo, error) {
	args := m.Called(id, req)
	return args.Get(0).(*models.UserInfo), args.Error(1)
}

func (m *MockAuthService) ChangePassword(userID uint, req *models.ChangePasswordRequest) error {
	args := m.Called(userID, req)
	return args.Error(0)
}

func (m *MockAuthService) DeleteUser(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*models.User, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) CreateDefaultAdmin() error {
	args := m.Called()
	return args.Error(0)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock auth service
	mockAuthService := new(MockAuthService)

	// Mock user to return on valid token
	mockUser := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     models.UserRoleAdmin,
		IsActive: true,
	}

	// Setup mock expectations
	mockAuthService.On("ValidateToken", "valid-token").Return(mockUser, nil)

	// Setup router with auth middleware
	router := gin.New()
	router.Use(AuthMiddleware(mockAuthService))
	router.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")
		role, _ := c.Get("role")

		c.JSON(200, gin.H{
			"user_id":  userID,
			"username": username,
			"role":     role,
		})
	})

	// Make request with valid token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := new(MockAuthService)

	router := gin.New()
	router.Use(AuthMiddleware(mockAuthService))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := new(MockAuthService)

	// Setup mock to return error for invalid token
	mockAuthService.On("ValidateToken", "invalid-token").Return(nil, errors.New("invalid token"))

	router := gin.New()
	router.Use(AuthMiddleware(mockAuthService))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthMiddleware_MalformedAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := new(MockAuthService)

	router := gin.New()
	router.Use(AuthMiddleware(mockAuthService))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "invalid-auth-header")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminOnlyMiddleware_AdminUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	router.Use(AdminOnlyMiddleware())
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "admin access granted"})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminOnlyMiddleware_NonAdminUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "user")
		c.Next()
	})
	router.Use(AdminOnlyMiddleware())
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "admin access granted"})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminOnlyMiddleware_NoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(AdminOnlyMiddleware())
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "admin access granted"})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
