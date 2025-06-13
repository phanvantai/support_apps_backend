package services

import (
	"errors"
	"support-app-backend/internal/models"
	"support-app-backend/internal/repositories"
)

var (
	ErrSupportRequestNotFound = errors.New("support request not found")
	ErrInvalidRequest         = errors.New("invalid request")
)

// SupportRequestService defines the interface for support request business logic
type SupportRequestService interface {
	CreateSupportRequest(req *models.CreateSupportRequestRequest) (*models.SupportRequestResponse, error)
	GetSupportRequest(id uint) (*models.SupportRequestResponse, error)
	GetAllSupportRequests(page, pageSize int) ([]*models.SupportRequestResponse, int64, error)
	UpdateSupportRequest(id uint, req *models.UpdateSupportRequestRequest) (*models.SupportRequestResponse, error)
	DeleteSupportRequest(id uint) error
}

// supportRequestService implements SupportRequestService
type supportRequestService struct {
	repo repositories.SupportRequestRepository
}

// NewSupportRequestService creates a new support request service
func NewSupportRequestService(repo repositories.SupportRequestRepository) SupportRequestService {
	return &supportRequestService{
		repo: repo,
	}
}

// CreateSupportRequest creates a new support request
func (s *supportRequestService) CreateSupportRequest(req *models.CreateSupportRequestRequest) (*models.SupportRequestResponse, error) {
	if req == nil {
		return nil, ErrInvalidRequest
	}

	// Create the support request model
	supportRequest := &models.SupportRequest{
		Type:        req.Type,
		UserEmail:   req.UserEmail,
		Message:     req.Message,
		Platform:    req.Platform,
		AppVersion:  req.AppVersion,
		DeviceModel: req.DeviceModel,
		App:         req.App, // Fix: Include the App field
		Status:      models.StatusNew, // Always start with 'new' status
	}

	// Save to repository
	if err := s.repo.Create(supportRequest); err != nil {
		return nil, err
	}

	return supportRequest.ToResponse(), nil
}

// GetSupportRequest retrieves a support request by ID
func (s *supportRequestService) GetSupportRequest(id uint) (*models.SupportRequestResponse, error) {
	supportRequest, err := s.repo.GetByID(id)
	if err != nil {
		return nil, ErrSupportRequestNotFound
	}

	return supportRequest.ToResponse(), nil
}

// GetAllSupportRequests retrieves all support requests with pagination
func (s *supportRequestService) GetAllSupportRequests(page, pageSize int) ([]*models.SupportRequestResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Default page size
	}

	offset := (page - 1) * pageSize

	supportRequests, total, err := s.repo.GetAll(offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*models.SupportRequestResponse, len(supportRequests))
	for i, req := range supportRequests {
		responses[i] = req.ToResponse()
	}

	return responses, total, nil
}

// UpdateSupportRequest updates a support request
func (s *supportRequestService) UpdateSupportRequest(id uint, req *models.UpdateSupportRequestRequest) (*models.SupportRequestResponse, error) {
	if req == nil {
		return nil, ErrInvalidRequest
	}

	// Get existing support request
	supportRequest, err := s.repo.GetByID(id)
	if err != nil {
		return nil, ErrSupportRequestNotFound
	}

	// Update fields if provided
	if req.Status != nil {
		supportRequest.Status = *req.Status
	}
	if req.AdminNotes != nil {
		supportRequest.AdminNotes = req.AdminNotes
	}

	// Save updated request
	if err := s.repo.Update(supportRequest); err != nil {
		return nil, err
	}

	return supportRequest.ToResponse(), nil
}

// DeleteSupportRequest deletes a support request
func (s *supportRequestService) DeleteSupportRequest(id uint) error {
	// Check if the support request exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		return ErrSupportRequestNotFound
	}

	return s.repo.Delete(id)
}
