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
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// UserInfo represents user information for responses
type UserInfo struct {
	ID       uint     `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Role     UserRole `json:"role"`
	IsActive bool     `json:"is_active"`
}

// CreateUserRequest represents the payload for creating a user
type CreateUserRequest struct {
	Username string   `json:"username" binding:"required,min=3,max=50"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=8"`
	Role     UserRole `json:"role" binding:"required,oneof=admin user"`
}

// UpdateUserRequest represents the payload for updating a user
type UpdateUserRequest struct {
	Email    *string   `json:"email,omitempty" binding:"omitempty,email"`
	Role     *UserRole `json:"role,omitempty" binding:"omitempty,oneof=admin user"`
	IsActive *bool     `json:"is_active,omitempty"`
}

// ChangePasswordRequest represents the payload for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
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
