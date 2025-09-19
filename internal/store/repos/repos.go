package repos

import (
	"database/sql"

	"github.com/sxwebdev/sentinel/internal/store/repos/repo_agents"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_incidents"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_service_states"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_services"
)

type Repos struct {
	agents        *repo_agents.CustomQueries
	services      *repo_services.CustomQueries
	serviceStates *repo_service_states.Queries
	incidents     *repo_incidents.Queries
}

func New(sqlite *sql.DB) *Repos {
	return &Repos{
		agents:        repo_agents.NewCustom(sqlite),
		services:      repo_services.NewCustom(sqlite),
		serviceStates: repo_service_states.New(sqlite),
		incidents:     repo_incidents.New(sqlite),
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
func (s *Repos) ServiceStates(opts ...Option) repo_service_states.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.serviceStates.WithTx(options.Tx)
	}

	return s.serviceStates
}

// Incidents returns repo for incidents
func (s *Repos) Incidents(opts ...Option) repo_incidents.Querier {
	options := parseOptions(opts...)

	if options.Tx != nil {
		return s.incidents.WithTx(options.Tx)
	}

	return s.incidents
}
