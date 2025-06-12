package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"support-app-backend/internal/models"
	"support-app-backend/internal/services"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) CreateUser(req *models.CreateUserRequest) (*models.UserInfo, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserInfo), args.Error(1)
}

func (m *MockAuthService) GetUserByID(id uint) (*models.UserInfo, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserInfo), args.Error(1)
}

func (m *MockAuthService) GetAllUsers(page, pageSize int) ([]*models.UserInfo, int64, error) {
	args := m.Called(page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.UserInfo), args.Get(1).(int64), args.Error(2)
}

func (m *MockAuthService) UpdateUser(id uint, req *models.UpdateUserRequest) (*models.UserInfo, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserInfo), args.Error(1)
}

func (m *MockAuthService) DeleteUser(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAuthService) ChangePassword(userID uint, req *models.ChangePasswordRequest) error {
	args := m.Called(userID, req)
	return args.Error(0)
}

func (m *MockAuthService) CreateDefaultAdmin() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*models.User, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func setupAuthHandler() (*AuthHandler, *MockAuthService) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	return handler, mockService
}

func TestAuthHandler_Login_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	loginReq := &models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	loginResp := &models.LoginResponse{
		Token:     "jwt.token.here",
		ExpiresAt: time.Now().Add(time.Hour),
		User: models.UserInfo{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
			Role:     models.UserRoleUser,
			IsActive: true,
		},
	}

	mockService.On("Login", loginReq).Return(loginResp, nil)

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	handler, mockService := setupAuthHandler()

	loginReq := &models.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	mockService.On("Login", loginReq).Return(nil, services.ErrInvalidCredentials)

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_UserInactive(t *testing.T) {
	handler, mockService := setupAuthHandler()

	loginReq := &models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	mockService.On("Login", loginReq).Return(nil, services.ErrUserInactive)

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	handler, _ := setupAuthHandler()

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_InternalServerError(t *testing.T) {
	handler, mockService := setupAuthHandler()

	loginReq := &models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	mockService.On("Login", loginReq).Return(nil, assert.AnError)

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_CreateUser_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	createReq := &models.CreateUserRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
		Role:     models.UserRoleUser,
	}

	userInfo := &models.UserInfo{
		ID:       2,
		Username: "newuser",
		Email:    "new@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}

	mockService.On("CreateUser", createReq).Return(userInfo, nil)

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateUser(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_CreateUser_UserExists(t *testing.T) {
	handler, mockService := setupAuthHandler()

	createReq := &models.CreateUserRequest{
		Username: "existinguser",
		Email:    "existing@example.com",
		Password: "password123",
		Role:     models.UserRoleUser,
	}

	mockService.On("CreateUser", createReq).Return(nil, services.ErrUserExists)

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateUser(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_CreateUser_InvalidJSON(t *testing.T) {
	handler, _ := setupAuthHandler()

	req := httptest.NewRequest(http.MethodPost, "/auth/users", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_CreateUser_InvalidRequest(t *testing.T) {
	handler, mockService := setupAuthHandler()

	createReq := &models.CreateUserRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
		Role:     models.UserRoleUser,
	}

	mockService.On("CreateUser", createReq).Return(nil, services.ErrInvalidRequest)

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_CreateUser_InternalServerError(t *testing.T) {
	handler, mockService := setupAuthHandler()

	createReq := &models.CreateUserRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
		Role:     models.UserRoleUser,
	}

	mockService.On("CreateUser", createReq).Return(nil, assert.AnError)

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.CreateUser(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetUser_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	userInfo := &models.UserInfo{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}

	mockService.On("GetUserByID", uint(1)).Return(userInfo, nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/users/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.GetUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetUser_InvalidID(t *testing.T) {
	handler, _ := setupAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/auth/users/invalid", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	handler.GetUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_GetUser_NotFound(t *testing.T) {
	handler, mockService := setupAuthHandler()

	mockService.On("GetUserByID", uint(999)).Return(nil, services.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodGet, "/auth/users/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	handler.GetUser(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetUser_InternalServerError(t *testing.T) {
	handler, mockService := setupAuthHandler()

	mockService.On("GetUserByID", uint(1)).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/auth/users/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.GetUser(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetAllUsers_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	users := []*models.UserInfo{
		{ID: 1, Username: "user1", Email: "user1@example.com", Role: models.UserRoleUser, IsActive: true},
		{ID: 2, Username: "user2", Email: "user2@example.com", Role: models.UserRoleAdmin, IsActive: true},
	}

	mockService.On("GetAllUsers", 1, 20).Return(users, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/users", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GetAllUsers(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "pagination")

	pagination := response["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(20), pagination["page_size"])
	assert.Equal(t, float64(2), pagination["total"])
	assert.Equal(t, float64(1), pagination["total_pages"])

	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetAllUsers_InternalServerError(t *testing.T) {
	handler, mockService := setupAuthHandler()

	mockService.On("GetAllUsers", 1, 20).Return(nil, int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/auth/users", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GetAllUsers(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetAllUsers_WithCustomPagination(t *testing.T) {
	handler, mockService := setupAuthHandler()

	users := []*models.UserInfo{
		{ID: 1, Username: "user1", Email: "user1@example.com", Role: models.UserRoleUser, IsActive: true},
	}

	mockService.On("GetAllUsers", 2, 10).Return(users, int64(15), nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/users?page=2&page_size=10", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GetAllUsers(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	pagination := response["pagination"].(map[string]interface{})
	assert.Equal(t, float64(2), pagination["page"])
	assert.Equal(t, float64(10), pagination["page_size"])
	assert.Equal(t, float64(15), pagination["total"])
	assert.Equal(t, float64(2), pagination["total_pages"])

	mockService.AssertExpectations(t)
}

func TestAuthHandler_UpdateUser_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	email := "updated@example.com"
	active := true
	updateReq := &models.UpdateUserRequest{
		Email:    &email,
		IsActive: &active,
	}

	userInfo := &models.UserInfo{
		ID:       1,
		Username: "testuser",
		Email:    "updated@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}

	mockService.On("UpdateUser", uint(1), updateReq).Return(userInfo, nil)

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPatch, "/auth/users/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.UpdateUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_UpdateUser_InvalidJSON(t *testing.T) {
	handler, _ := setupAuthHandler()

	req := httptest.NewRequest(http.MethodPatch, "/auth/users/1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.UpdateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_UpdateUser_InvalidID(t *testing.T) {
	handler, _ := setupAuthHandler()

	updateReq := &models.UpdateUserRequest{}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPatch, "/auth/users/invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	handler.UpdateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_UpdateUser_NotFound(t *testing.T) {
	handler, mockService := setupAuthHandler()

	email := "updated@example.com"
	updateReq := &models.UpdateUserRequest{
		Email: &email,
	}

	mockService.On("UpdateUser", uint(999), updateReq).Return(nil, services.ErrUserNotFound)

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPatch, "/auth/users/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	handler.UpdateUser(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_UpdateUser_InvalidRequest(t *testing.T) {
	handler, mockService := setupAuthHandler()

	email := "updated@example.com"
	updateReq := &models.UpdateUserRequest{
		Email: &email,
	}

	mockService.On("UpdateUser", uint(1), updateReq).Return(nil, services.ErrInvalidRequest)

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPatch, "/auth/users/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.UpdateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_UpdateUser_InternalServerError(t *testing.T) {
	handler, mockService := setupAuthHandler()

	email := "updated@example.com"
	updateReq := &models.UpdateUserRequest{
		Email: &email,
	}

	mockService.On("UpdateUser", uint(1), updateReq).Return(nil, assert.AnError)

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPatch, "/auth/users/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.UpdateUser(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_DeleteUser_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	mockService.On("DeleteUser", uint(1)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/auth/users/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.DeleteUser(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_DeleteUser_InvalidID(t *testing.T) {
	handler, _ := setupAuthHandler()

	req := httptest.NewRequest(http.MethodDelete, "/auth/users/invalid", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	handler.DeleteUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_DeleteUser_NotFound(t *testing.T) {
	handler, mockService := setupAuthHandler()

	mockService.On("DeleteUser", uint(999)).Return(services.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/auth/users/999", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	handler.DeleteUser(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_DeleteUser_InternalServerError(t *testing.T) {
	handler, mockService := setupAuthHandler()

	mockService.On("DeleteUser", uint(1)).Return(assert.AnError)

	req := httptest.NewRequest(http.MethodDelete, "/auth/users/1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.DeleteUser(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	changeReq := &models.ChangePasswordRequest{
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword123",
	}

	mockService.On("ChangePassword", uint(1), changeReq).Return(nil)

	body, _ := json.Marshal(changeReq)
	req := httptest.NewRequest(http.MethodPatch, "/auth/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword_NoUserID(t *testing.T) {
	handler, _ := setupAuthHandler()

	changeReq := &models.ChangePasswordRequest{
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword123",
	}

	body, _ := json.Marshal(changeReq)
	req := httptest.NewRequest(http.MethodPatch, "/auth/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_ChangePassword_InvalidJSON(t *testing.T) {
	handler, _ := setupAuthHandler()

	req := httptest.NewRequest(http.MethodPatch, "/auth/password", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_ChangePassword_UserNotFound(t *testing.T) {
	handler, mockService := setupAuthHandler()

	changeReq := &models.ChangePasswordRequest{
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword123",
	}

	mockService.On("ChangePassword", uint(1), changeReq).Return(services.ErrUserNotFound)

	body, _ := json.Marshal(changeReq)
	req := httptest.NewRequest(http.MethodPatch, "/auth/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword_InvalidRequest(t *testing.T) {
	handler, mockService := setupAuthHandler()

	changeReq := &models.ChangePasswordRequest{
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword123",
	}

	mockService.On("ChangePassword", uint(1), changeReq).Return(services.ErrInvalidRequest)

	body, _ := json.Marshal(changeReq)
	req := httptest.NewRequest(http.MethodPatch, "/auth/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword_InternalServerError(t *testing.T) {
	handler, mockService := setupAuthHandler()

	changeReq := &models.ChangePasswordRequest{
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword123",
	}

	mockService.On("ChangePassword", uint(1), changeReq).Return(assert.AnError)

	body, _ := json.Marshal(changeReq)
	req := httptest.NewRequest(http.MethodPatch, "/auth/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetCurrentUser_Success(t *testing.T) {
	handler, mockService := setupAuthHandler()

	userInfo := &models.UserInfo{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}

	mockService.On("GetUserByID", uint(1)).Return(userInfo, nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler.GetCurrentUser(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetCurrentUser_NoUserID(t *testing.T) {
	handler, _ := setupAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GetCurrentUser(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_GetCurrentUser_NotFound(t *testing.T) {
	handler, mockService := setupAuthHandler()

	mockService.On("GetUserByID", uint(1)).Return(nil, services.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler.GetCurrentUser(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetCurrentUser_InternalServerError(t *testing.T) {
	handler, mockService := setupAuthHandler()

	mockService.On("GetUserByID", uint(1)).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uint(1))

	handler.GetCurrentUser(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
