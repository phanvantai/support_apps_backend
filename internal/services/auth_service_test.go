package services

import (
	"support-app-backend/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) GetAll(offset, limit int) ([]*models.User, int64, error) {
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) UpdateLastLogin(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserRepository) UserExists(username, email string) (bool, error) {
	args := m.Called(username, email)
	return args.Bool(0), args.Error(1)
}

func setupAuthService() (AuthService, *MockUserRepository) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-jwt-secret-key-that-is-long-enough-for-testing")
	return service, mockRepo
}

func TestAuthService_Login_Success(t *testing.T) {
	service, mockRepo := setupAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	user.SetPassword("password123")

	req := &models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", "testuser").Return(user, nil)
	mockRepo.On("UpdateLastLogin", uint(1)).Return(nil)

	response, err := service.Login(req)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, user.Username, response.User.Username)
	assert.Equal(t, user.Email, response.User.Email)
	assert.Equal(t, user.Role, response.User.Role)
	assert.Equal(t, user.IsActive, response.User.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_NilRequest(t *testing.T) {
	service, _ := setupAuthService()

	response, err := service.Login(nil)

	assert.Nil(t, response)
	assert.Equal(t, ErrInvalidRequest, err)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	service, mockRepo := setupAuthService()

	req := &models.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	response, err := service.Login(req)

	assert.Nil(t, response)
	assert.Equal(t, ErrInvalidCredentials, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserInactive(t *testing.T) {
	service, mockRepo := setupAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: false,
	}
	user.SetPassword("password123")

	req := &models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", "testuser").Return(user, nil)

	response, err := service.Login(req)

	assert.Nil(t, response)
	assert.Equal(t, ErrUserInactive, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	service, mockRepo := setupAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	user.SetPassword("password123")

	req := &models.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	mockRepo.On("GetByUsername", "testuser").Return(user, nil)

	response, err := service.Login(req)

	assert.Nil(t, response)
	assert.Equal(t, ErrInvalidCredentials, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_CreateUser_Success(t *testing.T) {
	service, mockRepo := setupAuthService()

	req := &models.CreateUserRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
		Role:     models.UserRoleUser,
	}

	mockRepo.On("UserExists", "newuser", "new@example.com").Return(false, nil)
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	response, err := service.CreateUser(req)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, req.Username, response.Username)
	assert.Equal(t, req.Email, response.Email)
	assert.Equal(t, req.Role, response.Role)
	assert.True(t, response.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_CreateUser_NilRequest(t *testing.T) {
	service, _ := setupAuthService()

	response, err := service.CreateUser(nil)

	assert.Nil(t, response)
	assert.Equal(t, ErrInvalidRequest, err)
}

func TestAuthService_CreateUser_UsernameExists(t *testing.T) {
	service, mockRepo := setupAuthService()

	req := &models.CreateUserRequest{
		Username: "existinguser",
		Email:    "new@example.com",
		Password: "password123",
		Role:     models.UserRoleUser,
	}

	mockRepo.On("UserExists", "existinguser", "new@example.com").Return(true, nil)

	response, err := service.CreateUser(req)

	assert.Nil(t, response)
	assert.Equal(t, ErrUserExists, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_CreateUser_EmailExists(t *testing.T) {
	service, mockRepo := setupAuthService()

	req := &models.CreateUserRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "password123",
		Role:     models.UserRoleUser,
	}

	mockRepo.On("UserExists", "newuser", "existing@example.com").Return(true, nil)

	response, err := service.CreateUser(req)

	assert.Nil(t, response)
	assert.Equal(t, ErrUserExists, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_GetUserByID_Success(t *testing.T) {
	service, mockRepo := setupAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}

	mockRepo.On("GetByID", uint(1)).Return(user, nil)

	response, err := service.GetUserByID(1)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, user.ID, response.ID)
	assert.Equal(t, user.Username, response.Username)
	assert.Equal(t, user.Email, response.Email)
	assert.Equal(t, user.Role, response.Role)
	assert.Equal(t, user.IsActive, response.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_GetUserByID_NotFound(t *testing.T) {
	service, mockRepo := setupAuthService()

	mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	response, err := service.GetUserByID(999)

	assert.Nil(t, response)
	assert.Equal(t, ErrUserNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_GetAllUsers_Success(t *testing.T) {
	service, mockRepo := setupAuthService()

	users := []*models.User{
		{ID: 1, Username: "user1", Email: "user1@example.com", Role: models.UserRoleUser, IsActive: true},
		{ID: 2, Username: "user2", Email: "user2@example.com", Role: models.UserRoleAdmin, IsActive: true},
	}

	mockRepo.On("GetAll", 0, 20).Return(users, int64(2), nil)

	response, total, err := service.GetAllUsers(1, 20)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response, 2)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, users[0].Username, response[0].Username)
	assert.Equal(t, users[1].Username, response[1].Username)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_GetAllUsers_InvalidPagination(t *testing.T) {
	service, mockRepo := setupAuthService()

	// The service now auto-corrects invalid pagination, so we need to expect the corrected calls
	// Test negative page (should be corrected to page 1)
	mockRepo.On("GetAll", 0, 20).Return([]*models.User{}, int64(0), nil).Once()
	response, total, err := service.GetAllUsers(-1, 20)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, int64(0), total)

	// Test zero page size (should be corrected to 20)
	mockRepo.On("GetAll", 0, 20).Return([]*models.User{}, int64(0), nil).Once()
	response, total, err = service.GetAllUsers(1, 0)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, int64(0), total)

	// Test large page size (should be corrected to 20)
	mockRepo.On("GetAll", 0, 20).Return([]*models.User{}, int64(0), nil).Once()
	response, total, err = service.GetAllUsers(1, 101)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, int64(0), total)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_UpdateUser_Success(t *testing.T) {
	service, mockRepo := setupAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}

	email := "updated@example.com"
	role := models.UserRoleAdmin
	active := false

	req := &models.UpdateUserRequest{
		Email:    &email,
		Role:     &role,
		IsActive: &active,
	}

	mockRepo.On("GetByID", uint(1)).Return(user, nil)
	mockRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)

	response, err := service.UpdateUser(1, req)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, email, response.Email)
	assert.Equal(t, role, response.Role)
	assert.Equal(t, active, response.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_UpdateUser_NotFound(t *testing.T) {
	service, mockRepo := setupAuthService()

	req := &models.UpdateUserRequest{}

	mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	response, err := service.UpdateUser(999, req)

	assert.Nil(t, response)
	assert.Equal(t, ErrUserNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_DeleteUser_Success(t *testing.T) {
	service, mockRepo := setupAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockRepo.On("GetByID", uint(1)).Return(user, nil)
	mockRepo.On("Delete", uint(1)).Return(nil)

	err := service.DeleteUser(1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_DeleteUser_NotFound(t *testing.T) {
	service, mockRepo := setupAuthService()

	mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	err := service.DeleteUser(999)

	assert.Equal(t, ErrUserNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ChangePassword_Success(t *testing.T) {
	service, mockRepo := setupAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}
	user.SetPassword("oldpassword")

	req := &models.ChangePasswordRequest{
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword123",
	}

	mockRepo.On("GetByID", uint(1)).Return(user, nil)
	mockRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)

	err := service.ChangePassword(1, req)

	assert.NoError(t, err)
	assert.True(t, user.CheckPassword("newpassword123"))
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ChangePassword_UserNotFound(t *testing.T) {
	service, mockRepo := setupAuthService()

	req := &models.ChangePasswordRequest{
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword123",
	}

	mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	err := service.ChangePassword(999, req)

	assert.Equal(t, ErrUserNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ChangePassword_WrongCurrentPassword(t *testing.T) {
	service, mockRepo := setupAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}
	user.SetPassword("oldpassword")

	req := &models.ChangePasswordRequest{
		CurrentPassword: "wrongpassword",
		NewPassword:     "newpassword123",
	}

	mockRepo.On("GetByID", uint(1)).Return(user, nil)

	err := service.ChangePassword(1, req)

	assert.Equal(t, ErrInvalidCredentials, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken_Success(t *testing.T) {
	service, mockRepo := setupAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}

	// Generate a valid token first
	authSvc := service.(*authService)
	token, _, err := authSvc.generateJWT(user)
	require.NoError(t, err)

	mockRepo.On("GetByID", uint(1)).Return(user, nil)

	validatedUser, err := service.ValidateToken(token)

	require.NoError(t, err)
	assert.NotNil(t, validatedUser)
	assert.Equal(t, user.ID, validatedUser.ID)
	assert.Equal(t, user.Username, validatedUser.Username)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	service, _ := setupAuthService()

	validatedUser, err := service.ValidateToken("invalid.token.here")

	assert.Nil(t, validatedUser)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestAuthService_ValidateToken_UserNotFound(t *testing.T) {
	service, mockRepo := setupAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}

	// Generate a valid token first
	authSvc := service.(*authService)
	token, _, err := authSvc.generateJWT(user)
	require.NoError(t, err)

	mockRepo.On("GetByID", uint(1)).Return(nil, gorm.ErrRecordNotFound)

	validatedUser, err := service.ValidateToken(token)

	assert.Nil(t, validatedUser)
	assert.Equal(t, ErrUserNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken_UserInactive(t *testing.T) {
	service, mockRepo := setupAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}

	// Generate a valid token first
	authSvc := service.(*authService)
	token, _, err := authSvc.generateJWT(user)
	require.NoError(t, err)

	// Make user inactive
	user.IsActive = false

	mockRepo.On("GetByID", uint(1)).Return(user, nil)

	validatedUser, err := service.ValidateToken(token)

	assert.Nil(t, validatedUser)
	assert.Equal(t, ErrUserInactive, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_CreateDefaultAdmin_Success(t *testing.T) {
	service, mockRepo := setupAuthService()

	mockRepo.On("GetByUsername", "admin").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	err := service.CreateDefaultAdmin()

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_CreateDefaultAdmin_AlreadyExists(t *testing.T) {
	service, mockRepo := setupAuthService()

	existingAdmin := &models.User{
		ID:       1,
		Username: "admin",
		Email:    "admin@example.com",
		Role:     models.UserRoleAdmin,
	}

	mockRepo.On("GetByUsername", "admin").Return(existingAdmin, nil)

	err := service.CreateDefaultAdmin()

	assert.NoError(t, err) // Should not error if admin already exists
	mockRepo.AssertExpectations(t)
}

func TestAuthService_GenerateJWT_Success(t *testing.T) {
	service, _ := setupAuthService()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
	}

	authSvc := service.(*authService)
	token, expiresAt, err := authSvc.generateJWT(user)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, expiresAt.After(time.Now()))
	assert.True(t, expiresAt.Before(time.Now().Add(25*time.Hour)))
}
