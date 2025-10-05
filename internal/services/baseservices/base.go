package baseservices

import (
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/dispatcher"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/services/agents"
	"github.com/sxwebdev/sentinel/internal/services/auth"
	"github.com/sxwebdev/sentinel/internal/services/incidents"
	"github.com/sxwebdev/sentinel/internal/services/notifications"
	"github.com/sxwebdev/sentinel/internal/services/projects"
	"github.com/sxwebdev/sentinel/internal/services/service"
	"github.com/sxwebdev/sentinel/internal/services/servicestate"
	"github.com/sxwebdev/sentinel/internal/services/system"
	"github.com/sxwebdev/sentinel/internal/services/users"
	"github.com/sxwebdev/sentinel/internal/store"
	"github.com/tkcrm/mx/logger"
)

type BaseServices struct {
	dispatcher *dispatcher.Dispatcher

	usersService         *users.Service
	systemService        *system.Service
	authService          *auth.Service[*models.User]
	projectsService      *projects.Service
	agentsService        *agents.Service
	servicesService      *service.Service
	serviceStateService  *servicestate.Service
	incidentsService     *incidents.Service
	notificationsService *notifications.Service
}

func New(
	l logger.Logger,
	st *store.Store,
	authConfig config.AuthConfig,
	receiver *receiver.Receiver,
	dispatcher *dispatcher.Dispatcher,
) *BaseServices {
	usersService := users.New(st)

	systemService := system.New(st, usersService)

	authService := auth.New(
		l,
		st.TokenRepo(),
		usersService,
		authConfig.AccessTokenSecretKey,
		time.Minute*15,
		authConfig.RefreshTokenSecretKey,
		time.Hour*24*30, // 30 days
	)

	projectsService := projects.New(st)
	agentsService := agents.New(st, dispatcher)
	servicesService := service.New(st, receiver)
	serviceStateService := servicestate.New(st)
	incidentsService := incidents.New(st)
	notificationsService := notifications.New(l, st, dispatcher)

	return &BaseServices{
		dispatcher:           dispatcher,
		usersService:         usersService,
		systemService:        systemService,
		authService:          authService,
		projectsService:      projectsService,
		agentsService:        agentsService,
		servicesService:      servicesService,
		serviceStateService:  serviceStateService,
		incidentsService:     incidentsService,
		notificationsService: notificationsService,
	}
}

// Dispatcher returns dispatcher
func (b *BaseServices) Dispatcher() *dispatcher.Dispatcher {
	return b.dispatcher
}

// Users returns users service
func (b *BaseServices) Users() *users.Service {
	return b.usersService
}

// System returns system service
func (b *BaseServices) System() *system.Service {
	return b.systemService
}

// Auth returns auth service
func (b *BaseServices) Auth() *auth.Service[*models.User] {
	return b.authService
}

// Projects returns projects service
func (b *BaseServices) Projects() *projects.Service {
	return b.projectsService
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
