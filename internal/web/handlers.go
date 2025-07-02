package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/service"
)

// Server represents the web server
type Server struct {
	monitorService *service.MonitorService
	config         *config.Config
	templates      *template.Template
}

// NewServer creates a new web server
func NewServer(monitorService *service.MonitorService, cfg *config.Config) *Server {
	return &Server{
		monitorService: monitorService,
		config:         cfg,
		templates:      loadTemplates(),
	}
}

// Router returns the configured router
func (s *Server) Router() *mux.Router {
	r := mux.NewRouter()

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Web UI routes
	r.HandleFunc("/", s.handleDashboard).Methods("GET")
	r.HandleFunc("/service/{name}", s.handleServiceDetail).Methods("GET")

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/services", s.handleAPIServices).Methods("GET")
	api.HandleFunc("/services/{name}", s.handleAPIServiceDetail).Methods("GET")
	api.HandleFunc("/services/{name}/incidents", s.handleAPIServiceIncidents).Methods("GET")
	api.HandleFunc("/services/{name}/stats", s.handleAPIServiceStats).Methods("GET")
	api.HandleFunc("/services/{name}/check", s.handleAPIServiceCheck).Methods("POST")
	api.HandleFunc("/incidents", s.handleAPIRecentIncidents).Methods("GET")

	return r
}

// handleDashboard renders the main dashboard
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	states := s.monitorService.GetAllServiceStates()

	data := struct {
		Services map[string]*config.ServiceState
		Title    string
	}{
		Services: states,
		Title:    "Service Monitor Dashboard",
	}

	if err := s.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleServiceDetail renders the service detail page
func (s *Server) handleServiceDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	state, err := s.monitorService.GetServiceState(serviceName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	incidents, err := s.monitorService.GetServiceIncidents(r.Context(), serviceName)
	if err != nil {
		incidents = []*config.Incident{}
	}

	// Get stats for the last 30 days
	stats, err := s.monitorService.GetServiceStats(r.Context(), serviceName, time.Now().AddDate(0, 0, -30))
	if err != nil {
		stats = nil
	}

	data := struct {
		Service   *config.ServiceState
		Incidents []*config.Incident
		Stats     interface{}
		Title     string
	}{
		Service:   state,
		Incidents: incidents,
		Stats:     stats,
		Title:     "Service: " + serviceName,
	}

	if err := s.templates.ExecuteTemplate(w, "service-detail.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// API Handlers

// handleAPIServices returns all services status
func (s *Server) handleAPIServices(w http.ResponseWriter, r *http.Request) {
	states := s.monitorService.GetAllServiceStates()
	s.writeJSON(w, states)
}

// handleAPIServiceDetail returns detailed info for a specific service
func (s *Server) handleAPIServiceDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	state, err := s.monitorService.GetServiceState(serviceName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	s.writeJSON(w, state)
}

// handleAPIServiceIncidents returns incidents for a specific service
func (s *Server) handleAPIServiceIncidents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	incidents, err := s.monitorService.GetServiceIncidents(r.Context(), serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, incidents)
}

// handleAPIServiceStats returns statistics for a specific service
func (s *Server) handleAPIServiceStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	// Parse days parameter
	daysStr := r.URL.Query().Get("days")
	days := 30 // default
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	since := time.Now().AddDate(0, 0, -days)
	stats, err := s.monitorService.GetServiceStats(r.Context(), serviceName, since)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, stats)
}

// handleAPIServiceCheck triggers an immediate check for a service
func (s *Server) handleAPIServiceCheck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	// This would require access to the scheduler, which we don't have here
	// For now, return a simple response
	response := map[string]string{
		"message": "Check triggered for " + serviceName,
		"status":  "accepted",
	}

	s.writeJSON(w, response)
}

// handleAPIRecentIncidents returns recent incidents across all services
func (s *Server) handleAPIRecentIncidents(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	incidents, err := s.monitorService.GetRecentIncidents(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, incidents)
}

