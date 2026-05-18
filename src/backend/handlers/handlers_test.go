package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter() *gin.Engine {
	r := gin.Default()
	r.Use(CORS())
	r.GET("/health", HealthCheck)
	agents := r.Group("/api/v1/agents")
	{
		agents.POST("/register", RegisterAgent)
		agents.POST("/heartbeat", Heartbeat)
		agents.GET("", ListAgents)
		agents.GET("/:id", GetAgent)
	}
	events := r.Group("/api/v1/events")
	{
		events.POST("", CreateEvent)
		events.GET("", ListEvents)
		events.GET("/:id", GetEvent)
		events.GET("/agent/:agent_id", GetAgentEvents)
	}
	alerts := r.Group("/api/v1/alerts")
	{
		alerts.GET("", ListAlerts)
		alerts.GET("/:id", GetAlert)
		alerts.PUT("/:id/resolve", ResolveAlert)
	}
	return r
}

func TestHealthCheck_NoDB(t *testing.T) {
	DB = nil
	router := setupTestRouter()
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestRegisterAgent_InvalidJSON(t *testing.T) {
	DB = nil
	router := setupTestRouter()
	req, _ := http.NewRequest("POST", "/api/v1/agents/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRegisterAgent_MissingHostname(t *testing.T) {
	DB = nil
	router := setupTestRouter()
	body := `{"os_info": "Linux"}`
	req, _ := http.NewRequest("POST", "/api/v1/agents/register", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateEvent_InvalidJSON(t *testing.T) {
	DB = nil
	router := setupTestRouter()
	body := `invalid json`
	req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateEvent_MissingRequiredFields(t *testing.T) {
	DB = nil
	router := setupTestRouter()
	body := `{"agent_id": "test"}`
	req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestResolveAlert_InvalidJSON(t *testing.T) {
	DB = nil
	router := setupTestRouter()
	body := `invalid`
	req, _ := http.NewRequest("PUT", "/api/v1/alerts/test-id/resolve", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCORS_Headers(t *testing.T) {
	DB = nil
	router := setupTestRouter()
	req, _ := http.NewRequest("OPTIONS", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("Expected status 204 for OPTIONS, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Missing CORS header")
	}
}

func TestValidateUUID_Valid(t *testing.T) {
	valid := "550e8400-e29b-41d4-a716-446655440000"
	if !ValidateUUID(valid) {
		t.Error("Expected true for valid UUID")
	}
}

func TestValidateUUID_Invalid(t *testing.T) {
	invalid := []string{
		"invalid-uuid",
		"",
		"550e8400-e29b-41d4-a716",
		"not-a-uuid-at-all",
	}

	for _, uuid := range invalid {
		if ValidateUUID(uuid) {
			t.Errorf("Expected false for invalid UUID: %s", uuid)
		}
	}
}

func TestGetEvent_DBError(t *testing.T) {
	DB = nil
	router := setupTestRouter()
	req, _ := http.NewRequest("GET", "/api/v1/events/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestGetAlert_DBError(t *testing.T) {
	DB = nil
	router := setupTestRouter()
	req, _ := http.NewRequest("GET", "/api/v1/alerts/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestResolveAlert_DBError(t *testing.T) {
	DB = nil
	router := setupTestRouter()
	body := `{"resolved_by": "admin"}`
	req, _ := http.NewRequest("PUT", "/api/v1/alerts/550e8400-e29b-41d4-a716-446655440000/resolve", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestListAgents_DBError(t *testing.T) {
	DB = nil
	router := setupTestRouter()
	req, _ := http.NewRequest("GET", "/api/v1/agents", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestListEvents_DBError(t *testing.T) {
	DB = nil
	router := setupTestRouter()
	req, _ := http.NewRequest("GET", "/api/v1/events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestListAlerts_DBError(t *testing.T) {
	DB = nil
	router := setupTestRouter()
	req, _ := http.NewRequest("GET", "/api/v1/alerts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
