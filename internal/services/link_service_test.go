package services

import (
	"errors"
	"testing"
	"time"

	"github.com/Edofo/bitly-clone/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockLinkRepository struct {
	mock.Mock
}

func (m *MockLinkRepository) CreateLink(link *models.Link) error {
	args := m.Called(link)
	return args.Error(0)
}

func (m *MockLinkRepository) GetLinkByShortCode(shortCode string) (*models.Link, error) {
	args := m.Called(shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Link), args.Error(1)
}

func (m *MockLinkRepository) GetAllLinks() ([]models.Link, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Link), args.Error(1)
}

func (m *MockLinkRepository) CountClicksByLinkID(linkID uint) (int, error) {
	args := m.Called(linkID)
	return args.Int(0), args.Error(1)
}

func TestNewLinkService(t *testing.T) {
	mockRepo := &MockLinkRepository{}
	service := NewLinkService(mockRepo)
	
	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.linkRepo)
}

func TestGenerateShortCode(t *testing.T) {
	service := &LinkService{}
	
	// Test génération d'un code de 6 caractères
	code, err := service.GenerateShortCode(6)
	assert.NoError(t, err)
	assert.Len(t, code, 6)
	
	// Test génération d'un code de 10 caractères
	code, err = service.GenerateShortCode(10)
	assert.NoError(t, err)
	assert.Len(t, code, 10)
	
	// Test que le code ne contient que des caractères alphanumériques
	for _, char := range code {
		assert.Contains(t, charset, string(char))
	}
}

func TestGenerateShortCode_InvalidLength(t *testing.T) {
	service := &LinkService{}
	
	// Test avec une longueur négative
	_, err := service.GenerateShortCode(-1)
	assert.Error(t, err)
}

func TestCreateLink_Success(t *testing.T) {
	mockRepo := &MockLinkRepository{}
	service := NewLinkService(mockRepo)
	
	longURL := "https://www.example.com"
	
	// Mock pour GetLinkByShortCode - retourne ErrRecordNotFound (code unique)
	mockRepo.On("GetLinkByShortCode", mock.AnythingOfType("string")).Return(nil, gorm.ErrRecordNotFound)
	
	// Mock pour CreateLink - succès
	mockRepo.On("CreateLink", mock.AnythingOfType("*models.Link")).Return(nil)
	
	link, err := service.CreateLink(longURL)
	
	assert.NoError(t, err)
	assert.NotNil(t, link)
	assert.Equal(t, longURL, link.LongURL)
	assert.Len(t, link.ShortCode, 6)
	assert.WithinDuration(t, time.Now(), link.CreatedAt, 2*time.Second)
	
	mockRepo.AssertExpectations(t)
}

func TestCreateLink_RetryOnCollision(t *testing.T) {
	mockRepo := &MockLinkRepository{}
	service := NewLinkService(mockRepo)
	
	longURL := "https://www.example.com"
	
	// Mock pour GetLinkByShortCode - collision les 2 premières fois, puis succès
	mockRepo.On("GetLinkByShortCode", mock.AnythingOfType("string")).Return(
		&models.Link{}, nil, // Première collision
	).Once()
	mockRepo.On("GetLinkByShortCode", mock.AnythingOfType("string")).Return(
		&models.Link{}, nil, // Deuxième collision
	).Once()
	mockRepo.On("GetLinkByShortCode", mock.AnythingOfType("string")).Return(
		nil, gorm.ErrRecordNotFound, // Code unique trouvé
	).Once()
	
	// Mock pour CreateLink - succès
	mockRepo.On("CreateLink", mock.AnythingOfType("*models.Link")).Return(nil)
	
	link, err := service.CreateLink(longURL)
	
	assert.NoError(t, err)
	assert.NotNil(t, link)
	assert.Equal(t, longURL, link.LongURL)
	
	mockRepo.AssertExpectations(t)
}

