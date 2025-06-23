package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLink_Struct(t *testing.T) {
	now := time.Now()
	link := Link{
		ID:        1,
		ShortCode: "abc123",
		LongURL:   "https://www.example.com",
		CreatedAt: now,
	}
	
	assert.Equal(t, uint(1), link.ID)
	assert.Equal(t, "abc123", link.ShortCode)
	assert.Equal(t, "https://www.example.com", link.LongURL)
	assert.Equal(t, now, link.CreatedAt)
}

func TestLink_EmptyStruct(t *testing.T) {
	link := Link{}
	
	assert.Equal(t, uint(0), link.ID)
	assert.Equal(t, "", link.ShortCode)
	assert.Equal(t, "", link.LongURL)
	assert.Equal(t, time.Time{}, link.CreatedAt)
} 