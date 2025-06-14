package models

import (
	"time"

	"gorm.io/gorm"
)

// SupportRequestType represents the type of request
type SupportRequestType string

const (
	SupportRequestTypeSupport        SupportRequestType = "support"
	SupportRequestTypeFeedback       SupportRequestType = "feedback"
	SupportRequestTypeBugReport      SupportRequestType = "bug_report"
	SupportRequestTypeFeatureRequest SupportRequestType = "feature_request"
)

// Platform represents the platform
type Platform string

const (
	PlatformIOS     Platform = "iOS"
	PlatformAndroid Platform = "Android"
	PlatformWeb     Platform = "Web"
)

// Status represents the request status
type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusResolved   Status = "resolved"
)

// SupportRequest represents a support ticket or feedback request
type SupportRequest struct {
	ID          uint               `json:"id" gorm:"primaryKey"`
	Type        SupportRequestType `json:"type" gorm:"not null" binding:"required,oneof=support feedback bug_report feature_request"`
	UserEmail   *string            `json:"user_email,omitempty" gorm:"type:varchar(255)"`
	Message     string             `json:"message" gorm:"not null;type:text" binding:"required"`
	Platform    Platform           `json:"platform" gorm:"not null" binding:"required,oneof=iOS Android Web"`
	AppVersion  string             `json:"app_version" gorm:"not null" binding:"required"`
	DeviceModel string             `json:"device_model" gorm:"not null" binding:"required"`
	App         string             `json:"app" gorm:"not null" binding:"required"`
	Status      Status             `json:"status" gorm:"not null;default:new"`
	AdminNotes  *string            `json:"admin_notes,omitempty" gorm:"type:text"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `json:"-" gorm:"index"`
}

// CreateSupportRequestRequest represents the payload for creating a support request
// @Description Request payload for creating a new support request
type CreateSupportRequestRequest struct {
	Type        SupportRequestType `json:"type" binding:"required,oneof=support feedback bug_report feature_request" example:"support"`               // Type of request (support, feedback, bug_report, or feature_request)
	UserEmail   *string            `json:"user_email,omitempty" example:"user@example.com"`                                // Optional user email
	Message     string             `json:"message" binding:"required" example:"I'm having trouble with the login feature"` // Support request message
	Platform    Platform           `json:"platform" binding:"required,oneof=iOS Android Web" example:"iOS"`                    // Platform (iOS, Android, or Web)
	AppVersion  string             `json:"app_version" binding:"required" example:"1.2.3"`                                 // Application version
	DeviceModel string             `json:"device_model" binding:"required" example:"iPhone 14 Pro"`                        // Device model
	App         string             `json:"app" binding:"required" example:"my-awesome-app"`                                // Application name
}

// UpdateSupportRequestRequest represents the payload for updating a support request
// @Description Request payload for updating support request status and admin notes
type UpdateSupportRequestRequest struct {
	Status     *Status `json:"status,omitempty" binding:"omitempty,oneof=new in_progress resolved" example:"in_progress"` // New status
	AdminNotes *string `json:"admin_notes,omitempty" example:"Contacted user for more details"`                           // Admin notes
}

// SupportRequestResponse represents the API response for support requests
// @Description Support request response with all details
type SupportRequestResponse struct {
	ID          uint               `json:"id" example:"1"`                                                  // Support request ID
	Type        SupportRequestType `json:"type" example:"support"`                                          // Type of request
	UserEmail   *string            `json:"user_email,omitempty" example:"user@example.com"`                 // User email (optional)
	Message     string             `json:"message" example:"I'm having trouble with the login feature"`     // Support request message
	Platform    Platform           `json:"platform" example:"iOS"`                                          // Platform (iOS, Android, or Web)
	AppVersion  string             `json:"app_version" example:"1.2.3"`                                     // Application version
	DeviceModel string             `json:"device_model" example:"iPhone 14 Pro"`                            // Device model
	App         string             `json:"app" example:"my-awesome-app"`                                    // Application name
	Status      Status             `json:"status" example:"new"`                                            // Current status
	AdminNotes  *string            `json:"admin_notes,omitempty" example:"Contacted user for more details"` // Admin notes (optional)
	CreatedAt   time.Time          `json:"created_at" example:"2023-12-01T10:00:00Z"`                       // Creation timestamp
	UpdatedAt   time.Time          `json:"updated_at" example:"2023-12-01T10:00:00Z"`                       // Last update timestamp
}

// ToResponse converts SupportRequest to SupportRequestResponse
func (sr *SupportRequest) ToResponse() *SupportRequestResponse {
	return &SupportRequestResponse{
		ID:          sr.ID,
		Type:        sr.Type,
		UserEmail:   sr.UserEmail,
		Message:     sr.Message,
		Platform:    sr.Platform,
		AppVersion:  sr.AppVersion,
		DeviceModel: sr.DeviceModel,
		App:         sr.App,
		Status:      sr.Status,
		AdminNotes:  sr.AdminNotes,
		CreatedAt:   sr.CreatedAt,
		UpdatedAt:   sr.UpdatedAt,
	}
}

// TableName returns the table name for GORM
func (SupportRequest) TableName() string {
	return "support_requests"
}
