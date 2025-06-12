package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

// User represents a system user
type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Username     string         `json:"username" gorm:"not null;size:50;unique"`
	Email        string         `json:"email" gorm:"not null;size:255;unique"`
	PasswordHash string         `json:"-" gorm:"not null"`
	Role         UserRole       `json:"role" gorm:"not null;default:user"`
	IsActive     bool           `json:"is_active" gorm:"not null;default:true"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// LoginRequest represents the login request payload
// @Description User login request
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`              // Username for login
	Password string `json:"password" binding:"required" example:"securePassword@123"` // Password for login
}

// LoginResponse represents the login response
// @Description User login response with JWT token
type LoginResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT access token
	ExpiresAt time.Time `json:"expires_at" example:"2023-12-31T23:59:59Z"`               // Token expiration time
	User      UserInfo  `json:"user"`                                                    // User information
}

// UserInfo represents user information for responses
// @Description User information response
type UserInfo struct {
	ID       uint     `json:"id" example:"1"`                    // User ID
	Username string   `json:"username" example:"admin"`          // Username
	Email    string   `json:"email" example:"admin@example.com"` // User email
	Role     UserRole `json:"role" example:"admin"`              // User role (admin/user)
	IsActive bool     `json:"is_active" example:"true"`          // Whether user is active
}

// CreateUserRequest represents the payload for creating a user
// @Description Request payload for creating a new user
type CreateUserRequest struct {
	Username string   `json:"username" binding:"required,min=3,max=50" example:"newuser"`     // Username (3-50 characters)
	Email    string   `json:"email" binding:"required,email" example:"newuser@example.com"`   // Valid email address
	Password string   `json:"password" binding:"required,min=8" example:"securePassword@123"` // Password (min 8 characters)
	Role     UserRole `json:"role" binding:"required,oneof=admin user" example:"user"`        // User role (admin or user)
}

// UpdateUserRequest represents the payload for updating a user
// @Description Request payload for updating user information
type UpdateUserRequest struct {
	Email    *string   `json:"email,omitempty" binding:"omitempty,email" example:"updated@example.com"` // New email address
	Role     *UserRole `json:"role,omitempty" binding:"omitempty,oneof=admin user" example:"admin"`     // New user role
	IsActive *bool     `json:"is_active,omitempty" example:"false"`                                     // Active status
}

// ChangePasswordRequest represents the payload for changing password
// @Description Request payload for changing user password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"oldPassword123"`   // Current password
	NewPassword     string `json:"new_password" binding:"required,min=8" example:"newPassword123"` // New password (min 8 characters)
}

// TableName returns the table name for GORM
func (User) TableName() string {
	return "users"
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword verifies the provided password against the user's password hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// ToUserInfo converts User to UserInfo
func (u *User) ToUserInfo() UserInfo {
	return UserInfo{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Role:     u.Role,
		IsActive: u.IsActive,
	}
}
