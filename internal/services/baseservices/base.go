package baseservices

import (
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/services/service"
	"github.com/sxwebdev/sentinel/internal/store"
)

type BaseServices struct {
	services *service.Service
}

func New(
	st *store.Store,
	receiver *receiver.Receiver,
) *BaseServices {
	servicesService := service.New(st, receiver)

	return &BaseServices{
		services: servicesService,
	}
}

// Services returns services service
func (b *BaseServices) Services() *service.Service {
	return b.services
}
