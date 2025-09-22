package repos

import (
	"database/sql"

	"github.com/sxwebdev/sentinel/internal/store/repos/repo_agents"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_incident_states"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_incidents"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_notification_history"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_notification_providers"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_service_states"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_services"
)

type Repos struct {
	agents                *repo_agents.CustomQueries
	services              *repo_services.CustomQueries
	serviceStates         *repo_service_states.CustomQueries
	incidents             *repo_incidents.CustomQueries
	incidentsStates       *repo_incident_states.Queries
	notificationProviders *repo_notification_providers.Queries
	notificationHistory   *repo_notification_history.CustomQueries
}

func New(sqlite *sql.DB) *Repos {
	return &Repos{
		agents:                repo_agents.NewCustom(sqlite),
		services:              repo_services.NewCustom(sqlite),
		serviceStates:         repo_service_states.NewCustom(sqlite),
		incidents:             repo_incidents.NewCustom(sqlite),
		incidentsStates:       repo_incident_states.New(sqlite),
		notificationProviders: repo_notification_providers.New(sqlite),
		notificationHistory:   repo_notification_history.NewCustom(sqlite),
	}
}

// Agents returns repo for agents
func (s *Repos) Agents(opts ...Option) repo_agents.ICustomQuerier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.agents.WithTx(options.Tx)
	}

	return s.agents
}

// Services returns repo for services
func (s *Repos) Services(opts ...Option) repo_services.ICustomQuerier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.services.WithTx(options.Tx)
	}

	return s.services
}

// ServiceStates returns repo for service states
func (s *Repos) ServiceStates(opts ...Option) repo_service_states.ICustomQuerier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.serviceStates.WithTx(options.Tx)
	}

	return s.serviceStates
}

// Incidents returns repo for incidents
func (s *Repos) Incidents(opts ...Option) repo_incidents.ICustomQuerier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.incidents.WithTx(options.Tx)
	}

	return s.incidents
}

// IncidentStates returns repo for incident states
func (s *Repos) IncidentStates(opts ...Option) repo_incident_states.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.incidentsStates.WithTx(options.Tx)
	}

	return s.incidentsStates
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
