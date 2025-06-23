package repository

import (
	"github.com/Edofo/bitly-clone/internal/models"
	"gorm.io/gorm"
)

type ClickRepository interface {
	CreateClick(click *models.Click) error
	CountClicksByLinkID(linkID uint) (int, error)
}

type GormClickRepository struct {
	db *gorm.DB
}

func NewClickRepository(db *gorm.DB) *GormClickRepository {
	return &GormClickRepository{db: db}
}

func (r *GormClickRepository) CreateClick(click *models.Click) error {
	return r.db.Create(click).Error
}

func (r *GormClickRepository) CountClicksByLinkID(linkID uint) (int, error) {
	var count int64
	err := r.db.Model(&models.Click{}).Where("link_id = ?", linkID).Count(&count).Error
	return int(count), err
}
