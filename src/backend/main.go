package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Mephimeow/MEDOED/backend/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := handlers.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer handlers.CloseDB()

	log.Println("Connected to PostgreSQL database")

	r := gin.Default()

	r.Use(handlers.CORS())

	r.GET("/health", handlers.HealthCheck)

	agents := r.Group("/api/v1/agents")
	{
		agents.POST("/register", handlers.RegisterAgent)
		agents.POST("/heartbeat", handlers.Heartbeat)
		agents.GET("", handlers.ListAgents)
		agents.GET("/:id", handlers.GetAgent)
		agents.GET("/:id/status", handlers.GetAgentStatus)
	}

	events := r.Group("/api/v1/events")
	{
		events.POST("", handlers.CreateEvent)
		events.GET("", handlers.ListEvents)
		events.GET("/:id", handlers.GetEvent)
		events.GET("/agent/:agent_id", handlers.GetAgentEvents)
	}

	alerts := r.Group("/api/v1/alerts")
	{
		alerts.GET("", handlers.ListAlerts)
		alerts.GET("/:id", handlers.GetAlert)
		alerts.PUT("/:id/resolve", handlers.ResolveAlert)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Starting MEDOED Backend on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}