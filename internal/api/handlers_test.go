package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Edofo/bitly-clone/cmd"
	"github.com/Edofo/bitly-clone/internal/config"
	"github.com/Edofo/bitly-clone/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockLinkService struct {
	mock.Mock
}

func (m *MockLinkService) GenerateShortCode(length int) (string, error) {
	args := m.Called(length)
	return args.String(0), args.Error(1)
}

func (m *MockLinkService) CreateLink(longURL string) (*models.Link, error) {
	args := m.Called(longURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Link), args.Error(1)
}

func (m *MockLinkService) GetLinkByShortCode(shortCode string) (*models.Link, error) {
	args := m.Called(shortCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Link), args.Error(1)
}

func (m *MockLinkService) GetLinkStats(shortCode string) (*models.Link, int, error) {
	args := m.Called(shortCode)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).(*models.Link), args.Int(1), args.Error(2)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func setupTestConfig() {
}

func TestHealthCheckHandler(t *testing.T) {
	router := setupTestRouter()
	router.GET("/health", HealthCheckHandler)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestCreateShortLinkHandler_Success(t *testing.T) {
	cmd.Cfg = &config.Config{}
	cmd.Cfg.Server.Port = 8080
	cmd.Cfg.Server.BaseURL = "http://localhost:8080"
	cmd.Cfg.Database.Name = "url_shortener.db"
	cmd.Cfg.Analytics.BufferSize = 1000
	cmd.Cfg.Analytics.Workers = 5
	cmd.Cfg.Monitor.IntervalMinutes = 5
	router := setupTestRouter()
	mockService := &MockLinkService{}
	
	api := router.Group("/api/v1")
	{
		api.POST("/links", CreateShortLinkHandler(mockService))
	}
	
	requestBody := CreateLinkRequest{
		LongURL: "https://www.example.com",
	}
	jsonData, _ := json.Marshal(requestBody)
	
	expectedLink := &models.Link{
		ID:        1,
		ShortCode: "abc123",
		LongURL:   "https://www.example.com",
		CreatedAt: time.Now(),
	}
	mockService.On("CreateLink", "https://www.example.com").Return(expectedLink, nil)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/links", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "abc123", response["short_code"])
	assert.Equal(t, "https://www.example.com", response["long_url"])
	assert.Contains(t, response, "full_short_url")
	
	mockService.AssertExpectations(t)
}

func TestCreateShortLinkHandler_InvalidURL(t *testing.T) {
	router := setupTestRouter()
	mockService := &MockLinkService{}
	
	api := router.Group("/api/v1")
	{
		api.POST("/links", CreateShortLinkHandler(mockService))
	}
	
	requestBody := CreateLinkRequest{
		LongURL: "not-a-valid-url",
	}
	jsonData, _ := json.Marshal(requestBody)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/links", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestCreateShortLinkHandler_ServiceError(t *testing.T) {
	router := setupTestRouter()
	mockService := &MockLinkService{}
	
	api := router.Group("/api/v1")
	{
		api.POST("/links", CreateShortLinkHandler(mockService))
	}
	
	requestBody := CreateLinkRequest{
		LongURL: "https://www.example.com",
	}
	jsonData, _ := json.Marshal(requestBody)
	
	mockService.On("CreateLink", "https://www.example.com").Return(nil, assert.AnError)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/links", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Equal(t, "Failed to create link", response["error"])
	
	mockService.AssertExpectations(t)
}

func TestRedirectHandler_Success(t *testing.T) {
	router := setupTestRouter()
	mockService := &MockLinkService{}
	clickEventsChan := make(chan models.ClickEvent, 1)
	
	router.GET("/:shortCode", RedirectHandler(mockService, clickEventsChan))
	
	expectedLink := &models.Link{
		ID:        1,
		ShortCode: "abc123",
		LongURL:   "https://www.example.com",
		CreatedAt: time.Now(),
	}
	mockService.On("GetLinkByShortCode", "abc123").Return(expectedLink, nil)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/abc123", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://www.example.com", w.Header().Get("Location"))
	
	mockService.AssertExpectations(t)
}

