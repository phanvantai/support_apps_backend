package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSupportRequest_ToResponse_CompleteRequest(t *testing.T) {
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
		App:         "my-awesome-app", // Add app field
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
	assert.Equal(t, supportRequest.App, response.App) // Test app field
	assert.Equal(t, supportRequest.CreatedAt, response.CreatedAt)
	assert.Equal(t, supportRequest.UpdatedAt, response.UpdatedAt)
}

func TestSupportRequest_ToResponse_MinimalRequest(t *testing.T) {
	// Arrange
	now := time.Now()
	supportRequest := &SupportRequest{
		ID:          2,
		Type:        SupportRequestTypeFeedback,
		Message:     "Test feedback",
		Platform:    PlatformAndroid,
		AppVersion:  "2.0.0",
		DeviceModel: "Samsung Galaxy",
		Status:      StatusResolved,
		App:         "another-app", // Add app field
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Act
	response := supportRequest.ToResponse()

	// Assert
	assert.Equal(t, supportRequest.ID, response.ID)
	assert.Equal(t, supportRequest.Type, response.Type)
	assert.Nil(t, response.UserEmail)
	assert.Equal(t, supportRequest.Message, response.Message)
	assert.Equal(t, supportRequest.Platform, response.Platform)
	assert.Equal(t, supportRequest.AppVersion, response.AppVersion)
	assert.Equal(t, supportRequest.DeviceModel, response.DeviceModel)
	assert.Equal(t, supportRequest.Status, response.Status)
	assert.Nil(t, response.AdminNotes)
	assert.Equal(t, supportRequest.App, response.App) // Test app field
	assert.Equal(t, supportRequest.CreatedAt, response.CreatedAt)
	assert.Equal(t, supportRequest.UpdatedAt, response.UpdatedAt)
}

func TestSupportRequest_TableName_ReturnsCorrectName(t *testing.T) {
	// Arrange
	sr := SupportRequest{}

	// Act
	tableName := sr.TableName()

	// Assert
	assert.Equal(t, "support_requests", tableName)
}

func TestSupportRequestType_Constants_ValidValues(t *testing.T) {
	assert.Equal(t, SupportRequestType("support"), SupportRequestTypeSupport)
	assert.Equal(t, SupportRequestType("feedback"), SupportRequestTypeFeedback)
}

func TestSupportRequestType_Constants_StringValues(t *testing.T) {
	assert.Equal(t, "support", string(SupportRequestTypeSupport))
	assert.Equal(t, "feedback", string(SupportRequestTypeFeedback))
}

func TestPlatform_Constants_ValidValues(t *testing.T) {
	assert.Equal(t, Platform("iOS"), PlatformIOS)
	assert.Equal(t, Platform("Android"), PlatformAndroid)
	assert.Equal(t, Platform("Web"), PlatformWeb)
}

func TestPlatform_Constants_StringValues(t *testing.T) {
	assert.Equal(t, "iOS", string(PlatformIOS))
	assert.Equal(t, "Android", string(PlatformAndroid))
	assert.Equal(t, "Web", string(PlatformWeb))
}

func TestStatus_Constants_ValidValues(t *testing.T) {
	assert.Equal(t, Status("new"), StatusNew)
	assert.Equal(t, Status("in_progress"), StatusInProgress)
	assert.Equal(t, Status("resolved"), StatusResolved)
}

func TestStatus_Constants_StringValues(t *testing.T) {
	assert.Equal(t, "new", string(StatusNew))
	assert.Equal(t, "in_progress", string(StatusInProgress))
	assert.Equal(t, "resolved", string(StatusResolved))
}

// Additional validation and edge case tests

func TestSupportRequest_ToResponse_NilPointers(t *testing.T) {
	// Arrange - test with nil optional fields
	supportRequest := &SupportRequest{
		ID:          3,
		Type:        SupportRequestTypeSupport,
		UserEmail:   nil, // nil email
		Message:     "Test message without email",
		Platform:    PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
		App:         "test-app-nil", // Add app field
		Status:      StatusNew,
		AdminNotes:  nil, // nil admin notes
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Act
	response := supportRequest.ToResponse()

	// Assert
	assert.Equal(t, supportRequest.ID, response.ID)
	assert.Equal(t, supportRequest.Type, response.Type)
	assert.Nil(t, response.UserEmail)
	assert.Equal(t, supportRequest.Message, response.Message)
	assert.Equal(t, supportRequest.Platform, response.Platform)
	assert.Equal(t, supportRequest.AppVersion, response.AppVersion)
	assert.Equal(t, supportRequest.DeviceModel, response.DeviceModel)
	assert.Equal(t, supportRequest.App, response.App) // Test app field
	assert.Equal(t, supportRequest.Status, response.Status)
	assert.Nil(t, response.AdminNotes)
	assert.Equal(t, supportRequest.CreatedAt, response.CreatedAt)
	assert.Equal(t, supportRequest.UpdatedAt, response.UpdatedAt)
}

func TestSupportRequest_ToResponse_ZeroValues(t *testing.T) {
	// Arrange - test with zero values
	supportRequest := &SupportRequest{}

	// Act
	response := supportRequest.ToResponse()

	// Assert
	assert.Equal(t, uint(0), response.ID)
	assert.Equal(t, SupportRequestType(""), response.Type)
	assert.Nil(t, response.UserEmail)
	assert.Equal(t, "", response.Message)
	assert.Equal(t, Platform(""), response.Platform)
	assert.Equal(t, "", response.AppVersion)
	assert.Equal(t, "", response.DeviceModel)
	assert.Equal(t, "", response.App) // Test app field for zero values
	assert.Equal(t, Status(""), response.Status)
	assert.Nil(t, response.AdminNotes)
	assert.True(t, response.CreatedAt.IsZero())
	assert.True(t, response.UpdatedAt.IsZero())
}

func TestSupportRequestType_TypeSafety(t *testing.T) {
	// Test that SupportRequestType is properly typed
	var reqType SupportRequestType = SupportRequestTypeSupport
	assert.Equal(t, "support", string(reqType))

	reqType = SupportRequestTypeFeedback
	assert.Equal(t, "feedback", string(reqType))

	// Test invalid type
	invalidType := SupportRequestType("invalid")
	assert.Equal(t, "invalid", string(invalidType))
}

func TestPlatform_TypeSafety(t *testing.T) {
	// Test that Platform is properly typed
	var platform Platform = PlatformIOS
	assert.Equal(t, "iOS", string(platform))

	platform = PlatformAndroid
	assert.Equal(t, "Android", string(platform))

	// Test invalid platform
	invalidPlatform := Platform("Windows")
	assert.Equal(t, "Windows", string(invalidPlatform))
}

func TestStatus_TypeSafety(t *testing.T) {
	// Test that Status is properly typed
	var status Status = StatusNew
	assert.Equal(t, "new", string(status))

	status = StatusInProgress
	assert.Equal(t, "in_progress", string(status))

	status = StatusResolved
	assert.Equal(t, "resolved", string(status))

	// Test invalid status
	invalidStatus := Status("cancelled")
	assert.Equal(t, "cancelled", string(invalidStatus))
}

func TestSupportRequest_ToResponse_AppField(t *testing.T) {
	// Arrange
	supportRequest := &SupportRequest{
		ID:          1,
		Type:        SupportRequestTypeSupport,
		Message:     "Test message",
		Platform:    PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
		Status:      StatusNew,
		App:         "test-app",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Act
	response := supportRequest.ToResponse()

	// Assert
	assert.Equal(t, "test-app", response.App)
	assert.Equal(t, supportRequest.App, response.App)
}

func TestSupportRequest_ToResponse_EmptyAppField(t *testing.T) {
	// Arrange
	supportRequest := &SupportRequest{
		ID:          1,
		Type:        SupportRequestTypeSupport,
		Message:     "Test message",
		Platform:    PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
		Status:      StatusNew,
		App:         "", // Empty app
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Act
	response := supportRequest.ToResponse()

	// Assert
	assert.Equal(t, "", response.App)
	assert.Equal(t, supportRequest.App, response.App)
}
