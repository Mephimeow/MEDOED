package handlers

import (
	"html/template"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardData struct {
	CurrentPage  string
	Stats        DashboardStats
	RecentEvents []EventRow
	ActiveAgents []AgentRow
}

type DashboardStats struct {
	TotalAgents   int
	OnlineAgents  int
	OfflineAgents int
	ActiveAlerts  int
}

type AgentRow struct {
	Hostname      string
	IPAddress     string
	OSInfo        string
	KernelVersion string
	Status        string
	LastHeartbeat string
	CreatedAt     string
}

type EventRow struct {
	AgentID     string
	EventType   string
	Description string
	CreatedAt   string
}

type AlertRow struct {
	AgentID     string
	RuleName    string
	Description string
	Severity    string
	Status      string
	CreatedAt   string
}

func loadTemplates() *template.Template {
	tmpl := template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/dashboard.html",
		"templates/agents.html",
		"templates/events.html",
		"templates/alerts.html",
	))
	return tmpl
}

func getDashboardData() DashboardData {
	data := DashboardData{
		CurrentPage: "dashboard",
		Stats:       DashboardStats{},
	}

	if DB == nil {
		return data
	}

	rows, err := DB.Query(`
		SELECT COUNT(*) FROM agents
	`)
	if err == nil {
		defer rows.Close()
		if rows.Next() {
			rows.Scan(&data.Stats.TotalAgents)
		}
	}

	rows2, err := DB.Query(`
		SELECT COUNT(*) FROM agents WHERE status = 'online'
	`)
	if err == nil {
		defer rows2.Close()
		if rows2.Next() {
			rows2.Scan(&data.Stats.OnlineAgents)
		}
	}
	data.Stats.OfflineAgents = data.Stats.TotalAgents - data.Stats.OnlineAgents

	rows3, err := DB.Query(`
		SELECT COUNT(*) FROM alerts WHERE status = 'open'
	`)
	if err == nil {
		defer rows3.Close()
		if rows3.Next() {
			rows3.Scan(&data.Stats.ActiveAlerts)
		}
	}

	rows4, err := DB.Query(`
		SELECT agent_id, event_type, description, created_at 
		FROM events 
		ORDER BY created_at DESC 
		LIMIT 10
	`)
	if err == nil {
		defer rows4.Close()
		for rows4.Next() {
			var e EventRow
			var t time.Time
			if rows4.Scan(&e.AgentID, &e.EventType, &e.Description, &t) == nil {
				e.CreatedAt = t.Format("02.01.2006 15:04")
				data.RecentEvents = append(data.RecentEvents, e)
			}
		}
	}

	rows5, err := DB.Query(`
		SELECT hostname, ip_address, os_info, kernel_version, status, last_heartbeat, created_at 
		FROM agents 
		WHERE status = 'online'
		ORDER BY last_heartbeat DESC 
		LIMIT 10
	`)
	if err == nil {
		defer rows5.Close()
		for rows5.Next() {
			var a AgentRow
			var t1, t2 time.Time
			if rows5.Scan(&a.Hostname, &a.IPAddress, &a.OSInfo, &a.KernelVersion, &a.Status, &t1, &t2) == nil {
				a.LastHeartbeat = t1.Format("02.01.2006 15:04")
				a.CreatedAt = t2.Format("02.01.2006 15:04")
				data.ActiveAgents = append(data.ActiveAgents, a)
			}
		}
	}

	return data
}

func DashboardHandler(c *gin.Context) {
	tmpl := loadTemplates()
	data := getDashboardData()
	tmpl.ExecuteTemplate(c.Writer, "base", data)
}

func AgentsPageHandler(c *gin.Context) {
	tmpl := loadTemplates()

	data := struct {
		CurrentPage  string
		TotalCount   int
		OnlineCount  int
		OfflineCount int
		Agents       []AgentRow
	}{
		CurrentPage: "agents",
		Agents:      []AgentRow{},
	}

	if DB != nil {
		rows, err := DB.Query(`
			SELECT hostname, ip_address, os_info, kernel_version, status, last_heartbeat, created_at 
			FROM agents 
			ORDER BY created_at DESC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var a AgentRow
				var t1, t2 time.Time
				if rows.Scan(&a.Hostname, &a.IPAddress, &a.OSInfo, &a.KernelVersion, &a.Status, &t1, &t2) == nil {
					a.LastHeartbeat = t1.Format("02.01.2006 15:04")
					a.CreatedAt = t2.Format("02.01.2006 15:04")
					data.Agents = append(data.Agents, a)
					data.TotalCount++
					if a.Status == "online" {
						data.OnlineCount++
					} else {
						data.OfflineCount++
					}
				}
			}
		}
	}

	tmpl.ExecuteTemplate(c.Writer, "base", data)
}

func EventsPageHandler(c *gin.Context) {
	tmpl := loadTemplates()

	data := struct {
		CurrentPage string
		TotalCount  int
		Events      []EventRow
	}{
		CurrentPage: "events",
		Events:      []EventRow{},
	}

	if DB != nil {
		rows, err := DB.Query(`
			SELECT agent_id, event_type, description, created_at 
			FROM events 
			ORDER BY created_at DESC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var e EventRow
				var t time.Time
				if rows.Scan(&e.AgentID, &e.EventType, &e.Description, &t) == nil {
					e.CreatedAt = t.Format("02.01.2006 15:04")
					data.Events = append(data.Events, e)
					data.TotalCount++
				}
			}
		}
	}

	tmpl.ExecuteTemplate(c.Writer, "base", data)
}

func AlertsPageHandler(c *gin.Context) {
	tmpl := loadTemplates()

	data := struct {
		CurrentPage string
		ActiveCount int
		TotalCount  int
		Alerts      []AlertRow
	}{
		CurrentPage: "alerts",
		Alerts:      []AlertRow{},
	}

	if DB != nil {
		rows, err := DB.Query(`
			SELECT agent_id, rule_name, description, severity, status, created_at 
			FROM alerts 
			ORDER BY created_at DESC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var a AlertRow
				var t time.Time
				if rows.Scan(&a.AgentID, &a.RuleName, &a.Description, &a.Severity, &a.Status, &t) == nil {
					a.CreatedAt = t.Format("02.01.2006 15:04")
					data.Alerts = append(data.Alerts, a)
					data.TotalCount++
					if a.Status == "active" {
						data.ActiveCount++
					}
				}
			}
		}
	}

	tmpl.ExecuteTemplate(c.Writer, "base", data)
}
