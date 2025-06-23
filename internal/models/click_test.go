package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClick_Struct(t *testing.T) {
	now := time.Now()
	link := Link{
		ID:        1,
		ShortCode: "abc123",
		LongURL:   "https://www.example.com",
		CreatedAt: now,
	}
	
	click := Click{
		ID:        1,
		LinkID:    1,
		Link:      link,
		Timestamp: now,
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}
	
	assert.Equal(t, uint(1), click.ID)
	assert.Equal(t, uint(1), click.LinkID)
	assert.Equal(t, link, click.Link)
	assert.Equal(t, now, click.Timestamp)
	assert.Equal(t, "Mozilla/5.0", click.UserAgent)
	assert.Equal(t, "192.168.1.1", click.IPAddress)
}

func TestClickEvent_Struct(t *testing.T) {
	now := time.Now()
	clickEvent := ClickEvent{
		LinkID:    1,
		Timestamp: now,
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}
	
	assert.Equal(t, uint(1), clickEvent.LinkID)
	assert.Equal(t, now, clickEvent.Timestamp)
	assert.Equal(t, "Mozilla/5.0", clickEvent.UserAgent)
	assert.Equal(t, "192.168.1.1", clickEvent.IPAddress)
} 