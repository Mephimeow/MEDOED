package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Mephimeow/MEDOED/backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "medoed")
	password := getEnv("DB_PASSWORD", "medoed_secret")
	dbname := getEnv("DB_NAME", "medoed")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	for i := 0; i < 30; i++ {
		err = DB.Ping()
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	return err
}

func InitDBWithDSN(dsn string) error {
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	return DB.Ping()
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func HealthCheck(c *gin.Context) {
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, models.StatusResponse{Status: "unhealthy"})
		return
	}
	err := DB.Ping()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.StatusResponse{Status: "healthy"})
}

func RegisterAgent(c *gin.Context) {
	var req models.HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	var agent models.Agent
	err := DB.QueryRow(`
		INSERT INTO agents (hostname, os_info, kernel_version, ip_address, status, last_heartbeat)
		VALUES ($1, $2, $3, $4, 'online', NOW())
		ON CONFLICT (hostname) DO UPDATE SET
			os_info = EXCLUDED.os_info,
			kernel_version = EXCLUDED.kernel_version,
			ip_address = EXCLUDED.ip_address,
			status = 'online',
			last_heartbeat = NOW(),
			updated_at = NOW()
		RETURNING id, hostname, os_info, kernel_version, ip_address, status, last_heartbeat, created_at, updated_at
	`, req.Hostname, req.OSInfo, req.KernelVersion, req.IPAddress).Scan(
		&agent.ID, &agent.Hostname, &agent.OSInfo, &agent.KernelVersion,
		&agent.IPAddress, &agent.Status, &agent.LastHeartbeat, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func Heartbeat(c *gin.Context) {
	var req models.HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	var agent models.Agent
	err := DB.QueryRow(`
		UPDATE agents SET
			status = 'online',
			last_heartbeat = NOW(),
			updated_at = NOW()
		WHERE hostname = $1
		RETURNING id, hostname, os_info, kernel_version, ip_address, status, last_heartbeat, created_at, updated_at
	`, req.Hostname).Scan(
		&agent.ID, &agent.Hostname, &agent.OSInfo, &agent.KernelVersion,
		&agent.IPAddress, &agent.Status, &agent.LastHeartbeat, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			RegisterAgent(c)
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func ListAgents(c *gin.Context) {
	status := c.Query("status")

	query := `SELECT id, hostname, os_info, kernel_version, ip_address, status, last_heartbeat, created_at, updated_at FROM agents`
	var args []interface{}

	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}
	query += " ORDER BY last_heartbeat DESC"

	rows, err := DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	defer rows.Close()

	var agents []models.Agent
	for rows.Next() {
		var agent models.Agent
		err := rows.Scan(&agent.ID, &agent.Hostname, &agent.OSInfo, &agent.KernelVersion,
			&agent.IPAddress, &agent.Status, &agent.LastHeartbeat, &agent.CreatedAt, &agent.UpdatedAt)
		if err != nil {
			continue
		}
		agents = append(agents, agent)
	}

	if agents == nil {
		agents = []models.Agent{}
	}
	c.JSON(http.StatusOK, agents)
}

func GetAgent(c *gin.Context) {
	id := c.Param("id")

	var agent models.Agent
	err := DB.QueryRow(`
		SELECT id, hostname, os_info, kernel_version, ip_address, status, last_heartbeat, created_at, updated_at
		FROM agents WHERE id = $1
	`, id).Scan(&agent.ID, &agent.Hostname, &agent.OSInfo, &agent.KernelVersion,
		&agent.IPAddress, &agent.Status, &agent.LastHeartbeat, &agent.CreatedAt, &agent.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func GetAgentStatus(c *gin.Context) {
	id := c.Param("id")

	var agent models.Agent
	err := DB.QueryRow(`
		SELECT id, status, last_heartbeat FROM agents WHERE id = $1
	`, id).Scan(&agent.ID, &agent.Status, &agent.LastHeartbeat)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func CreateEvent(c *gin.Context) {
	var req models.EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	payloadJSON, err := json.Marshal(req.Payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid payload"})
		return
	}

	var event models.Event
	err = DB.QueryRow(`
		INSERT INTO events (agent_id, event_type, timestamp, severity, payload, source, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, agent_id, event_type, timestamp, severity, payload, source, description, created_at
	`, req.AgentID, req.EventType, req.Timestamp, req.Severity, payloadJSON, req.Source, req.Description).Scan(
		&event.ID, &event.AgentID, &event.EventType, &event.Timestamp,
		&event.Severity, &payloadJSON, &event.Source, &event.Description, &event.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	json.Unmarshal(payloadJSON, &event.Payload)
	c.JSON(http.StatusCreated, event)
}

func ListEvents(c *gin.Context) {
	eventType := c.Query("type")
	severity := c.Query("severity")
	limit := c.DefaultQuery("limit", "100")
	offset := c.DefaultQuery("offset", "0")

	query := `SELECT id, agent_id, event_type, timestamp, severity, payload, source, description, created_at FROM events WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if eventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", argNum)
		args = append(args, eventType)
		argNum++
	}
	if severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argNum)
		args = append(args, severity)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		var payloadBytes []byte
		err := rows.Scan(&event.ID, &event.AgentID, &event.EventType, &event.Timestamp,
			&event.Severity, &payloadBytes, &event.Source, &event.Description, &event.CreatedAt)
		if err != nil {
			continue
		}
		json.Unmarshal(payloadBytes, &event.Payload)
		events = append(events, event)
	}

	if events == nil {
		events = []models.Event{}
	}
	c.JSON(http.StatusOK, events)
}

func GetEvent(c *gin.Context) {
	id := c.Param("id")

	var event models.Event
	var payloadBytes []byte
	err := DB.QueryRow(`
		SELECT id, agent_id, event_type, timestamp, severity, payload, source, description, created_at
		FROM events WHERE id = $1
	`, id).Scan(&event.ID, &event.AgentID, &event.EventType, &event.Timestamp,
		&event.Severity, &payloadBytes, &event.Source, &event.Description, &event.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	json.Unmarshal(payloadBytes, &event.Payload)
	c.JSON(http.StatusOK, event)
}

func GetAgentEvents(c *gin.Context) {
	agentID := c.Param("agent_id")
	limit := c.DefaultQuery("limit", "100")

	var events []models.Event
	rows, err := DB.Query(`
		SELECT id, agent_id, event_type, timestamp, severity, payload, source, description, created_at
		FROM events WHERE agent_id = $1 ORDER BY timestamp DESC LIMIT $2
	`, agentID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var event models.Event
		var payloadBytes []byte
		err := rows.Scan(&event.ID, &event.AgentID, &event.EventType, &event.Timestamp,
			&event.Severity, &payloadBytes, &event.Source, &event.Description, &event.CreatedAt)
		if err != nil {
			continue
		}
		json.Unmarshal(payloadBytes, &event.Payload)
		events = append(events, event)
	}

	if events == nil {
		events = []models.Event{}
	}
	c.JSON(http.StatusOK, events)
}

func ListAlerts(c *gin.Context) {
	status := c.DefaultQuery("status", "open")

	rows, err := DB.Query(`
		SELECT id, event_id, agent_id, rule_name, severity, status, description, created_at, resolved_at, resolved_by
		FROM alerts WHERE status = $1 ORDER BY created_at DESC
	`, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var alert models.Alert
		err := rows.Scan(&alert.ID, &alert.EventID, &alert.AgentID, &alert.RuleName,
			&alert.Severity, &alert.Status, &alert.Description, &alert.CreatedAt,
			&alert.ResolvedAt, &alert.ResolvedBy)
		if err != nil {
			continue
		}
		alerts = append(alerts, alert)
	}

	if alerts == nil {
		alerts = []models.Alert{}
	}
	c.JSON(http.StatusOK, alerts)
}

func GetAlert(c *gin.Context) {
	id := c.Param("id")

	var alert models.Alert
	err := DB.QueryRow(`
		SELECT id, event_id, agent_id, rule_name, severity, status, description, created_at, resolved_at, resolved_by
		FROM alerts WHERE id = $1
	`, id).Scan(&alert.ID, &alert.EventID, &alert.AgentID, &alert.RuleName,
		&alert.Severity, &alert.Status, &alert.Description, &alert.CreatedAt,
		&alert.ResolvedAt, &alert.ResolvedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "alert not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, alert)
}

func ResolveAlert(c *gin.Context) {
	id := c.Param("id")

	var req models.ResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	result, err := DB.Exec(`
		UPDATE alerts SET status = 'resolved', resolved_at = NOW(), resolved_by = $1
		WHERE id = $2 AND status = 'open'
	`, req.ResolvedBy, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "alert not found or already resolved"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "alert resolved"})
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func GetDB() *sql.DB {
	return DB
}

func SetDB(db *sql.DB) {
	DB = db
}

func ValidateUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func GetAgentIDByHostname(hostname string) (uuid.UUID, error) {
	var id uuid.UUID
	err := DB.QueryRow("SELECT id FROM agents WHERE hostname = $1", hostname).Scan(&id)
	return id, err
}