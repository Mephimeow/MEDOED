package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var db *sql.DB

type Agent struct {
	ID            uuid.UUID  `json:"id"`
	Hostname      string     `json:"hostname"`
	OSInfo        string     `json:"os_info"`
	KernelVersion string     `json:"kernel_version"`
	IPAddress     string     `json:"ip_address"`
	Status        string     `json:"status"`
	LastHeartbeat *time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type HeartbeatRequest struct {
	Hostname      string `json:"hostname"`
	OSInfo        string `json:"os_info"`
	KernelVersion string `json:"kernel_version"`
	IPAddress     string `json:"ip_address"`
}

type EventRequest struct {
	AgentID     string                 `json:"agent_id"`
	EventType   string                 `json:"event_type"`
	Timestamp   time.Time              `json:"timestamp"`
	Severity    string                 `json:"severity"`
	Payload     map[string]interface{} `json:"payload"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
}

type Event struct {
	ID          uuid.UUID              `json:"id"`
	AgentID     uuid.UUID              `json:"agent_id"`
	EventType   string                 `json:"event_type"`
	Timestamp   time.Time              `json:"timestamp"`
	Severity    string                 `json:"severity"`
	Payload     map[string]interface{} `json:"payload"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"created_at"`
}

type Alert struct {
	ID          uuid.UUID  `json:"id"`
	EventID     *uuid.UUID `json:"event_id"`
	AgentID     uuid.UUID  `json:"agent_id"`
	RuleName    string     `json:"rule_name"`
	Severity    string     `json:"severity"`
	Status      string     `json:"status"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	ResolvedBy  *string    `json:"resolved_by"`
}

func initDB() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "medoed")
	password := getEnv("DB_PASSWORD", "medoed_secret")
	dbname := getEnv("DB_NAME", "medoed")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Waiting for database... (attempt %d/30)", i+1)
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to PostgreSQL database")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	initDB()
	defer db.Close()

	r := gin.Default()

	r.Use(corsMiddleware())

	r.GET("/health", healthCheck)

	agents := r.Group("/api/v1/agents")
	{
		agents.POST("/register", registerAgent)
		agents.POST("/heartbeat", heartbeat)
		agents.GET("", listAgents)
		agents.GET("/:id", getAgent)
		agents.GET("/:id/status", getAgentStatus)
	}

	events := r.Group("/api/v1/events")
	{
		events.POST("", createEvent)
		events.GET("", listEvents)
		events.GET("/:id", getEvent)
		events.GET("/agent/:agent_id", getAgentEvents)
	}

	alerts := r.Group("/api/v1/alerts")
	{
		alerts.GET("", listAlerts)
		alerts.GET("/:id", getAlert)
		alerts.PUT("/:id/resolve", resolveAlert)
	}

	log.Println("Starting MEDOED Backend on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
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

func healthCheck(c *gin.Context) {
	err := db.Ping()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func registerAgent(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var agent Agent
	err := db.QueryRow(`
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func heartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var agent Agent
	err := db.QueryRow(`
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
			registerAgent(c)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func listAgents(c *gin.Context) {
	status := c.Query("status")

	query := `SELECT id, hostname, os_info, kernel_version, ip_address, status, last_heartbeat, created_at, updated_at FROM agents`
	var args []interface{}

	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}
	query += " ORDER BY last_heartbeat DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		err := rows.Scan(&agent.ID, &agent.Hostname, &agent.OSInfo, &agent.KernelVersion,
			&agent.IPAddress, &agent.Status, &agent.LastHeartbeat, &agent.CreatedAt, &agent.UpdatedAt)
		if err != nil {
			continue
		}
		agents = append(agents, agent)
	}

	if agents == nil {
		agents = []Agent{}
	}
	c.JSON(http.StatusOK, agents)
}

func getAgent(c *gin.Context) {
	id := c.Param("id")

	var agent Agent
	err := db.QueryRow(`
		SELECT id, hostname, os_info, kernel_version, ip_address, status, last_heartbeat, created_at, updated_at
		FROM agents WHERE id = $1
	`, id).Scan(&agent.ID, &agent.Hostname, &agent.OSInfo, &agent.KernelVersion,
		&agent.IPAddress, &agent.Status, &agent.LastHeartbeat, &agent.CreatedAt, &agent.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func getAgentStatus(c *gin.Context) {
	id := c.Param("id")

	var agent Agent
	err := db.QueryRow(`
		SELECT id, status, last_heartbeat FROM agents WHERE id = $1
	`, id).Scan(&agent.ID, &agent.Status, &agent.LastHeartbeat)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func createEvent(c *gin.Context) {
	var req EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payloadJSON, err := json.Marshal(req.Payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	var event Event
	err = db.QueryRow(`
		INSERT INTO events (agent_id, event_type, timestamp, severity, payload, source, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, agent_id, event_type, timestamp, severity, payload, source, description, created_at
	`, req.AgentID, req.EventType, req.Timestamp, req.Severity, payloadJSON, req.Source, req.Description).Scan(
		&event.ID, &event.AgentID, &event.EventType, &event.Timestamp,
		&event.Severity, &payloadJSON, &event.Source, &event.Description, &event.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	json.Unmarshal(payloadJSON, &event.Payload)
	c.JSON(http.StatusCreated, event)
}

func listEvents(c *gin.Context) {
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

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
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
		events = []Event{}
	}
	c.JSON(http.StatusOK, events)
}

func getEvent(c *gin.Context) {
	id := c.Param("id")

	var event Event
	var payloadBytes []byte
	err := db.QueryRow(`
		SELECT id, agent_id, event_type, timestamp, severity, payload, source, description, created_at
		FROM events WHERE id = $1
	`, id).Scan(&event.ID, &event.AgentID, &event.EventType, &event.Timestamp,
		&event.Severity, &payloadBytes, &event.Source, &event.Description, &event.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	json.Unmarshal(payloadBytes, &event.Payload)
	c.JSON(http.StatusOK, event)
}

func getAgentEvents(c *gin.Context) {
	agentID := c.Param("agent_id")
	limit := c.DefaultQuery("limit", "100")

	var events []Event
	rows, err := db.Query(`
		SELECT id, agent_id, event_type, timestamp, severity, payload, source, description, created_at
		FROM events WHERE agent_id = $1 ORDER BY timestamp DESC LIMIT $2
	`, agentID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var event Event
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
		events = []Event{}
	}
	c.JSON(http.StatusOK, events)
}

func listAlerts(c *gin.Context) {
	status := c.DefaultQuery("status", "open")

	rows, err := db.Query(`
		SELECT id, event_id, agent_id, rule_name, severity, status, description, created_at, resolved_at, resolved_by
		FROM alerts WHERE status = $1 ORDER BY created_at DESC
	`, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var alert Alert
		err := rows.Scan(&alert.ID, &alert.EventID, &alert.AgentID, &alert.RuleName,
			&alert.Severity, &alert.Status, &alert.Description, &alert.CreatedAt,
			&alert.ResolvedAt, &alert.ResolvedBy)
		if err != nil {
			continue
		}
		alerts = append(alerts, alert)
	}

	if alerts == nil {
		alerts = []Alert{}
	}
	c.JSON(http.StatusOK, alerts)
}

func getAlert(c *gin.Context) {
	id := c.Param("id")

	var alert Alert
	err := db.QueryRow(`
		SELECT id, event_id, agent_id, rule_name, severity, status, description, created_at, resolved_at, resolved_by
		FROM alerts WHERE id = $1
	`, id).Scan(&alert.ID, &alert.EventID, &alert.AgentID, &alert.RuleName,
		&alert.Severity, &alert.Status, &alert.Description, &alert.CreatedAt,
		&alert.ResolvedAt, &alert.ResolvedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alert)
}

type ResolveRequest struct {
	ResolvedBy string `json:"resolved_by"`
}

func resolveAlert(c *gin.Context) {
	id := c.Param("id")

	var req ResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec(`
		UPDATE alerts SET status = 'resolved', resolved_at = NOW(), resolved_by = $1
		WHERE id = $2 AND status = 'open'
	`, req.ResolvedBy, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found or already resolved"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alert resolved"})
}