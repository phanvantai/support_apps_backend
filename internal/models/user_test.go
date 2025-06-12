package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_SetPassword_Success(t *testing.T) {
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

func TestUser_SetPassword_EmptyPassword(t *testing.T) {
	// Arrange
	user := &User{}
	password := ""

	// Act
	err := user.SetPassword(password)

	// Assert
	assert.NoError(t, err) // bcrypt handles empty passwords
	assert.NotEmpty(t, user.PasswordHash)
}

func TestUser_CheckPassword_ValidPassword(t *testing.T) {
	// Arrange
	user := &User{}
	password := "testPassword123"
	user.SetPassword(password)

	// Act & Assert
	assert.True(t, user.CheckPassword(password))
}

func TestUser_CheckPassword_InvalidPassword(t *testing.T) {
	// Arrange
	user := &User{}
	password := "testPassword123"
	user.SetPassword(password)

	// Act & Assert
	assert.False(t, user.CheckPassword("wrongPassword"))
}

func TestUser_CheckPassword_EmptyPassword(t *testing.T) {
	// Arrange
	user := &User{}
	password := "testPassword123"
	user.SetPassword(password)

	// Act & Assert
	assert.False(t, user.CheckPassword(""))
}

func TestUser_CheckPassword_NoPasswordSet(t *testing.T) {
	// Arrange
	user := &User{} // No password set

	// Act & Assert
	assert.False(t, user.CheckPassword("anyPassword"))
}

func TestUser_ToUserInfo_Success(t *testing.T) {
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

func TestUser_ToUserInfo_InactiveUser(t *testing.T) {
	// Arrange
	user := &User{
		ID:           2,
		Username:     "inactiveuser",
		Email:        "inactive@example.com",
		PasswordHash: "hashed_password",
		Role:         UserRoleUser,
		IsActive:     false,
	}

	// Act
	userInfo := user.ToUserInfo()

	// Assert
	assert.Equal(t, user.ID, userInfo.ID)
	assert.Equal(t, user.Username, userInfo.Username)
	assert.Equal(t, user.Email, userInfo.Email)
	assert.Equal(t, user.Role, userInfo.Role)
	assert.False(t, userInfo.IsActive)
}

func TestUser_TableName_ReturnsCorrectName(t *testing.T) {
	// Arrange
	user := User{}

	// Act
	tableName := user.TableName()

	// Assert
	assert.Equal(t, "users", tableName)
}

func TestUserRole_Constants_ValidValues(t *testing.T) {
	assert.Equal(t, UserRole("admin"), UserRoleAdmin)
	assert.Equal(t, UserRole("user"), UserRoleUser)
}

func TestUserRole_Constants_StringValues(t *testing.T) {
	assert.Equal(t, "admin", string(UserRoleAdmin))
	assert.Equal(t, "user", string(UserRoleUser))
}

// Additional validation and edge case tests

func TestUser_SetPassword_BcryptError(t *testing.T) {
	// This test verifies error handling in SetPassword
	// In normal operation, bcrypt.GenerateFromPassword rarely fails
	// but we test the error path exists
	user := &User{}

	// Test with a very long password that might cause issues
	longPassword := string(make([]byte, 100000)) // Very long password
	err := user.SetPassword(longPassword)

	// bcrypt should handle this gracefully or return an error
	// The key is that we don't panic
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.NotEmpty(t, user.PasswordHash)
	}
}

func TestUser_CheckPassword_EmptyHash(t *testing.T) {
	// Arrange
	user := &User{} // No password hash set

	// Act & Assert
	assert.False(t, user.CheckPassword("anyPassword"))
	assert.False(t, user.CheckPassword(""))
}

func TestUser_CheckPassword_CorruptedHash(t *testing.T) {
	// Arrange
	user := &User{
		PasswordHash: "not-a-valid-bcrypt-hash",
	}

	// Act & Assert
	assert.False(t, user.CheckPassword("anyPassword"))
}

func TestUser_ToUserInfo_ZeroValues(t *testing.T) {
	// Arrange
	user := &User{} // All zero values

	// Act
	userInfo := user.ToUserInfo()

	// Assert
	assert.Equal(t, uint(0), userInfo.ID)
	assert.Equal(t, "", userInfo.Username)
	assert.Equal(t, "", userInfo.Email)
	assert.Equal(t, UserRole(""), userInfo.Role)
	assert.False(t, userInfo.IsActive)
}

func TestUserRole_TypeSafety(t *testing.T) {
	// Test that UserRole is properly typed
	var role UserRole = UserRoleAdmin
	assert.Equal(t, "admin", string(role))

	role = UserRoleUser
	assert.Equal(t, "user", string(role))

	// Test invalid role
	invalidRole := UserRole("invalid")
	assert.Equal(t, "invalid", string(invalidRole))
}
