package repositories

import (
	"support-app-backend/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo UserRepository
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(suite.T(), err)

	// Auto migrate
	err = db.AutoMigrate(&models.User{})
	require.NoError(suite.T(), err)

	suite.db = db
	suite.repo = NewUserRepository(db)
}

func (suite *UserRepositoryTestSuite) TearDownSuite() {
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	// Clean up database before each test
	suite.db.Exec("DELETE FROM users")
}

func (suite *UserRepositoryTestSuite) TestCreate_Success() {
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	err := user.SetPassword("password123")
	require.NoError(suite.T(), err)

	err = suite.repo.Create(user)

	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), user.ID)
	assert.NotZero(suite.T(), user.CreatedAt)
	assert.NotZero(suite.T(), user.UpdatedAt)
}

func (suite *UserRepositoryTestSuite) TestCreate_DuplicateUsername() {
	user1 := &models.User{
		Username: "testuser",
		Email:    "test1@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	user1.SetPassword("password123")

	user2 := &models.User{
		Username: "testuser", // Same username
		Email:    "test2@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	user2.SetPassword("password123")

	err := suite.repo.Create(user1)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(user2)
	assert.Error(suite.T(), err)
}

func (suite *UserRepositoryTestSuite) TestCreate_DuplicateEmail() {
	user1 := &models.User{
		Username: "testuser1",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	user1.SetPassword("password123")

	user2 := &models.User{
		Username: "testuser2",
		Email:    "test@example.com", // Same email
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	user2.SetPassword("password123")

	err := suite.repo.Create(user1)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(user2)
	assert.Error(suite.T(), err)
}

func (suite *UserRepositoryTestSuite) TestGetByID_Success() {
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	user.SetPassword("password123")

	err := suite.repo.Create(user)
	require.NoError(suite.T(), err)

	foundUser, err := suite.repo.GetByID(user.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), foundUser)
	assert.Equal(suite.T(), user.ID, foundUser.ID)
	assert.Equal(suite.T(), user.Username, foundUser.Username)
	assert.Equal(suite.T(), user.Email, foundUser.Email)
	assert.Equal(suite.T(), user.Role, foundUser.Role)
	assert.Equal(suite.T(), user.IsActive, foundUser.IsActive)
}

func (suite *UserRepositoryTestSuite) TestGetByID_NotFound() {
	foundUser, err := suite.repo.GetByID(999)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundUser)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *UserRepositoryTestSuite) TestGetByUsername_Success() {
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	user.SetPassword("password123")

	err := suite.repo.Create(user)
	require.NoError(suite.T(), err)

	foundUser, err := suite.repo.GetByUsername("testuser")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), foundUser)
	assert.Equal(suite.T(), user.ID, foundUser.ID)
	assert.Equal(suite.T(), user.Username, foundUser.Username)
	assert.Equal(suite.T(), user.Email, foundUser.Email)
}

func (suite *UserRepositoryTestSuite) TestGetByUsername_NotFound() {
	foundUser, err := suite.repo.GetByUsername("nonexistent")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundUser)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *UserRepositoryTestSuite) TestGetByEmail_Success() {
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	user.SetPassword("password123")

	err := suite.repo.Create(user)
	require.NoError(suite.T(), err)

	foundUser, err := suite.repo.GetByEmail("test@example.com")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), foundUser)
	assert.Equal(suite.T(), user.ID, foundUser.ID)
	assert.Equal(suite.T(), user.Username, foundUser.Username)
	assert.Equal(suite.T(), user.Email, foundUser.Email)
}

func (suite *UserRepositoryTestSuite) TestGetByEmail_NotFound() {
	foundUser, err := suite.repo.GetByEmail("nonexistent@example.com")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundUser)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *UserRepositoryTestSuite) TestUpdate_Success() {
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	user.SetPassword("password123")

	err := suite.repo.Create(user)
	require.NoError(suite.T(), err)

	// Update user
	user.Email = "updated@example.com"
	user.Role = models.UserRoleAdmin
	user.IsActive = false

	err = suite.repo.Update(user)
	assert.NoError(suite.T(), err)

	// Verify update
	updatedUser, err := suite.repo.GetByID(user.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "updated@example.com", updatedUser.Email)
	assert.Equal(suite.T(), models.UserRoleAdmin, updatedUser.Role)
	assert.False(suite.T(), updatedUser.IsActive)
}

