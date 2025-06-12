package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSupportRequest_ToResponse(t *testing.T) {
	// Arrange
	userEmail := "test@example.com"
	adminNotes := "Test admin notes"
	now := time.Now()

	supportRequest := &SupportRequest{
		ID:          1,
		Type:        SupportRequestTypeSupport,
		UserEmail:   &userEmail,
		Message:     "Test message",
		Platform:    PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
		Status:      StatusNew,
		AdminNotes:  &adminNotes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Act
	response := supportRequest.ToResponse()

	// Assert
	assert.Equal(t, supportRequest.ID, response.ID)
	assert.Equal(t, supportRequest.Type, response.Type)
	assert.Equal(t, supportRequest.UserEmail, response.UserEmail)
	assert.Equal(t, supportRequest.Message, response.Message)
	assert.Equal(t, supportRequest.Platform, response.Platform)
	assert.Equal(t, supportRequest.AppVersion, response.AppVersion)
	assert.Equal(t, supportRequest.DeviceModel, response.DeviceModel)
	assert.Equal(t, supportRequest.Status, response.Status)
	assert.Equal(t, supportRequest.AdminNotes, response.AdminNotes)
	assert.Equal(t, supportRequest.CreatedAt, response.CreatedAt)
	assert.Equal(t, supportRequest.UpdatedAt, response.UpdatedAt)
}

func TestSupportRequest_TableName(t *testing.T) {
	// Arrange
	sr := SupportRequest{}

	// Act
	tableName := sr.TableName()

	// Assert
	assert.Equal(t, "support_requests", tableName)
}

func TestSupportRequestType_Constants(t *testing.T) {
	assert.Equal(t, SupportRequestType("support"), SupportRequestTypeSupport)
	assert.Equal(t, SupportRequestType("feedback"), SupportRequestTypeFeedback)
}

func TestPlatform_Constants(t *testing.T) {
	assert.Equal(t, Platform("iOS"), PlatformIOS)
	assert.Equal(t, Platform("Android"), PlatformAndroid)
}

func TestStatus_Constants(t *testing.T) {
	assert.Equal(t, Status("new"), StatusNew)
	assert.Equal(t, Status("in_progress"), StatusInProgress)
	assert.Equal(t, Status("resolved"), StatusResolved)
}
