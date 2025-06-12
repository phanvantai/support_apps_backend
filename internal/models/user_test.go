package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_SetPassword(t *testing.T) {
	// Arrange
	user := &User{}
	password := "testPassword123"

	// Act
	err := user.SetPassword(password)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, password, user.PasswordHash) // Should be hashed
}

func TestUser_CheckPassword(t *testing.T) {
	// Arrange
	user := &User{}
	password := "testPassword123"
	user.SetPassword(password)

	// Act & Assert
	assert.True(t, user.CheckPassword(password))
	assert.False(t, user.CheckPassword("wrongPassword"))
}

func TestUser_ToUserInfo(t *testing.T) {
	// Arrange
	user := &User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         UserRoleAdmin,
		IsActive:     true,
	}

	// Act
	userInfo := user.ToUserInfo()

	// Assert
	assert.Equal(t, user.ID, userInfo.ID)
	assert.Equal(t, user.Username, userInfo.Username)
	assert.Equal(t, user.Email, userInfo.Email)
	assert.Equal(t, user.Role, userInfo.Role)
	assert.Equal(t, user.IsActive, userInfo.IsActive)
}

func TestUser_TableName(t *testing.T) {
	// Arrange
	user := User{}

	// Act
	tableName := user.TableName()

	// Assert
	assert.Equal(t, "users", tableName)
}

func TestUserRole_Constants(t *testing.T) {
	assert.Equal(t, UserRole("admin"), UserRoleAdmin)
	assert.Equal(t, UserRole("user"), UserRoleUser)
}