func (suite *UserRepositoryTestSuite) TestDelete_Success() {
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	user.SetPassword("password123")

	err := suite.repo.Create(user)
	require.NoError(suite.T(), err)

	err = suite.repo.Delete(user.ID)
	assert.NoError(suite.T(), err)

	// Verify deletion (soft delete)
	deletedUser, err := suite.repo.GetByID(user.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), deletedUser)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *UserRepositoryTestSuite) TestDelete_NotFound() {
	err := suite.repo.Delete(999)
	assert.NoError(suite.T(), err) // GORM doesn't error on deleting non-existent records
}

func (suite *UserRepositoryTestSuite) TestGetAll_Success() {
	// Create test users
	users := []*models.User{
		{Username: "user1", Email: "user1@example.com", Role: models.UserRoleUser, IsActive: true},
		{Username: "user2", Email: "user2@example.com", Role: models.UserRoleAdmin, IsActive: true},
		{Username: "user3", Email: "user3@example.com", Role: models.UserRoleUser, IsActive: false},
	}

	for _, user := range users {
		user.SetPassword("password123")
		err := suite.repo.Create(user)
		require.NoError(suite.T(), err)
	}

	// Test getting all users
	foundUsers, total, err := suite.repo.GetAll(0, 10)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), foundUsers, 3)
	assert.Equal(suite.T(), int64(3), total)
}

func (suite *UserRepositoryTestSuite) TestGetAll_WithPagination() {
	// Create test users
	users := []*models.User{
		{Username: "user1", Email: "user1@example.com", Role: models.UserRoleUser, IsActive: true},
		{Username: "user2", Email: "user2@example.com", Role: models.UserRoleAdmin, IsActive: true},
		{Username: "user3", Email: "user3@example.com", Role: models.UserRoleUser, IsActive: false},
	}

	for _, user := range users {
		user.SetPassword("password123")
		err := suite.repo.Create(user)
		require.NoError(suite.T(), err)
	}

	// Test pagination - get first 2 users
	foundUsers, total, err := suite.repo.GetAll(0, 2)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), foundUsers, 2)
	assert.Equal(suite.T(), int64(3), total)

	// Test pagination - get next user
	foundUsers, total, err = suite.repo.GetAll(2, 2)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), foundUsers, 1)
	assert.Equal(suite.T(), int64(3), total)
}

func (suite *UserRepositoryTestSuite) TestGetAll_Empty() {
	foundUsers, total, err := suite.repo.GetAll(0, 10)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), foundUsers, 0)
	assert.Equal(suite.T(), int64(0), total)
}

func (suite *UserRepositoryTestSuite) TestUpdateLastLogin_Success() {
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	user.SetPassword("password123")

	err := suite.repo.Create(user)
	require.NoError(suite.T(), err)

	// Initially last login should be nil
	assert.Nil(suite.T(), user.LastLoginAt)

	err = suite.repo.UpdateLastLogin(user.ID)
	assert.NoError(suite.T(), err)

	// Verify last login was updated
	updatedUser, err := suite.repo.GetByID(user.ID)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedUser.LastLoginAt)
	assert.True(suite.T(), updatedUser.LastLoginAt.After(time.Now().Add(-time.Minute)))
}

func (suite *UserRepositoryTestSuite) TestUpdateLastLogin_NotFound() {
	err := suite.repo.UpdateLastLogin(999)
	assert.NoError(suite.T(), err) // GORM doesn't error on updating non-existent records
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (suite *UserRepositoryTestSuite) TestUserExists_UserExists() {
	// Arrange
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
		IsActive: true,
	}
	user.SetPassword("password123")

	err := suite.repo.Create(user)
	require.NoError(suite.T(), err)

	// Act & Assert - Check by username
	exists, err := suite.repo.UserExists("testuser", "different@example.com")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)

	// Act & Assert - Check by email
	exists, err = suite.repo.UserExists("differentuser", "test@example.com")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)

	// Act & Assert - Check by both
	exists, err = suite.repo.UserExists("testuser", "test@example.com")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

func (suite *UserRepositoryTestSuite) TestUserExists_UserDoesNotExist() {
	// Act & Assert
	exists, err := suite.repo.UserExists("nonexistent", "nonexistent@example.com")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}
