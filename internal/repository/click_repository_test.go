package repository

import (
	"testing"
	"time"

	"github.com/Edofo/bitly-clone/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupClickTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	
	err = db.AutoMigrate(&models.Link{}, &models.Click{})
	assert.NoError(t, err)
	
	return db
}

func TestNewClickRepository(t *testing.T) {
	db := setupClickTestDB(t)
	repo := NewClickRepository(db)
	
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestGormClickRepository_CreateClick(t *testing.T) {
	db := setupClickTestDB(t)
	repo := NewClickRepository(db)
	
	link := &models.Link{
		ShortCode: "abc123",
		LongURL:   "https://www.example.com",
		CreatedAt: time.Now(),
	}
	err := db.Create(link).Error
	assert.NoError(t, err)
	
	click := &models.Click{
		LinkID:    link.ID,
		Timestamp: time.Now(),
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}
	
	err = repo.CreateClick(click)
	
	assert.NoError(t, err)
	assert.NotZero(t, click.ID)
	
	var savedClick models.Click
	err = db.First(&savedClick, click.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, click.LinkID, savedClick.LinkID)
	assert.Equal(t, click.UserAgent, savedClick.UserAgent)
	assert.Equal(t, click.IPAddress, savedClick.IPAddress)
}

func TestGormClickRepository_CountClicksByLinkID(t *testing.T) {
	db := setupClickTestDB(t)
	repo := NewClickRepository(db)
	
	link := &models.Link{
		ShortCode: "abc123",
		LongURL:   "https://www.example.com",
		CreatedAt: time.Now(),
	}
	err := db.Create(link).Error
	assert.NoError(t, err)
	
	clicks := []models.Click{
		{LinkID: link.ID, Timestamp: time.Now(), UserAgent: "Mozilla/5.0", IPAddress: "192.168.1.1"},
		{LinkID: link.ID, Timestamp: time.Now(), UserAgent: "Chrome/90.0", IPAddress: "192.168.1.2"},
		{LinkID: link.ID, Timestamp: time.Now(), UserAgent: "Safari/14.0", IPAddress: "192.168.1.3"},
	}
	
	for _, click := range clicks {
		err := db.Create(&click).Error
		assert.NoError(t, err)
	}
	
	count, err := repo.CountClicksByLinkID(link.ID)
	
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestGormClickRepository_CountClicksByLinkID_NoClicks(t *testing.T) {
	db := setupClickTestDB(t)
	repo := NewClickRepository(db)
	
	link := &models.Link{
		ShortCode: "abc123",
		LongURL:   "https://www.example.com",
		CreatedAt: time.Now(),
	}
	err := db.Create(link).Error
	assert.NoError(t, err)
	
	count, err := repo.CountClicksByLinkID(link.ID)
	
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
} 