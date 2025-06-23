package monitor

import (
	"log"
	"net/http"
	"sync"
	"time"

	_ "github.com/Edofo/bitly-clone/internal/models"
	"github.com/Edofo/bitly-clone/internal/repository"
)

type UrlMonitor struct {
	linkRepo    repository.LinkRepository
	interval    time.Duration
	knownStates map[uint]bool
	mu          sync.Mutex
}

func NewUrlMonitor(linkRepo repository.LinkRepository, interval time.Duration) *UrlMonitor {
	return &UrlMonitor{
		linkRepo:    linkRepo,
		interval:    interval,
		knownStates: make(map[uint]bool),
	}
}

func (m *UrlMonitor) Start() {
	log.Printf("[MONITOR] Starting URL monitor with interval %v...", m.interval)
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	m.checkUrls()

	for range ticker.C {
		m.checkUrls()
	}
}

func (m *UrlMonitor) checkUrls() {
	log.Println("[MONITOR] Starting URL status check...")

	links, err := m.linkRepo.GetAllLinks()
	if err != nil {
		log.Printf("[MONITOR] ERROR retrieving links for monitoring: %v", err)
		return
	}

	for _, link := range links {
		currentState := m.isUrlAccessible(link.LongURL)

		m.mu.Lock()
		previousState, exists := m.knownStates[link.ID]
		m.knownStates[link.ID] = currentState
		m.mu.Unlock()

		if !exists {
			log.Printf("[MONITOR] Initial state for link %s (%s): %s",
				link.ShortCode, link.LongURL, formatState(currentState))
			continue
		}

		if previousState != currentState {
			log.Printf("[NOTIFICATION] Link %s (%s) changed from %s to %s!",
				link.ShortCode, link.LongURL, formatState(previousState), formatState(currentState))
		}
	}
	log.Println("[MONITOR] URL status check completed.")
}

func (m *UrlMonitor) isUrlAccessible(url string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Head(url)
	if err != nil {
		log.Printf("[MONITOR] Error accessing URL '%s': %v", url, err)
		return false
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("[MONITOR] Warning: Failed to close response body: %v", err)
		}
	}()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func formatState(accessible bool) string {
	if accessible {
		return "ACCESSIBLE"
	}
	return "INACCESSIBLE"
}
