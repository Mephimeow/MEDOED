package handlers

import (
	"log"

	"github.com/Mephimeow/MEDOED/backend/models"
)

type Rule struct {
	Name        string
	Description string
	Severity    string
	Condition   func(event models.Event) bool
}

type RuleConfig struct {
	Enabled           bool `json:"enabled"`
	CriticalSeverity  bool `json:"critical_severity"`
	HighSeverity      bool `json:"high_severity"`
	SuspiciousProcess bool `json:"suspicious_process"`
}

var DefaultRuleConfig = RuleConfig{
	Enabled:           true,
	CriticalSeverity:  true,
	HighSeverity:      true,
	SuspiciousProcess: true,
}

var BuiltinRules []Rule

func init() {
	BuiltinRules = []Rule{
		{
			Name:        "critical_severity",
			Description: "Event with critical severity detected",
			Severity:    "critical",
			Condition: func(event models.Event) bool {
				return event.Severity == "critical"
			},
		},
		{
			Name:        "high_severity",
			Description: "Event with high severity detected",
			Severity:    "high",
			Condition: func(event models.Event) bool {
				return event.Severity == "high"
			},
		},
		{
			Name:        "suspicious_process",
			Description: "Suspicious process detected on agent",
			Severity:    "high",
			Condition: func(event models.Event) bool {
				return event.EventType == "suspicious_process"
			},
		},
	}
}

func EvaluateEvent(event models.Event, cfg RuleConfig) []models.Alert {
	if !cfg.Enabled {
		return nil
	}

	var alerts []models.Alert
	for _, rule := range BuiltinRules {
		if !isRuleEnabled(rule.Name, cfg) {
			continue
		}
		if rule.Condition(event) {
			alerts = append(alerts, models.Alert{
				EventID:     &event.ID,
				AgentID:     event.AgentID,
				RuleName:    rule.Name,
				Severity:    rule.Severity,
				Status:      "open",
				Description: rule.Description,
			})
		}
	}
	return alerts
}

func isRuleEnabled(name string, cfg RuleConfig) bool {
	switch name {
	case "critical_severity":
		return cfg.CriticalSeverity
	case "high_severity":
		return cfg.HighSeverity
	case "suspicious_process":
		return cfg.SuspiciousProcess
	}
	return true
}

func CreateAlertsForEvent(event models.Event, cfg RuleConfig) {
	alerts := EvaluateEvent(event, cfg)
	for _, alert := range alerts {
		_, err := DB.Exec(`
			INSERT INTO alerts (event_id, agent_id, rule_name, severity, status, description)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, alert.EventID, alert.AgentID, alert.RuleName, alert.Severity, alert.Status, alert.Description)
		if err != nil {
			log.Printf("Failed to create alert: %v", err)
		}
	}
}

func GetActiveAlertCount() int {
	if DB == nil {
		return 0
	}
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM alerts WHERE status = 'open'`).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}