func TestRedirectHandler_LinkNotFound(t *testing.T) {
	router := setupTestRouter()
	mockService := &MockLinkService{}
	clickEventsChan := make(chan models.ClickEvent, 1)
	
	router.GET("/:shortCode", RedirectHandler(mockService, clickEventsChan))
	
	mockService.On("GetLinkByShortCode", "nonexistent").Return(nil, gorm.ErrRecordNotFound)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Equal(t, "Link not found", response["error"])
	
	mockService.AssertExpectations(t)
}

func TestRedirectHandler_ServiceError(t *testing.T) {
	router := setupTestRouter()
	mockService := &MockLinkService{}
	clickEventsChan := make(chan models.ClickEvent, 1)
	
	router.GET("/:shortCode", RedirectHandler(mockService, clickEventsChan))
	
	mockService.On("GetLinkByShortCode", "error").Return(nil, assert.AnError)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Equal(t, "Internal server error", response["error"])
	
	mockService.AssertExpectations(t)
}

func TestGetLinkStatsHandler_Success(t *testing.T) {
	router := setupTestRouter()
	mockService := &MockLinkService{}
	
	api := router.Group("/api/v1")
	{
		api.GET("/links/:shortCode/stats", GetLinkStatsHandler(mockService))
	}
	
	expectedLink := &models.Link{
		ID:        1,
		ShortCode: "abc123",
		LongURL:   "https://www.example.com",
		CreatedAt: time.Now(),
	}
	mockService.On("GetLinkStats", "abc123").Return(expectedLink, 42, nil)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/links/abc123/stats", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "abc123", response["short_code"])
	assert.Equal(t, "https://www.example.com", response["long_url"])
	assert.Equal(t, float64(42), response["total_clicks"])
	
	mockService.AssertExpectations(t)
}

func TestGetLinkStatsHandler_LinkNotFound(t *testing.T) {
	router := setupTestRouter()
	mockService := &MockLinkService{}
	
	api := router.Group("/api/v1")
	{
		api.GET("/links/:shortCode/stats", GetLinkStatsHandler(mockService))
	}
	
	mockService.On("GetLinkStats", "nonexistent").Return(nil, 0, gorm.ErrRecordNotFound)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/links/nonexistent/stats", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Equal(t, "Link not found", response["error"])
	
	mockService.AssertExpectations(t)
}

func TestGetLinkStatsHandler_ServiceError(t *testing.T) {
	router := setupTestRouter()
	mockService := &MockLinkService{}
	
	api := router.Group("/api/v1")
	{
		api.GET("/links/:shortCode/stats", GetLinkStatsHandler(mockService))
	}
	
	mockService.On("GetLinkStats", "error").Return(nil, 0, assert.AnError)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/links/error/stats", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Equal(t, "Internal server error", response["error"])
	
	mockService.AssertExpectations(t)
}

func TestRedirectHandler_ClickEventChannel(t *testing.T) {
	clickEventsChan := make(chan models.ClickEvent, 1)
	
	router := setupTestRouter()
	mockService := &MockLinkService{}
	
	router.GET("/:shortCode", RedirectHandler(mockService, clickEventsChan))
	
	expectedLink := &models.Link{
		ID:        1,
		ShortCode: "abc123",
		LongURL:   "https://www.example.com",
		CreatedAt: time.Now(),
	}
	mockService.On("GetLinkByShortCode", "abc123").Return(expectedLink, nil)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/abc123", nil)
	req.Header.Set("User-Agent", "TestAgent")
	req.RemoteAddr = "192.168.1.1:12345"
	
	go router.ServeHTTP(w, req)
	
	select {
	case clickEvent := <-clickEventsChan:
		assert.Equal(t, uint(1), clickEvent.LinkID)
		assert.Equal(t, "TestAgent", clickEvent.UserAgent)
		assert.Equal(t, "192.168.1.1", clickEvent.IPAddress)
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for click event")
	}
	
	mockService.AssertExpectations(t)
}