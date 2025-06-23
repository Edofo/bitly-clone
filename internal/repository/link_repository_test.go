package repository

import (
	"testing"
	"time"

	"github.com/Edofo/bitly-clone/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	
	err = db.AutoMigrate(&models.Link{}, &models.Click{})
	assert.NoError(t, err)
	
	return db
}

func TestNewLinkRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLinkRepository(db)
	
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestGormLinkRepository_CreateLink(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLinkRepository(db)
	
	link := &models.Link{
		ShortCode: "abc123",
		LongURL:   "https://www.example.com",
		CreatedAt: time.Now(),
	}
	
	err := repo.CreateLink(link)
	
	assert.NoError(t, err)
	assert.NotZero(t, link.ID)
	
	var savedLink models.Link
	err = db.First(&savedLink, link.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, link.ShortCode, savedLink.ShortCode)
	assert.Equal(t, link.LongURL, savedLink.LongURL)
}

func TestGormLinkRepository_GetLinkByShortCode(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLinkRepository(db)
	
	link := &models.Link{
		ShortCode: "abc123",
		LongURL:   "https://www.example.com",
		CreatedAt: time.Now(),
	}
	err := repo.CreateLink(link)
	assert.NoError(t, err)
	
	retrievedLink, err := repo.GetLinkByShortCode("abc123")
	
	assert.NoError(t, err)
	assert.NotNil(t, retrievedLink)
	assert.Equal(t, link.ID, retrievedLink.ID)
	assert.Equal(t, link.ShortCode, retrievedLink.ShortCode)
	assert.Equal(t, link.LongURL, retrievedLink.LongURL)
}

func TestGormLinkRepository_GetLinkByShortCode_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLinkRepository(db)
	
	link, err := repo.GetLinkByShortCode("nonexistent")
	
	assert.Error(t, err)
	assert.Nil(t, link)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestGormLinkRepository_GetAllLinks(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLinkRepository(db)
	
	links := []*models.Link{
		{ShortCode: "abc123", LongURL: "https://www.example1.com", CreatedAt: time.Now()},
		{ShortCode: "def456", LongURL: "https://www.example2.com", CreatedAt: time.Now()},
		{ShortCode: "ghi789", LongURL: "https://www.example3.com", CreatedAt: time.Now()},
	}
	
	for _, link := range links {
		err := repo.CreateLink(link)
		assert.NoError(t, err)
	}
	
	allLinks, err := repo.GetAllLinks()
	
	assert.NoError(t, err)
	assert.Len(t, allLinks, 3)
	
	codes := make(map[string]bool)
	for _, link := range allLinks {
		codes[link.ShortCode] = true
	}
	
	assert.True(t, codes["abc123"])
	assert.True(t, codes["def456"])
	assert.True(t, codes["ghi789"])
}

func TestGormLinkRepository_CountClicksByLinkID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLinkRepository(db)
	
	link := &models.Link{
		ShortCode: "abc123",
		LongURL:   "https://www.example.com",
		CreatedAt: time.Now(),
	}
	err := repo.CreateLink(link)
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

func TestGormLinkRepository_CountClicksByLinkID_NoClicks(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLinkRepository(db)
	
	link := &models.Link{
		ShortCode: "abc123",
		LongURL:   "https://www.example.com",
		CreatedAt: time.Now(),
	}
	err := repo.CreateLink(link)
	assert.NoError(t, err)
	
	count, err := repo.CountClicksByLinkID(link.ID)
	
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
} 