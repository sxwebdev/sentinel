package baseservices

import (
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/services/agents"
	"github.com/sxwebdev/sentinel/internal/services/incidents"
	"github.com/sxwebdev/sentinel/internal/services/notifications"
	"github.com/sxwebdev/sentinel/internal/services/service"
	"github.com/sxwebdev/sentinel/internal/services/servicestate"
	"github.com/sxwebdev/sentinel/internal/store"
	"github.com/tkcrm/mx/logger"
)

type BaseServices struct {
	agentsService        *agents.Service
	servicesService      *service.Service
	serviceStateService  *servicestate.Service
	incidentsService     *incidents.Service
	notificationsService *notifications.Service
}

func New(
	l logger.Logger,
	st *store.Store,
	receiver *receiver.Receiver,
) *BaseServices {
	agentsService := agents.New(st)
	servicesService := service.New(st, receiver)
	serviceStateService := servicestate.New(st)
	incidentsService := incidents.New(st)
	notificationsService := notifications.New(l, st)

	return &BaseServices{
		agentsService:        agentsService,
		servicesService:      servicesService,
		serviceStateService:  serviceStateService,
		incidentsService:     incidentsService,
		notificationsService: notificationsService,
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

// ServiceStates returns service states service
func (b *BaseServices) ServiceStates() *servicestate.Service {
	return b.serviceStateService
}

// Incidents returns incidents service
func (b *BaseServices) Incidents() *incidents.Service {
	return b.incidentsService
}

// Notifications returns notifications service
func (b *BaseServices) Notifications() *notifications.Service {
	return b.notificationsService
}
