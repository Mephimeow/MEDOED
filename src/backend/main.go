package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Mephimeow/MEDOED/backend/handlers"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	TLS      TLSConfig      `json:"tls"`
	Auth     AuthConfig     `json:"auth"`
	Logging  LoggingConfig  `json:"logging"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SSLMode  string `json:"sslmode"`
}

type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

type AuthConfig struct {
	Enabled bool     `json:"enabled"`
	APIKeys []string `json:"api_keys"`
}

type LoggingConfig struct {
	Level string `json:"level"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func getEnvOrDefault(env, defaultVal string) string {
	if val := os.Getenv(env); val != "" {
		return val
	}
	return defaultVal
}

func main() {
	configPath := getEnvOrDefault("CONFIG_PATH", "config.json")
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbHost := getEnvOrDefault("DB_HOST", cfg.Database.Host)
	dbPort := getEnvOrDefault("DB_PORT", fmt.Sprintf("%d", cfg.Database.Port))
	dbUser := getEnvOrDefault("DB_USER", cfg.Database.User)
	dbPassword := getEnvOrDefault("DB_PASSWORD", cfg.Database.Password)
	dbName := getEnvOrDefault("DB_NAME", cfg.Database.Name)
	dbSSLMode := getEnvOrDefault("DB_SSLMODE", cfg.Database.SSLMode)

	if err := handlers.InitDBWithDSN(fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode,
	)); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer handlers.CloseDB()

	log.Println("Connected to PostgreSQL database")

	serverHost := getEnvOrDefault("SERVER_HOST", cfg.Server.Host)
	serverPort := getEnvOrDefault("SERVER_PORT", strconv.Itoa(cfg.Server.Port))

	r := gin.Default()

	handlers.InitAuth(cfg.Auth.Enabled, cfg.Auth.APIKeys)

	r.Use(handlers.CORS())

	r.GET("/", handlers.DashboardHandler)
	r.GET("/agents", handlers.AgentsPageHandler)
	r.GET("/events", handlers.EventsPageHandler)
	r.GET("/alerts", handlers.AlertsPageHandler)

	r.GET("/health", handlers.HealthCheck)

	api := r.Group("/api/v1")
	api.Use(handlers.AuthMiddleware())
	{
		agents := api.Group("/agents")
		{
			agents.POST("/register", handlers.RegisterAgent)
			agents.POST("/heartbeat", handlers.Heartbeat)
			agents.GET("", handlers.ListAgents)
			agents.GET("/:id", handlers.GetAgent)
			agents.GET("/:id/status", handlers.GetAgentStatus)
		}

		events := api.Group("/events")
		{
			events.POST("", handlers.CreateEvent)
			events.GET("", handlers.ListEvents)
			events.GET("/:id", handlers.GetEvent)
			events.GET("/agent/:agent_id", handlers.GetAgentEvents)
		}

		alerts := api.Group("/alerts")
		{
			alerts.GET("", handlers.ListAlerts)
			alerts.GET("/:id", handlers.GetAlert)
			alerts.PUT("/:id/resolve", handlers.ResolveAlert)
		}
	}

	bindAddr := fmt.Sprintf("%s:%s", serverHost, serverPort)
	server := &http.Server{
		Addr:         bindAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if cfg.TLS.Enabled && cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
		log.Printf("Starting MEDOED Backend on %s (TLS)", bindAddr)
		if err := server.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	} else {
		log.Printf("Starting MEDOED Backend on %s", bindAddr)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}