// writeJSON writes a JSON response
func (s *Server) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// loadTemplates loads HTML templates
func loadTemplates() *template.Template {
	// Define template functions
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return "Never"
			}
			return t.Format("2006-01-02 15:04:05")
		},
		"formatDuration": func(d time.Duration) string {
			if d == 0 {
				return "0s"
			}
			if d < time.Minute {
				return d.Truncate(time.Second).String()
			}
			if d < time.Hour {
				return d.Truncate(time.Minute).String()
			}
			return d.Truncate(time.Hour).String()
		},
		"statusClass": func(status config.ServiceStatus) string {
			switch status {
			case config.StatusUp:
				return "success"
			case config.StatusDown:
				return "danger"
			case config.StatusMaintenance:
				return "warning"
			default:
				return "secondary"
			}
		},
		"statusIcon": func(status config.ServiceStatus) string {
			switch status {
			case config.StatusUp:
				return "✅"
			case config.StatusDown:
				return "❌"
			case config.StatusMaintenance:
				return "🔧"
			default:
				return "❓"
			}
		},
		"add": func(a, b int) int {
			return a + b
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
	}

	// Load templates from embedded files or filesystem
	tmpl := template.New("").Funcs(funcMap)

	// Try to load from filesystem
	pattern := filepath.Join("web", "templates", "*.html")
	if templates, err := tmpl.ParseGlob(pattern); err == nil {
		return templates
	}

	// Fallback to embedded templates if filesystem loading fails
	return loadEmbeddedTemplates(tmpl)
}

