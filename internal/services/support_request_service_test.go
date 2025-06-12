package services

import (
	"errors"
	"support-app-backend/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSupportRequestRepository is a mock implementation of SupportRequestRepository
type MockSupportRequestRepository struct {
	mock.Mock
}

func (m *MockSupportRequestRepository) Create(request *models.SupportRequest) error {
	args := m.Called(request)
	return args.Error(0)
}

func (m *MockSupportRequestRepository) GetByID(id uint) (*models.SupportRequest, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SupportRequest), args.Error(1)
}

func (m *MockSupportRequestRepository) GetAll(offset, limit int) ([]*models.SupportRequest, int64, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]*models.SupportRequest), args.Get(1).(int64), args.Error(2)
}

func (m *MockSupportRequestRepository) Update(request *models.SupportRequest) error {
	args := m.Called(request)
	return args.Error(0)
}

func (m *MockSupportRequestRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestSupportRequestService_CreateSupportRequest(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	userEmail := "test@example.com"
	request := &models.CreateSupportRequestRequest{
		Type:        models.SupportRequestTypeSupport,
		UserEmail:   &userEmail,
		Message:     "Test message",
		Platform:    models.PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
	}

	mockRepo.On("Create", mock.AnythingOfType("*models.SupportRequest")).Return(nil).Run(func(args mock.Arguments) {
		// Simulate setting ID and timestamps
		req := args.Get(0).(*models.SupportRequest)
		req.ID = 1
	})

	// Act
	response, err := service.CreateSupportRequest(request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, uint(1), response.ID)
	assert.Equal(t, models.SupportRequestTypeSupport, response.Type)
	assert.Equal(t, "Test message", response.Message)
	assert.Equal(t, models.StatusNew, response.Status)
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_CreateSupportRequest_NilRequest(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	// Act
	response, err := service.CreateSupportRequest(nil)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRequest, err)
	assert.Nil(t, response)
}

func TestSupportRequestService_CreateSupportRequest_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	userEmail := "test@example.com"
	request := &models.CreateSupportRequestRequest{
		Type:        models.SupportRequestTypeSupport,
		UserEmail:   &userEmail,
		Message:     "Test message",
		Platform:    models.PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
	}

	mockRepo.On("Create", mock.AnythingOfType("*models.SupportRequest")).Return(errors.New("database error"))

	// Act
	response, err := service.CreateSupportRequest(request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_GetSupportRequest(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	userEmail := "test@example.com"
	supportRequest := &models.SupportRequest{
		ID:          1,
		Type:        models.SupportRequestTypeSupport,
		UserEmail:   &userEmail,
		Message:     "Test message",
		Platform:    models.PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
		Status:      models.StatusNew,
	}

	mockRepo.On("GetByID", uint(1)).Return(supportRequest, nil)

	// Act
	response, err := service.GetSupportRequest(1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, uint(1), response.ID)
	assert.Equal(t, models.SupportRequestTypeSupport, response.Type)
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_GetSupportRequest_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	mockRepo.On("GetByID", uint(999)).Return(nil, errors.New("not found"))

	// Act
	response, err := service.GetSupportRequest(999)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrSupportRequestNotFound, err)
	assert.Nil(t, response)
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_GetSupportRequest_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	mockRepo.On("GetByID", uint(1)).Return(nil, errors.New("database error"))

	// Act
	response, err := service.GetSupportRequest(1)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, ErrSupportRequestNotFound, err) // Service converts all repo errors to ErrSupportRequestNotFound
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_GetAllSupportRequests(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	supportRequests := []*models.SupportRequest{
		{
			ID:          1,
			Type:        models.SupportRequestTypeSupport,
			Message:     "Test message 1",
			Platform:    models.PlatformIOS,
			AppVersion:  "1.0.0",
			DeviceModel: "iPhone 13",
			Status:      models.StatusNew,
		},
		{
			ID:          2,
			Type:        models.SupportRequestTypeFeedback,
			Message:     "Test message 2",
			Platform:    models.PlatformAndroid,
			AppVersion:  "1.0.1",
			DeviceModel: "Samsung Galaxy",
			Status:      models.StatusNew,
		},
	}

	mockRepo.On("GetAll", 0, 20).Return(supportRequests, int64(2), nil)

	// Act
	responses, total, err := service.GetAllSupportRequests(1, 20)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, responses, 2)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, uint(1), responses[0].ID)
	assert.Equal(t, uint(2), responses[1].ID)
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_GetAllSupportRequests_InvalidPagination(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	mockRepo.On("GetAll", 0, 20).Return([]*models.SupportRequest{}, int64(0), nil)

	// Act - Test with invalid page and pageSize
	responses, total, err := service.GetAllSupportRequests(0, 0)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, responses)
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_GetAllSupportRequests_EmptyResult(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	mockRepo.On("GetAll", 0, 20).Return([]*models.SupportRequest{}, int64(0), nil)

	// Act
	responses, total, err := service.GetAllSupportRequests(1, 20)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, responses)
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_GetAllSupportRequests_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	mockRepo.On("GetAll", 0, 20).Return([]*models.SupportRequest(nil), int64(0), errors.New("database error"))

	// Act
	responses, total, err := service.GetAllSupportRequests(1, 20)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, responses)
	assert.Equal(t, int64(0), total)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_GetAllSupportRequests_LargePage(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	// Test with very large page size (should be capped to 20)
	mockRepo.On("GetAll", 0, 20).Return([]*models.SupportRequest{}, int64(0), nil)

	// Act
	responses, total, err := service.GetAllSupportRequests(1, 1000)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, responses)
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_GetAllSupportRequests_NegativePage(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	// Test with negative page (should be corrected to page 1)
	mockRepo.On("GetAll", 0, 20).Return([]*models.SupportRequest{}, int64(0), nil)

	// Act
	responses, total, err := service.GetAllSupportRequests(-5, 20)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, responses)
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_UpdateSupportRequest(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	userEmail := "test@example.com"
	originalRequest := &models.SupportRequest{
		ID:          1,
		Type:        models.SupportRequestTypeSupport,
		UserEmail:   &userEmail,
		Message:     "Test message",
		Platform:    models.PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
		Status:      models.StatusNew,
	}

	newStatus := models.StatusInProgress
	adminNotes := "Admin updated this"
	updateRequest := &models.UpdateSupportRequestRequest{
		Status:     &newStatus,
		AdminNotes: &adminNotes,
	}

	mockRepo.On("GetByID", uint(1)).Return(originalRequest, nil)
	mockRepo.On("Update", mock.AnythingOfType("*models.SupportRequest")).Return(nil)

	// Act
	response, err := service.UpdateSupportRequest(1, updateRequest)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, models.StatusInProgress, response.Status)
	assert.Equal(t, "Admin updated this", *response.AdminNotes)
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_UpdateSupportRequest_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	newStatus := models.StatusInProgress
	updateRequest := &models.UpdateSupportRequestRequest{
		Status: &newStatus,
	}

	mockRepo.On("GetByID", uint(999)).Return(nil, errors.New("not found"))

	// Act
	response, err := service.UpdateSupportRequest(999, updateRequest)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrSupportRequestNotFound, err)
	assert.Nil(t, response)
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_UpdateSupportRequest_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	userEmail := "test@example.com"
	originalRequest := &models.SupportRequest{
		ID:          1,
		Type:        models.SupportRequestTypeSupport,
		UserEmail:   &userEmail,
		Message:     "Test message",
		Platform:    models.PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
		Status:      models.StatusNew,
	}

	newStatus := models.StatusInProgress
	updateRequest := &models.UpdateSupportRequestRequest{
		Status: &newStatus,
	}

	mockRepo.On("GetByID", uint(1)).Return(originalRequest, nil)
	mockRepo.On("Update", mock.AnythingOfType("*models.SupportRequest")).Return(errors.New("database error"))

	// Act
	response, err := service.UpdateSupportRequest(1, updateRequest)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_DeleteSupportRequest(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	supportRequest := &models.SupportRequest{
		ID:     1,
		Status: models.StatusNew,
	}

	mockRepo.On("GetByID", uint(1)).Return(supportRequest, nil)
	mockRepo.On("Delete", uint(1)).Return(nil)

	// Act
	err := service.DeleteSupportRequest(1)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_DeleteSupportRequest_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	mockRepo.On("GetByID", uint(999)).Return(nil, errors.New("not found"))

	// Act
	err := service.DeleteSupportRequest(999)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrSupportRequestNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_DeleteSupportRequest_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	mockRepo.On("GetByID", uint(1)).Return(&models.SupportRequest{ID: 1}, nil)
	mockRepo.On("Delete", uint(1)).Return(errors.New("database error"))

	// Act
	err := service.DeleteSupportRequest(1)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestSupportRequestService_UpdateSupportRequest_NilRequest(t *testing.T) {
	// Arrange
	mockRepo := new(MockSupportRequestRepository)
	service := NewSupportRequestService(mockRepo)

	// Act
	response, err := service.UpdateSupportRequest(1, nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, ErrInvalidRequest, err) // Service returns ErrInvalidRequest for nil requests
}
