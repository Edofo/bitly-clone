package services

import (
	"testing"
	"time"

	"github.com/Edofo/bitly-clone/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClickRepository est un mock du ClickRepository pour les tests
type MockClickRepository struct {
	mock.Mock
}

func (m *MockClickRepository) CreateClick(click *models.Click) error {
	args := m.Called(click)
	return args.Error(0)
}

func (m *MockClickRepository) CountClicksByLinkID(linkID uint) (int, error) {
	args := m.Called(linkID)
	return args.Int(0), args.Error(1)
}

func TestNewClickService(t *testing.T) {
	mockRepo := &MockClickRepository{}
	service := NewClickService(mockRepo)
	
	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.clickRepo)
}

func TestRecordClick_Success(t *testing.T) {
	mockRepo := &MockClickRepository{}
	service := NewClickService(mockRepo)
	
	click := &models.Click{
		LinkID:    1,
		Timestamp: time.Now(),
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}
	
	mockRepo.On("CreateClick", click).Return(nil)
	
	err := service.RecordClick(click)
	
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestRecordClick_Error(t *testing.T) {
	mockRepo := &MockClickRepository{}
	service := NewClickService(mockRepo)
	
	click := &models.Click{
		LinkID:    1,
		Timestamp: time.Now(),
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}
	
	mockRepo.On("CreateClick", click).Return(assert.AnError)
	
	err := service.RecordClick(click)
	
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetClicksCountByLinkID_Success(t *testing.T) {
	mockRepo := &MockClickRepository{}
	service := NewClickService(mockRepo)
	
	linkID := uint(1)
	expectedCount := 42
	
	mockRepo.On("CountClicksByLinkID", linkID).Return(expectedCount, nil)
	
	count, err := service.GetClicksCountByLinkID(linkID)
	
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	mockRepo.AssertExpectations(t)
}

func TestGetClicksCountByLinkID_Error(t *testing.T) {
	mockRepo := &MockClickRepository{}
	service := NewClickService(mockRepo)
	
	linkID := uint(1)
	
	mockRepo.On("CountClicksByLinkID", linkID).Return(0, assert.AnError)
	
	count, err := service.GetClicksCountByLinkID(linkID)
	
	assert.Error(t, err)
	assert.Equal(t, 0, count)
	mockRepo.AssertExpectations(t)
} 