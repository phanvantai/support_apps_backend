package repositories

import (
	"support-app-backend/internal/models"

	"gorm.io/gorm"
)

// SupportRequestRepository defines the interface for support request data operations
type SupportRequestRepository interface {
	Create(request *models.SupportRequest) error
	GetByID(id uint) (*models.SupportRequest, error)
	GetAll(offset, limit int) ([]*models.SupportRequest, int64, error)
	Update(request *models.SupportRequest) error
	Delete(id uint) error
}

// supportRequestRepository implements SupportRequestRepository
type supportRequestRepository struct {
	db *gorm.DB
}

// NewSupportRequestRepository creates a new support request repository
func NewSupportRequestRepository(db *gorm.DB) SupportRequestRepository {
	return &supportRequestRepository{
		db: db,
	}
}

// Create creates a new support request
func (r *supportRequestRepository) Create(request *models.SupportRequest) error {
	return r.db.Create(request).Error
}

// GetByID retrieves a support request by ID
func (r *supportRequestRepository) GetByID(id uint) (*models.SupportRequest, error) {
	var request models.SupportRequest
	err := r.db.First(&request, id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// GetAll retrieves all support requests with pagination
func (r *supportRequestRepository) GetAll(offset, limit int) ([]*models.SupportRequest, int64, error) {
	var requests []*models.SupportRequest
	var total int64

	// Count total records
	if err := r.db.Model(&models.SupportRequest{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

// Update updates a support request
func (r *supportRequestRepository) Update(request *models.SupportRequest) error {
	return r.db.Save(request).Error
}

// Delete soft deletes a support request
func (r *supportRequestRepository) Delete(id uint) error {
	return r.db.Delete(&models.SupportRequest{}, id).Error
}
