package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_WithValidConfigFile(t *testing.T) {
	tempConfig := `server:
  port: 9090
  base_url: "http://localhost:9090"
database:
  name: "test.db"
analytics:
  buffer_size: 500
  workers: 3
monitor:
  interval_minutes: 10`

	err := os.WriteFile("configs/config.yaml", []byte(tempConfig), 0644)
	assert.NoError(t, err)
	defer os.Remove("configs/config.yaml")
	
	cfg, err := LoadConfig()
	
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "http://localhost:9090", cfg.Server.BaseURL)
	assert.Equal(t, "test.db", cfg.Database.Name)
	assert.Equal(t, 500, cfg.Analytics.BufferSize)
	assert.Equal(t, 3, cfg.Analytics.Workers)
	assert.Equal(t, 10, cfg.Monitor.IntervalMinutes)
}

func TestLoadConfig_WithDefaultValues(t *testing.T) {
	os.Remove("configs/config.yaml")
	
	cfg, err := LoadConfig()
	
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "http://localhost:8080", cfg.Server.BaseURL)
	assert.Equal(t, "url_shortener.db", cfg.Database.Name)
	assert.Equal(t, 1000, cfg.Analytics.BufferSize)
	assert.Equal(t, 5, cfg.Analytics.Workers)
	assert.Equal(t, 5, cfg.Monitor.IntervalMinutes)
}

func TestLoadConfig_WithInvalidConfigFile(t *testing.T) {
	invalidConfig := `server:
  port: "not_a_number"`

	err := os.WriteFile("configs/config.yaml", []byte(invalidConfig), 0644)
	assert.NoError(t, err)
	defer os.Remove("configs/config.yaml")
	
	cfg, err := LoadConfig()
	
	assert.Error(t, err)
	assert.Nil(t, cfg)
} 