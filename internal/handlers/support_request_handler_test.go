package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"support-app-backend/internal/models"
	"support-app-backend/internal/services"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSupportRequestService is a mock implementation of SupportRequestService
type MockSupportRequestService struct {
	mock.Mock
}

func (m *MockSupportRequestService) CreateSupportRequest(req *models.CreateSupportRequestRequest) (*models.SupportRequestResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SupportRequestResponse), args.Error(1)
}

func (m *MockSupportRequestService) GetSupportRequest(id uint) (*models.SupportRequestResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SupportRequestResponse), args.Error(1)
}

func (m *MockSupportRequestService) GetAllSupportRequests(page, pageSize int) ([]*models.SupportRequestResponse, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]*models.SupportRequestResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockSupportRequestService) UpdateSupportRequest(id uint, req *models.UpdateSupportRequestRequest) (*models.SupportRequestResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SupportRequestResponse), args.Error(1)
}

func (m *MockSupportRequestService) DeleteSupportRequest(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestSupportRequestHandler_CreateSupportRequest(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.POST("/support-request", handler.CreateSupportRequest)

	userEmail := "test@example.com"
	request := &models.CreateSupportRequestRequest{
		Type:        models.SupportRequestTypeSupport,
		UserEmail:   &userEmail,
		Message:     "Test message",
		Platform:    models.PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
	}

	response := &models.SupportRequestResponse{
		ID:          1,
		Type:        models.SupportRequestTypeSupport,
		UserEmail:   &userEmail,
		Message:     "Test message",
		Platform:    models.PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
		Status:      models.StatusNew,
	}

	mockService.On("CreateSupportRequest", mock.AnythingOfType("*models.CreateSupportRequestRequest")).Return(response, nil)

	requestBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/support-request", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestSupportRequestHandler_CreateSupportRequest_InvalidJSON(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.POST("/support-request", handler.CreateSupportRequest)

	req, _ := http.NewRequest("POST", "/support-request", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSupportRequestHandler_GetSupportRequest(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.GET("/support-requests/:id", handler.GetSupportRequest)

	userEmail := "test@example.com"
	response := &models.SupportRequestResponse{
		ID:          1,
		Type:        models.SupportRequestTypeSupport,
		UserEmail:   &userEmail,
		Message:     "Test message",
		Platform:    models.PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
		Status:      models.StatusNew,
	}

	mockService.On("GetSupportRequest", uint(1)).Return(response, nil)

	req, _ := http.NewRequest("GET", "/support-requests/1", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestSupportRequestHandler_GetSupportRequest_InvalidID(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.GET("/support-requests/:id", handler.GetSupportRequest)

	req, _ := http.NewRequest("GET", "/support-requests/invalid", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSupportRequestHandler_GetSupportRequest_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.GET("/support-requests/:id", handler.GetSupportRequest)

	mockService.On("GetSupportRequest", uint(999)).Return(nil, services.ErrSupportRequestNotFound)

	req, _ := http.NewRequest("GET", "/support-requests/999", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestSupportRequestHandler_GetAllSupportRequests(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.GET("/support-requests", handler.GetAllSupportRequests)

	responses := []*models.SupportRequestResponse{
		{
			ID:          1,
			Type:        models.SupportRequestTypeSupport,
			Message:     "Test message 1",
			Platform:    models.PlatformIOS,
			AppVersion:  "1.0.0",
			DeviceModel: "iPhone 13",
			Status:      models.StatusNew,
		},
	}

	mockService.On("GetAllSupportRequests", 1, 20).Return(responses, int64(1), nil)

	req, _ := http.NewRequest("GET", "/support-requests", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestSupportRequestHandler_UpdateSupportRequest(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.PATCH("/support-requests/:id", handler.UpdateSupportRequest)

	newStatus := models.StatusInProgress
	updateRequest := &models.UpdateSupportRequestRequest{
		Status: &newStatus,
	}

	response := &models.SupportRequestResponse{
		ID:     1,
		Status: models.StatusInProgress,
	}

	mockService.On("UpdateSupportRequest", uint(1), mock.AnythingOfType("*models.UpdateSupportRequestRequest")).Return(response, nil)

	requestBody, _ := json.Marshal(updateRequest)
	req, _ := http.NewRequest("PATCH", "/support-requests/1", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestSupportRequestHandler_UpdateSupportRequest_InvalidID(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.PATCH("/support-requests/:id", handler.UpdateSupportRequest)

	req, _ := http.NewRequest("PATCH", "/support-requests/invalid", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid ID format", response["error"])
}

func TestSupportRequestHandler_UpdateSupportRequest_InvalidJSON(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.PATCH("/support-requests/:id", handler.UpdateSupportRequest)

	req, _ := http.NewRequest("PATCH", "/support-requests/1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSupportRequestHandler_UpdateSupportRequest_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.PATCH("/support-requests/:id", handler.UpdateSupportRequest)

	mockService.On("UpdateSupportRequest", uint(999), mock.AnythingOfType("*models.UpdateSupportRequestRequest")).
		Return(nil, services.ErrSupportRequestNotFound)

	req, _ := http.NewRequest("PATCH", "/support-requests/999", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Support request not found", response["error"])
}

func TestSupportRequestHandler_UpdateSupportRequest_InvalidRequest(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.PATCH("/support-requests/:id", handler.UpdateSupportRequest)

	mockService.On("UpdateSupportRequest", uint(1), mock.AnythingOfType("*models.UpdateSupportRequestRequest")).
		Return(nil, services.ErrInvalidRequest)

	req, _ := http.NewRequest("PATCH", "/support-requests/1", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSupportRequestHandler_UpdateSupportRequest_ServiceError(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.PATCH("/support-requests/:id", handler.UpdateSupportRequest)

	mockService.On("UpdateSupportRequest", uint(1), mock.AnythingOfType("*models.UpdateSupportRequestRequest")).
		Return(nil, errors.New("service error"))

	req, _ := http.NewRequest("PATCH", "/support-requests/1", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to update support request", response["error"])
}

func TestSupportRequestHandler_DeleteSupportRequest(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.DELETE("/support-requests/:id", handler.DeleteSupportRequest)

	mockService.On("DeleteSupportRequest", uint(1)).Return(nil)

	req, _ := http.NewRequest("DELETE", "/support-requests/1", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestSupportRequestHandler_DeleteSupportRequest_InvalidID(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.DELETE("/support-requests/:id", handler.DeleteSupportRequest)

	req, _ := http.NewRequest("DELETE", "/support-requests/invalid", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid ID format", response["error"])
}

func TestSupportRequestHandler_DeleteSupportRequest_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.DELETE("/support-requests/:id", handler.DeleteSupportRequest)

	mockService.On("DeleteSupportRequest", uint(999)).Return(services.ErrSupportRequestNotFound)

	req, _ := http.NewRequest("DELETE", "/support-requests/999", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Support request not found", response["error"])
}

func TestSupportRequestHandler_DeleteSupportRequest_ServiceError(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.DELETE("/support-requests/:id", handler.DeleteSupportRequest)

	mockService.On("DeleteSupportRequest", uint(1)).Return(errors.New("service error"))

	req, _ := http.NewRequest("DELETE", "/support-requests/1", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to delete support request", response["error"])
}

func TestSupportRequestHandler_HealthCheck(t *testing.T) {
	// Arrange
	mockService := new(MockSupportRequestService)
	handler := NewSupportRequestHandler(mockService)
	router := setupTestRouter()
	router.GET("/health", handler.HealthCheck)

	req, _ := http.NewRequest("GET", "/health", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "support-app-backend", response["service"])
}
