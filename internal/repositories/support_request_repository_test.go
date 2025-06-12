package repositories

import (
	"support-app-backend/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SupportRequestRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo SupportRequestRepository
}

func (suite *SupportRequestRepositoryTestSuite) SetupSuite() {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(postgres.Open("postgres://postgres:password@localhost:5432/test_db?sslmode=disable"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		suite.T().Skip("Skipping repository tests - PostgreSQL not available")
		return
	}

	suite.db = db
	suite.repo = NewSupportRequestRepository(db)

	// Auto migrate
	err = db.AutoMigrate(&models.SupportRequest{})
	suite.Require().NoError(err)
}

func (suite *SupportRequestRepositoryTestSuite) SetupTest() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}
	// Clean up before each test
	suite.db.Exec("DELETE FROM support_requests")
}

func (suite *SupportRequestRepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		sqlDB.Close()
	}
}

func (suite *SupportRequestRepositoryTestSuite) TestCreate() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Arrange
	userEmail := "test@example.com"
	request := &models.SupportRequest{
		Type:        models.SupportRequestTypeSupport,
		UserEmail:   &userEmail,
		Message:     "Test message",
		Platform:    models.PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
		Status:      models.StatusNew,
	}

	// Act
	err := suite.repo.Create(request)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), request.ID)
	assert.NotZero(suite.T(), request.CreatedAt)
	assert.NotZero(suite.T(), request.UpdatedAt)
}

func (suite *SupportRequestRepositoryTestSuite) TestGetByID() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Arrange
	userEmail := "test@example.com"
	originalRequest := &models.SupportRequest{
		Type:        models.SupportRequestTypeSupport,
		UserEmail:   &userEmail,
		Message:     "Test message",
		Platform:    models.PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
		Status:      models.StatusNew,
	}

	err := suite.repo.Create(originalRequest)
	suite.Require().NoError(err)

	// Act
	retrievedRequest, err := suite.repo.GetByID(originalRequest.ID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedRequest)
	assert.Equal(suite.T(), originalRequest.ID, retrievedRequest.ID)
	assert.Equal(suite.T(), originalRequest.Type, retrievedRequest.Type)
	assert.Equal(suite.T(), originalRequest.Message, retrievedRequest.Message)
}

func (suite *SupportRequestRepositoryTestSuite) TestGetByID_NotFound() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Act
	retrievedRequest, err := suite.repo.GetByID(999)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrievedRequest)
}

func (suite *SupportRequestRepositoryTestSuite) TestGetAll() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Arrange - Create multiple requests
	requests := []*models.SupportRequest{
		{
			Type:        models.SupportRequestTypeSupport,
			Message:     "Test message 1",
			Platform:    models.PlatformIOS,
			AppVersion:  "1.0.0",
			DeviceModel: "iPhone 13",
			Status:      models.StatusNew,
		},
		{
			Type:        models.SupportRequestTypeFeedback,
			Message:     "Test message 2",
			Platform:    models.PlatformAndroid,
			AppVersion:  "1.0.1",
			DeviceModel: "Samsung Galaxy",
			Status:      models.StatusNew,
		},
	}

	for _, req := range requests {
		err := suite.repo.Create(req)
		suite.Require().NoError(err)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Act
	retrievedRequests, total, err := suite.repo.GetAll(0, 10)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), total)
	assert.Len(suite.T(), retrievedRequests, 2)
	// Should be ordered by created_at DESC
	assert.Equal(suite.T(), "Test message 2", retrievedRequests[0].Message)
	assert.Equal(suite.T(), "Test message 1", retrievedRequests[1].Message)
}

func (suite *SupportRequestRepositoryTestSuite) TestUpdate() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Arrange
	userEmail := "test@example.com"
	request := &models.SupportRequest{
		Type:        models.SupportRequestTypeSupport,
		UserEmail:   &userEmail,
		Message:     "Test message",
		Platform:    models.PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
		Status:      models.StatusNew,
	}

	err := suite.repo.Create(request)
	suite.Require().NoError(err)

	// Act
	request.Status = models.StatusInProgress
	adminNotes := "Admin updated this"
	request.AdminNotes = &adminNotes
	err = suite.repo.Update(request)

	// Assert
	assert.NoError(suite.T(), err)

	// Verify update
	updatedRequest, err := suite.repo.GetByID(request.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.StatusInProgress, updatedRequest.Status)
	assert.Equal(suite.T(), "Admin updated this", *updatedRequest.AdminNotes)
}

func (suite *SupportRequestRepositoryTestSuite) TestDelete() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
		return
	}

	// Arrange
	userEmail := "test@example.com"
	request := &models.SupportRequest{
		Type:        models.SupportRequestTypeSupport,
		UserEmail:   &userEmail,
		Message:     "Test message",
		Platform:    models.PlatformIOS,
		AppVersion:  "1.0.0",
		DeviceModel: "iPhone 13",
		Status:      models.StatusNew,
	}

	err := suite.repo.Create(request)
	suite.Require().NoError(err)

	// Act
	err = suite.repo.Delete(request.ID)

	// Assert
	assert.NoError(suite.T(), err)

	// Verify deletion (soft delete)
	deletedRequest, err := suite.repo.GetByID(request.ID)
	assert.Error(suite.T(), err) // Should not be found due to soft delete
	assert.Nil(suite.T(), deletedRequest)
}

func TestSupportRequestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(SupportRequestRepositoryTestSuite))
}
