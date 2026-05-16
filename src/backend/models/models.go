package models

import (
	"time"

	"github.com/google/uuid"
)

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
	Hostname      string `json:"hostname" binding:"required"`
	OSInfo        string `json:"os_info"`
	KernelVersion string `json:"kernel_version"`
	IPAddress     string `json:"ip_address"`
}

type EventRequest struct {
	AgentID     string                 `json:"agent_id" binding:"required"`
	EventType   string                 `json:"event_type" binding:"required"`
	Timestamp   time.Time              `json:"timestamp" binding:"required"`
	Severity    string                 `json:"severity" binding:"required"`
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

type ResolveRequest struct {
	ResolvedBy string `json:"resolved_by"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type StatusResponse struct {
	Status string `json:"status"`
}