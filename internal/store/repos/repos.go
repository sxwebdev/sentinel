package repos

import (
	"database/sql"

	"github.com/sxwebdev/sentinel/internal/store/repos/repo_agents"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_alert_policies"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_alerts"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_incident_events"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_incidents"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_maintenance_windows"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_monitor_agents"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_monitor_checks"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_monitor_revisions"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_monitor_states"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_monitors"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_notification_history"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_notification_providers"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_projects"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_resource_agents"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_resources"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_users"
)

type Repos struct {
	users *repo_users.Queries

	projects *repo_projects.Queries

	agents *repo_agents.CustomQueries

	resources       *repo_resources.Queries
	resourcesAgents *repo_resource_agents.Queries

	monitors         *repo_monitors.Queries
	monitorRevisions *repo_monitor_revisions.Queries
	monitorStates    *repo_monitor_states.Queries
	monitorChecks    *repo_monitor_checks.Queries
	monitorAgents    *repo_monitor_agents.Queries

	incidents      *repo_incidents.CustomQueries
	incidentEvents *repo_incident_events.Queries

	notificationProviders *repo_notification_providers.Queries
	notificationHistory   *repo_notification_history.CustomQueries

	alertPolicies *repo_alert_policies.Queries
	alerts        *repo_alerts.Queries

	maintenanceWindows *repo_maintenance_windows.Queries
}

func New(sqlite *sql.DB) *Repos {
	return &Repos{
		users:                 repo_users.New(sqlite),
		projects:              repo_projects.New(sqlite),
		agents:                repo_agents.NewCustom(sqlite),
		resources:             repo_resources.New(sqlite),
		resourcesAgents:       repo_resource_agents.New(sqlite),
		monitors:              repo_monitors.New(sqlite),
		monitorRevisions:      repo_monitor_revisions.New(sqlite),
		monitorStates:         repo_monitor_states.New(sqlite),
		monitorChecks:         repo_monitor_checks.New(sqlite),
		monitorAgents:         repo_monitor_agents.New(sqlite),
		incidents:             repo_incidents.NewCustom(sqlite),
		incidentEvents:        repo_incident_events.New(sqlite),
		notificationProviders: repo_notification_providers.New(sqlite),
		notificationHistory:   repo_notification_history.NewCustom(sqlite),
		alertPolicies:         repo_alert_policies.New(sqlite),
		alerts:                repo_alerts.New(sqlite),
		maintenanceWindows:    repo_maintenance_windows.New(sqlite),
	}
}

// Users returns repo for users
func (s *Repos) Users(opts ...Option) repo_users.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.users.WithTx(options.Tx)
	}

	return s.users
}

// Projects returns repo for projects
func (s *Repos) Projects(opts ...Option) repo_projects.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.projects.WithTx(options.Tx)
	}

	return s.projects
}

// Agents returns repo for agents
func (s *Repos) Agents(opts ...Option) repo_agents.ICustomQuerier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.agents.WithTx(options.Tx)
	}

	return s.agents
}

// Resources returns repo for resources
func (s *Repos) Resources(opts ...Option) repo_resources.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.resources.WithTx(options.Tx)
	}

	return s.resources
}

// ResourcesAgents returns repo for resources agents
func (s *Repos) ResourcesAgents(opts ...Option) repo_resource_agents.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.resourcesAgents.WithTx(options.Tx)
	}

	return s.resourcesAgents
}

// Monitors returns repo for monitors
func (s *Repos) Monitors(opts ...Option) repo_monitors.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.monitors.WithTx(options.Tx)
	}

	return s.monitors
}

// MonitorRevisions returns repo for monitor revisions
func (s *Repos) MonitorRevisions(opts ...Option) repo_monitor_revisions.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.monitorRevisions.WithTx(options.Tx)
	}

	return s.monitorRevisions
}

// MonitorStates returns repo for monitor states
func (s *Repos) MonitorStates(opts ...Option) repo_monitor_states.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.monitorStates.WithTx(options.Tx)
	}

	return s.monitorStates
}

// MonitorChecks returns repo for monitor checks
func (s *Repos) MonitorChecks(opts ...Option) repo_monitor_checks.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.monitorChecks.WithTx(options.Tx)
	}

	return s.monitorChecks
}

// MonitorAgents returns repo for monitor agents
func (s *Repos) MonitorAgents(opts ...Option) repo_monitor_agents.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.monitorAgents.WithTx(options.Tx)
	}

	return s.monitorAgents
}

// Incidents returns repo for incidents
func (s *Repos) Incidents(opts ...Option) repo_incidents.ICustomQuerier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.incidents.WithTx(options.Tx)
	}

	return s.incidents
}

// IncidentEvents returns repo for incident events
func (s *Repos) IncidentEvents(opts ...Option) repo_incident_events.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.incidentEvents.WithTx(options.Tx)
	}

	return s.incidentEvents
}

// NotificationProviders returns repo for notification providers
func (s *Repos) NotificationProviders(opts ...Option) repo_notification_providers.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.notificationProviders.WithTx(options.Tx)
	}

	return s.notificationProviders
}

// NotificationHistory returns repo for notification history
func (s *Repos) NotificationHistory(opts ...Option) repo_notification_history.ICustomQuerier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.notificationHistory.WithTx(options.Tx)
	}

	return s.notificationHistory
}

// AlertPolicies returns repo for alert policies
func (s *Repos) AlertPolicies(opts ...Option) repo_alert_policies.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.alertPolicies.WithTx(options.Tx)
	}

	return s.alertPolicies
}

// Alerts returns repo for alerts
func (s *Repos) Alerts(opts ...Option) repo_alerts.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.alerts.WithTx(options.Tx)
	}

	return s.alerts
}

// MaintenanceWindows returns repo for maintenance windows
func (s *Repos) MaintenanceWindows(opts ...Option) repo_maintenance_windows.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.maintenanceWindows.WithTx(options.Tx)
	}

	return s.maintenanceWindows
}
