package services

import (
	"github.com/Edofo/bitly-clone/internal/models"
	"github.com/Edofo/bitly-clone/internal/repository"
)

type ClickService struct {
	clickRepo repository.ClickRepository
}

func NewClickService(clickRepo repository.ClickRepository) *ClickService {
	return &ClickService{
		clickRepo: clickRepo,
	}
}

func (s *ClickService) RecordClick(click *models.Click) error {
	return s.clickRepo.CreateClick(click)
}

func (s *ClickService) GetClicksCountByLinkID(linkID uint) (int, error) {
	return s.clickRepo.CountClicksByLinkID(linkID)
}