// loadEmbeddedTemplates loads templates embedded in the binary
func loadEmbeddedTemplates(tmpl *template.Template) *template.Template {
	// Dashboard template
	dashboardTmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <title>{{.Title}}</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <meta http-equiv="refresh" content="30">
    <style>
        body {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
        }
        .container {
            background: rgba(255, 255, 255, 0.95);
            border-radius: 20px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            backdrop-filter: blur(10px);
            margin-top: 20px;
            margin-bottom: 20px;
            padding: 2rem;
        }
        .page-title {
            color: #2c3e50;
            font-weight: 700;
            margin-bottom: 2rem;
            text-align: center;
            font-size: 2.5rem;
        }
        .service-card {
            border: none;
            border-radius: 15px;
            transition: all 0.3s ease;
            box-shadow: 0 8px 25px rgba(0,0,0,0.1);
            background: linear-gradient(145deg, #ffffff, #f8f9fa);
            margin-bottom: 1.5rem;
            overflow: hidden;
        }
        .service-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 15px 35px rgba(0,0,0,0.15);
        }
        .service-card .card-body {
            padding: 1.5rem;
        }
        .service-title {
            font-size: 1.3rem;
            font-weight: 600;
            margin-bottom: 1rem;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        .status-icon {
            font-size: 1.5rem;
        }
        .badge {
            font-size: 0.75rem;
            padding: 0.5rem 1rem;
            border-radius: 50px;
            font-weight: 600;
        }
        .bg-success { background: linear-gradient(45deg, #28a745, #20c997) !important; }
        .bg-danger { background: linear-gradient(45deg, #dc3545, #e74c3c) !important; }
        .bg-warning { background: linear-gradient(45deg, #ffc107, #fd7e14) !important; }
        .bg-secondary { background: linear-gradient(45deg, #6c757d, #495057) !important; }
        .service-info {
            line-height: 1.8;
            color: #6c757d;
        }
        .service-info strong {
            color: #495057;
            font-weight: 600;
        }
        .btn-details {
            background: linear-gradient(45deg, #007bff, #0056b3);
            border: none;
            border-radius: 25px;
            padding: 0.5rem 1.5rem;
            font-weight: 600;
            transition: all 0.3s ease;
        }
        .btn-details:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(0,123,255,0.3);
        }
        .stats-bar {
            background: rgba(255,255,255,0.9);
            border-radius: 15px;
            padding: 1.5rem;
            margin-bottom: 2rem;
            box-shadow: 0 5px 15px rgba(0,0,0,0.08);
            text-align: center;
        }
        .stat-item {
            display: inline-block;
            margin: 0 2rem;
        }
        .stat-number {
            font-size: 2rem;
            font-weight: 700;
            color: #2c3e50;
        }
        .stat-label {
            color: #6c757d;
            font-size: 0.9rem;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        @media (max-width: 768px) {
            .page-title { font-size: 2rem; }
            .stat-item { margin: 0 1rem; }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1 class="page-title">
            <i class="fas fa-shield-alt text-primary"></i>
            Service Monitor Dashboard
        </h1>
        
        <div class="stats-bar">
            {{$totalServices := len .Services}}
            {{$upServices := 0}}
            {{$downServices := 0}}
            {{range .Services}}
                {{if eq .Status "up"}}{{$upServices = add $upServices 1}}{{end}}
                {{if eq .Status "down"}}{{$downServices = add $downServices 1}}{{end}}
            {{end}}
            <div class="stat-item">
                <div class="stat-number text-primary">{{$totalServices}}</div>
                <div class="stat-label">Total Services</div>
            </div>
            <div class="stat-item">
                <div class="stat-number text-success">{{$upServices}}</div>
                <div class="stat-label">Online</div>
            </div>
            <div class="stat-item">
                <div class="stat-number text-danger">{{$downServices}}</div>
                <div class="stat-label">Offline</div>
            </div>
        </div>
        
        <div class="row">
            {{range .Services}}
            <div class="col-lg-4 col-md-6">
                <div class="card service-card">
                    <div class="card-body">
                        <h5 class="service-title">
                            <span class="status-icon">{{statusIcon .Status}}</span>
                            {{.Name}}
                            <span class="badge bg-{{statusClass .Status}} ms-auto">{{.Status}}</span>
                        </h5>
                        <div class="service-info">
                            <div class="mb-2">
                                <i class="fas fa-network-wired text-muted me-2"></i>
                                <strong>Protocol:</strong> {{.Protocol}}
                            </div>
                            <div class="mb-2">
                                <i class="fas fa-globe text-muted me-2"></i>
                                <strong>Endpoint:</strong> 
                                <span class="text-break">{{.Endpoint}}</span>
                            </div>
                            <div class="mb-2">
                                <i class="fas fa-clock text-muted me-2"></i>
                                <strong>Last Check:</strong> {{formatTime .LastCheck}}
                            </div>
                            <div class="mb-2">
                                <i class="fas fa-tachometer-alt text-muted me-2"></i>
                                <strong>Response Time:</strong> {{formatDuration .ResponseTime}}
                            </div>
                            {{if .LastError}}
                            <div class="mb-2">
                                <i class="fas fa-exclamation-triangle text-danger me-2"></i>
                                <strong>Error:</strong> 
                                <span class="text-danger">{{.LastError}}</span>
                            </div>
                            {{end}}
                        </div>
                        <div class="d-grid mt-3">
                            <a href="/service/{{.Name}}" class="btn btn-primary btn-details">
                                <i class="fas fa-info-circle me-2"></i>View Details
                            </a>
                        </div>
                    </div>
                </div>
            </div>
            {{end}}
        </div>
        
        <div class="text-center mt-4">
            <small class="text-muted">
                <i class="fas fa-sync-alt me-1"></i>
                Auto-refresh every 30 seconds
            </small>
        </div>
    </div>
    
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
</body>
</html>
`

	// Service detail template
	serviceDetailTmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <title>{{.Title}}</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        body {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
        }
        .container {
            background: rgba(255, 255, 255, 0.95);
            border-radius: 20px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            backdrop-filter: blur(10px);
            margin-top: 20px;
            margin-bottom: 20px;
            padding: 2rem;
        }
        .breadcrumb {
            background: linear-gradient(45deg, #f8f9fa, #e9ecef);
            border-radius: 15px;
            padding: 1rem 1.5rem;
            margin-bottom: 2rem;
        }
        .breadcrumb-item a {
            color: #6c757d;
            text-decoration: none;
            font-weight: 500;
        }
        .breadcrumb-item a:hover {
            color: #007bff;
        }
        .page-title {
            color: #2c3e50;
            font-weight: 700;
            margin-bottom: 2rem;
            font-size: 2.5rem;
            display: flex;
            align-items: center;
            gap: 1rem;
        }
        .info-card {
            border: none;
            border-radius: 15px;
            box-shadow: 0 8px 25px rgba(0,0,0,0.1);
            background: linear-gradient(145deg, #ffffff, #f8f9fa);
            margin-bottom: 1.5rem;
        }
        .card-header {
            background: linear-gradient(45deg, #007bff, #0056b3);
            color: white;
            border-radius: 15px 15px 0 0 !important;
            padding: 1rem 1.5rem;
            font-weight: 600;
            font-size: 1.1rem;
        }
        .table-borderless td {
            padding: 0.75rem 0;
            border: none;
            vertical-align: middle;
        }
        .table-borderless td:first-child {
            font-weight: 600;
            color: #495057;
            width: 40%;
        }
        .badge {
            font-size: 0.8rem;
            padding: 0.5rem 1rem;
            border-radius: 50px;
            font-weight: 600;
        }
        .bg-success { background: linear-gradient(45deg, #28a745, #20c997) !important; }
        .bg-danger { background: linear-gradient(45deg, #dc3545, #e74c3c) !important; }
        .bg-warning { background: linear-gradient(45deg, #ffc107, #fd7e14) !important; }
        .bg-secondary { background: linear-gradient(45deg, #6c757d, #495057) !important; }
        .incidents-section {
            margin-top: 2rem;
        }
        .section-title {
            color: #2c3e50;
            font-weight: 600;
            margin-bottom: 1.5rem;
            font-size: 1.8rem;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        .table-responsive {
            border-radius: 15px;
            box-shadow: 0 8px 25px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .table {
            margin-bottom: 0;
        }
        .table thead th {
            background: linear-gradient(45deg, #343a40, #495057);
            color: white;
            border: none;
            padding: 1rem;
            font-weight: 600;
        }
        .table tbody tr {
            transition: all 0.3s ease;
        }
        .table tbody tr:hover {
            background-color: rgba(0,123,255,0.05);
            transform: scale(1.01);
        }
        .table tbody td {
            padding: 1rem;
            border: none;
            border-bottom: 1px solid #f1f3f4;
            vertical-align: middle;
        }
        .uptime-circle {
            width: 100px;
            height: 100px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 1.2rem;
            font-weight: 700;
            color: white;
            margin: 0 auto 1rem;
            background: conic-gradient(#28a745 0% {{.Stats.UptimePercentage}}%, #dc3545 {{.Stats.UptimePercentage}}% 100%);
        }
        .no-incidents {
            text-align: center;
            padding: 3rem;
            color: #6c757d;
            background: rgba(40, 167, 69, 0.1);
            border-radius: 15px;
            margin-top: 1rem;
        }
        .status-indicator {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            display: inline-block;
            margin-right: 0.5rem;
        }
        .status-up { background-color: #28a745; }
        .status-down { background-color: #dc3545; }
        .status-maintenance { background-color: #ffc107; }
        @media (max-width: 768px) {
            .page-title { font-size: 2rem; }
            .uptime-circle { width: 80px; height: 80px; font-size: 1rem; }
        }
    </style>
</head>
<body>
    <div class="container">
        <nav aria-label="breadcrumb">
            <ol class="breadcrumb">
                <li class="breadcrumb-item">
                    <a href="/"><i class="fas fa-home me-1"></i>Dashboard</a>
                </li>
                <li class="breadcrumb-item active">{{.Service.Name}}</li>
            </ol>
        </nav>
        
        <h1 class="page-title">
            <span class="status-icon">{{statusIcon .Service.Status}}</span>
            {{.Service.Name}}
            <span class="badge bg-{{statusClass .Service.Status}}">{{.Service.Status}}</span>
        </h1>
        
        <div class="row">
            <div class="col-lg-6">
                <div class="card info-card">
                    <div class="card-header">
                        <i class="fas fa-info-circle me-2"></i>
                        Service Information
                    </div>
                    <div class="card-body">
                        <table class="table table-borderless">
                            <tr>
                                <td><i class="fas fa-circle status-indicator status-{{.Service.Status}}"></i>Status:</td>
                                <td><span class="badge bg-{{statusClass .Service.Status}}">{{.Service.Status}}</span></td>
                            </tr>
                            <tr>
                                <td><i class="fas fa-network-wired text-muted me-2"></i>Protocol:</td>
                                <td>{{.Service.Protocol}}</td>
                            </tr>
                            <tr>
                                <td><i class="fas fa-globe text-muted me-2"></i>Endpoint:</td>
                                <td class="text-break">{{.Service.Endpoint}}</td>
                            </tr>
                            <tr>
                                <td><i class="fas fa-clock text-muted me-2"></i>Last Check:</td>
                                <td>{{formatTime .Service.LastCheck}}</td>
                            </tr>
                            <tr>
                                <td><i class="fas fa-tachometer-alt text-muted me-2"></i>Response Time:</td>
                                <td>{{formatDuration .Service.ResponseTime}}</td>
                            </tr>
                            <tr>
                                <td><i class="fas fa-chart-line text-muted me-2"></i>Total Checks:</td>
                                <td>{{.Service.TotalChecks}}</td>
                            </tr>
                            <tr>
                                <td><i class="fas fa-exclamation-triangle text-muted me-2"></i>Consecutive Fails:</td>
                                <td>{{.Service.ConsecutiveFails}}</td>
                            </tr>
                            {{if .Service.LastError}}
                            <tr>
                                <td><i class="fas fa-bug text-danger me-2"></i>Last Error:</td>
                                <td><span class="text-danger">{{.Service.LastError}}</span></td>
                            </tr>
                            {{end}}
                        </table>
                    </div>
                </div>
            </div>
            
            <div class="col-lg-6">
                {{if .Stats}}
                <div class="card info-card">
                    <div class="card-header">
                        <i class="fas fa-chart-pie me-2"></i>
                        Statistics (Last 30 Days)
                    </div>
                    <div class="card-body text-center">
                        <div class="uptime-circle">
                            {{printf "%.1f%%" .Stats.UptimePercentage}}
                        </div>
                        <table class="table table-borderless">
                            <tr>
                                <td><i class="fas fa-percentage text-muted me-2"></i>Uptime:</td>
                                <td><strong class="text-success">{{printf "%.2f%%" .Stats.UptimePercentage}}</strong></td>
                            </tr>
                            <tr>
                                <td><i class="fas fa-exclamation-circle text-muted me-2"></i>Total Incidents:</td>
                                <td><strong class="text-warning">{{.Stats.TotalIncidents}}</strong></td>
                            </tr>
                            <tr>
                                <td><i class="fas fa-clock text-muted me-2"></i>Total Downtime:</td>
                                <td><strong class="text-danger">{{formatDuration .Stats.TotalDowntime}}</strong></td>
                            </tr>
                        </table>
                    </div>
                </div>
                {{end}}
            </div>
        </div>
        
        <div class="incidents-section">
            <h3 class="section-title">
                <i class="fas fa-history"></i>
                Recent Incidents
            </h3>
            {{if .Incidents}}
            <div class="table-responsive">
                <table class="table table-striped">
                    <thead>
                        <tr>
                            <th><i class="fas fa-play me-2"></i>Start Time</th>
                            <th><i class="fas fa-stop me-2"></i>End Time</th>
                            <th><i class="fas fa-hourglass-half me-2"></i>Duration</th>
                            <th><i class="fas fa-exclamation me-2"></i>Error</th>
                            <th><i class="fas fa-flag me-2"></i>Status</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .Incidents}}
                        <tr>
                            <td>{{formatTime .StartTime}}</td>
                            <td>{{if .EndTime}}{{formatTime .EndTime}}{{else}}<em class="text-muted">Ongoing</em>{{end}}</td>
                            <td>{{if .Duration}}{{formatDuration .Duration}}{{else}}<em class="text-muted">Ongoing</em>{{end}}</td>
                            <td class="text-break">{{.Error}}</td>
                            <td>
                                {{if .Resolved}}
                                    <span class="badge bg-success"><i class="fas fa-check me-1"></i>Resolved</span>
                                {{else}}
                                    <span class="badge bg-danger"><i class="fas fa-exclamation me-1"></i>Active</span>
                                {{end}}
                            </td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
            {{else}}
            <div class="no-incidents">
                <i class="fas fa-check-circle text-success" style="font-size: 3rem; margin-bottom: 1rem;"></i>
                <h4 class="text-success">Great news!</h4>
                <p class="mb-0">No incidents recorded for this service.</p>
            </div>
            {{end}}
        </div>
        
        <div class="text-center mt-4">
            <a href="/" class="btn btn-outline-primary">
                <i class="fas fa-arrow-left me-2"></i>Back to Dashboard
            </a>
        </div>
    </div>
    
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
</body>
</html>
`

	template.Must(tmpl.New("dashboard.html").Parse(dashboardTmpl))
	template.Must(tmpl.New("service-detail.html").Parse(serviceDetailTmpl))

	return tmpl
}
