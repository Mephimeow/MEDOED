package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAgent_JSON(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	agent := Agent{
		ID:            id,
		Hostname:      "test-host",
		OSInfo:        "Linux",
		KernelVersion: "5.10.0",
		IPAddress:     "192.168.1.1",
		Status:        "online",
		LastHeartbeat: &now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	data, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("Failed to marshal Agent: %v", err)
	}

	var parsed Agent
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal Agent: %v", err)
	}

	if parsed.Hostname != agent.Hostname {
		t.Errorf("Expected hostname %s, got %s", agent.Hostname, parsed.Hostname)
	}
	if parsed.Status != agent.Status {
		t.Errorf("Expected status %s, got %s", agent.Status, parsed.Status)
	}
}

func TestEvent_JSON(t *testing.T) {
	agentID := uuid.New()
	now := time.Now()

	event := Event{
		ID:        uuid.New(),
		AgentID:   agentID,
		EventType: "process_snapshot",
		Timestamp: now,
		Severity:  "info",
		Payload: map[string]interface{}{
			"total_processes": 100,
			"hostname":        "test-host",
		},
		Source:      "process_monitor",
		Description: "Test event",
		CreatedAt:   now,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal Event: %v", err)
	}

	var parsed Event
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal Event: %v", err)
	}

	if parsed.EventType != event.EventType {
		t.Errorf("Expected event_type %s, got %s", event.EventType, parsed.EventType)
	}
	if parsed.Severity != event.Severity {
		t.Errorf("Expected severity %s, got %s", event.Severity, parsed.Severity)
	}
}

func TestHeartbeatRequest_Validation(t *testing.T) {
	tests := []struct {
		name     string
		req      HeartbeatRequest
		hasError bool
	}{
		{
			name: "valid request",
			req: HeartbeatRequest{
				Hostname:      "test-host",
				OSInfo:        "Linux",
				KernelVersion: "5.10.0",
				IPAddress:     "192.168.1.1",
			},
			hasError: false,
		},
		{
			name: "only hostname",
			req: HeartbeatRequest{
				Hostname: "test-host",
			},
			hasError: false,
		},
		{
			name: "empty hostname",
			req: HeartbeatRequest{
				Hostname: "",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := json.Marshal(tt.req)
			var parsed map[string]interface{}
			json.Unmarshal(data, &parsed)

			if tt.hasError && tt.req.Hostname == "" {
				if parsed["hostname"] != "" {
					t.Error("Expected empty hostname")
				}
			}
		})
	}
}

func TestAlert_Status(t *testing.T) {
	now := time.Now()
	alertID := uuid.New()
	eventID := uuid.New()
	agentID := uuid.New()
	resolvedBy := "admin"

	alert := Alert{
		ID:          alertID,
		EventID:     &eventID,
		AgentID:     agentID,
		RuleName:    "test_rule",
		Severity:    "warning",
		Status:      "open",
		Description: "Test alert",
		CreatedAt:   now,
		ResolvedAt:  &now,
		ResolvedBy:  &resolvedBy,
	}

	if alert.Status != "open" {
		t.Errorf("Expected status 'open', got '%s'", alert.Status)
	}

	alert.Status = "resolved"
	if alert.Status != "resolved" {
		t.Errorf("Expected status 'resolved', got '%s'", alert.Status)
	}
}

func TestResolveRequest(t *testing.T) {
	req := ResolveRequest{
		ResolvedBy: "admin",
	}

	data, _ := json.Marshal(req)
	var parsed ResolveRequest
	json.Unmarshal(data, &parsed)

	if parsed.ResolvedBy != "admin" {
		t.Errorf("Expected resolved_by 'admin', got '%s'", parsed.ResolvedBy)
	}
}

func TestErrorResponse(t *testing.T) {
	errResp := ErrorResponse{
		Error: "something went wrong",
	}

	data, _ := json.Marshal(errResp)
	var parsed ErrorResponse
	json.Unmarshal(data, &parsed)

	if parsed.Error != "something went wrong" {
		t.Errorf("Expected error 'something went wrong', got '%s'", parsed.Error)
	}
}

func TestMessageResponse(t *testing.T) {
	msgResp := MessageResponse{
		Message: "operation successful",
	}

	data, _ := json.Marshal(msgResp)
	var parsed MessageResponse
	json.Unmarshal(data, &parsed)

	if parsed.Message != "operation successful" {
		t.Errorf("Expected message 'operation successful', got '%s'", parsed.Message)
	}
}

func TestPayload_JSONB(t *testing.T) {
	event := Event{
		EventType: "test",
		Severity:  "info",
		Payload: map[string]interface{}{
			"nested": map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
			"array": []interface{}{1, 2, 3},
			"bool":  true,
		},
	}

	data, _ := json.Marshal(event.Payload)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	nested := parsed["nested"].(map[string]interface{})
	if nested["key1"] != "value1" {
		t.Error("Failed to serialize nested payload")
	}
}
