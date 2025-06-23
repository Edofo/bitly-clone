package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"gorm.io/gorm"

	"github.com/Edofo/bitly-clone/internal/models"
	"github.com/Edofo/bitly-clone/internal/repository"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type LinkService struct {
	linkRepo repository.LinkRepository
}

func NewLinkService(linkRepo repository.LinkRepository) *LinkService {
	return &LinkService{
		linkRepo: linkRepo,
	}
}

func (s *LinkService) GenerateShortCode(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("error generating random index: %w", err)
		}
		result[i] = charset[randomIndex.Int64()]
	}
	return string(result), nil
}

func (s *LinkService) CreateLink(longURL string) (*models.Link, error) {
	var shortCode string
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		code, err := s.GenerateShortCode(6)
		if err != nil {
			return nil, fmt.Errorf("error generating short code: %w", err)
		}

		_, err = s.linkRepo.GetLinkByShortCode(code)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				shortCode = code
				break
			}
			return nil, fmt.Errorf("database error checking short code uniqueness: %w", err)
		}

		log.Printf("Short code '%s' already exists, retrying generation (%d/%d)...", code, i+1, maxRetries)
	}

	if shortCode == "" {
		return nil, errors.New("failed to generate unique short code after maximum retries")
	}

	link := &models.Link{
		ShortCode: shortCode,
		LongURL:   longURL,
		CreatedAt: time.Now(),
	}

	if err := s.linkRepo.CreateLink(link); err != nil {
		return nil, fmt.Errorf("error creating link: %w", err)
	}

	return link, nil
}

func (s *LinkService) GetLinkByShortCode(shortCode string) (*models.Link, error) {
	return s.linkRepo.GetLinkByShortCode(shortCode)
}

func (s *LinkService) GetLinkStats(shortCode string) (*models.Link, int, error) {
	link, err := s.linkRepo.GetLinkByShortCode(shortCode)
	if err != nil {
		return nil, 0, err
	}

	totalClicks, err := s.linkRepo.CountClicksByLinkID(link.ID)
	if err != nil {
		return nil, 0, err
	}

	return link, totalClicks, nil
}
