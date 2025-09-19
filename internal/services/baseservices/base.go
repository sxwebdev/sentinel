package baseservices

import (
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/services/agents"
	"github.com/sxwebdev/sentinel/internal/services/incidents"
	"github.com/sxwebdev/sentinel/internal/services/service"
	"github.com/sxwebdev/sentinel/internal/store"
)

type BaseServices struct {
	agentsService    *agents.Service
	servicesService  *service.Service
	incidentsService *incidents.Service
}

func New(
	st *store.Store,
	receiver *receiver.Receiver,
) *BaseServices {
	agentsService := agents.New(st)
	servicesService := service.New(st, receiver)
	incidentsService := incidents.New(st)

	return &BaseServices{
		agentsService:    agentsService,
		servicesService:  servicesService,
		incidentsService: incidentsService,
	}
}

// Agents returns agents service
func (b *BaseServices) Agents() *agents.Service {
	return b.agentsService
}

// Services returns services service
func (b *BaseServices) Services() *service.Service {
	return b.servicesService
}

// Incidents returns incidents service
func (b *BaseServices) Incidents() *incidents.Service {
	return b.incidentsService
}
