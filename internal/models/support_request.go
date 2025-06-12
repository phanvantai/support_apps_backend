package models

import (
	"time"

	"gorm.io/gorm"
)

// SupportRequestType represents the type of request
type SupportRequestType string

const (
	SupportRequestTypeSupport  SupportRequestType = "support"
	SupportRequestTypeFeedback SupportRequestType = "feedback"
)

// Platform represents the mobile platform
type Platform string

const (
	PlatformIOS     Platform = "iOS"
	PlatformAndroid Platform = "Android"
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
	Type        SupportRequestType `json:"type" gorm:"not null" binding:"required,oneof=support feedback"`
	UserEmail   *string            `json:"user_email,omitempty" gorm:"type:varchar(255)"`
	Message     string             `json:"message" gorm:"not null;type:text" binding:"required"`
	Platform    Platform           `json:"platform" gorm:"not null" binding:"required,oneof=iOS Android"`
	AppVersion  string             `json:"app_version" gorm:"not null" binding:"required"`
	DeviceModel string             `json:"device_model" gorm:"not null" binding:"required"`
	Status      Status             `json:"status" gorm:"not null;default:new"`
	AdminNotes  *string            `json:"admin_notes,omitempty" gorm:"type:text"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `json:"-" gorm:"index"`
}

// CreateSupportRequestRequest represents the payload for creating a support request
type CreateSupportRequestRequest struct {
	Type        SupportRequestType `json:"type" binding:"required,oneof=support feedback"`
	UserEmail   *string            `json:"user_email,omitempty"`
	Message     string             `json:"message" binding:"required"`
	Platform    Platform           `json:"platform" binding:"required,oneof=iOS Android"`
	AppVersion  string             `json:"app_version" binding:"required"`
	DeviceModel string             `json:"device_model" binding:"required"`
}

// UpdateSupportRequestRequest represents the payload for updating a support request
type UpdateSupportRequestRequest struct {
	Status     *Status `json:"status,omitempty" binding:"omitempty,oneof=new in_progress resolved"`
	AdminNotes *string `json:"admin_notes,omitempty"`
}

// SupportRequestResponse represents the API response for support requests
type SupportRequestResponse struct {
	ID          uint               `json:"id"`
	Type        SupportRequestType `json:"type"`
	UserEmail   *string            `json:"user_email,omitempty"`
	Message     string             `json:"message"`
	Platform    Platform           `json:"platform"`
	AppVersion  string             `json:"app_version"`
	DeviceModel string             `json:"device_model"`
	Status      Status             `json:"status"`
	AdminNotes  *string            `json:"admin_notes,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
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
