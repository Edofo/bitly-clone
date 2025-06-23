package repository

import (
	"github.com/Edofo/bitly-clone/internal/models"
	"gorm.io/gorm"
)

type LinkRepository interface {
	CreateLink(link *models.Link) error
	GetLinkByShortCode(shortCode string) (*models.Link, error)
	GetAllLinks() ([]models.Link, error)
	CountClicksByLinkID(linkID uint) (int, error)
}

type GormLinkRepository struct {
	db *gorm.DB
}

func NewLinkRepository(db *gorm.DB) *GormLinkRepository {
	return &GormLinkRepository{db: db}
}

func (r *GormLinkRepository) CreateLink(link *models.Link) error {
	return r.db.Create(link).Error
}

func (r *GormLinkRepository) GetLinkByShortCode(shortCode string) (*models.Link, error) {
	var link models.Link
	err := r.db.Where("short_code = ?", shortCode).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *GormLinkRepository) GetAllLinks() ([]models.Link, error) {
	var links []models.Link
	err := r.db.Find(&links).Error
	return links, err
}

func (r *GormLinkRepository) CountClicksByLinkID(linkID uint) (int, error) {
	var count int64
	err := r.db.Model(&models.Click{}).Where("link_id = ?", linkID).Count(&count).Error
	return int(count), err
}