func TestCreateLink_MaxRetriesExceeded(t *testing.T) {
	mockRepo := &MockLinkRepository{}
	service := NewLinkService(mockRepo)
	
	longURL := "https://www.example.com"
	
	// Mock pour GetLinkByShortCode - toujours des collisions
	mockRepo.On("GetLinkByShortCode", mock.AnythingOfType("string")).Return(
		&models.Link{}, nil, // Toujours des collisions
	).Times(5)
	
	link, err := service.CreateLink(longURL)
	
	assert.Error(t, err)
	assert.Nil(t, link)
	assert.Contains(t, err.Error(), "failed to generate unique short code after maximum retries")
	
	mockRepo.AssertExpectations(t)
}

func TestCreateLink_DatabaseError(t *testing.T) {
	mockRepo := &MockLinkRepository{}
	service := NewLinkService(mockRepo)
	
	longURL := "https://www.example.com"
	
	// Mock pour GetLinkByShortCode - erreur de base de données
	mockRepo.On("GetLinkByShortCode", mock.AnythingOfType("string")).Return(
		nil, errors.New("database connection error"),
	)
	
	link, err := service.CreateLink(longURL)
	
	assert.Error(t, err)
	assert.Nil(t, link)
	assert.Contains(t, err.Error(), "database error checking short code uniqueness")
	
	mockRepo.AssertExpectations(t)
}

func TestCreateLink_CreateLinkError(t *testing.T) {
	mockRepo := &MockLinkRepository{}
	service := NewLinkService(mockRepo)
	
	longURL := "https://www.example.com"
	
	// Mock pour GetLinkByShortCode - code unique
	mockRepo.On("GetLinkByShortCode", mock.AnythingOfType("string")).Return(
		nil, gorm.ErrRecordNotFound,
	)
	
	// Mock pour CreateLink - erreur
	mockRepo.On("CreateLink", mock.AnythingOfType("*models.Link")).Return(
		errors.New("create link error"),
	)
	
	link, err := service.CreateLink(longURL)
	
	assert.Error(t, err)
	assert.Nil(t, link)
	assert.Contains(t, err.Error(), "error creating link")
	
	mockRepo.AssertExpectations(t)
}

func TestGetLinkByShortCode(t *testing.T) {
	mockRepo := &MockLinkRepository{}
	service := NewLinkService(mockRepo)
	
	shortCode := "abc123"
	expectedLink := &models.Link{
		ID:        1,
		ShortCode: shortCode,
		LongURL:   "https://www.example.com",
		CreatedAt: time.Now(),
	}
	
	mockRepo.On("GetLinkByShortCode", shortCode).Return(expectedLink, nil)
	
	link, err := service.GetLinkByShortCode(shortCode)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedLink, link)
	
	mockRepo.AssertExpectations(t)
}

func TestGetLinkStats(t *testing.T) {
	mockRepo := &MockLinkRepository{}
	service := NewLinkService(mockRepo)
	
	shortCode := "abc123"
	expectedLink := &models.Link{
		ID:        1,
		ShortCode: shortCode,
		LongURL:   "https://www.example.com",
		CreatedAt: time.Now(),
	}
	expectedClicks := 42
	
	mockRepo.On("GetLinkByShortCode", shortCode).Return(expectedLink, nil)
	mockRepo.On("CountClicksByLinkID", uint(1)).Return(expectedClicks, nil)
	
	link, clicks, err := service.GetLinkStats(shortCode)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedLink, link)
	assert.Equal(t, expectedClicks, clicks)
	
	mockRepo.AssertExpectations(t)
}

func TestGetLinkStats_LinkNotFound(t *testing.T) {
	mockRepo := &MockLinkRepository{}
	service := NewLinkService(mockRepo)
	
	shortCode := "abc123"
	
	mockRepo.On("GetLinkByShortCode", shortCode).Return(nil, gorm.ErrRecordNotFound)
	
	link, clicks, err := service.GetLinkStats(shortCode)
	
	assert.Error(t, err)
	assert.Nil(t, link)
	assert.Equal(t, 0, clicks)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	
	mockRepo.AssertExpectations(t)
} 