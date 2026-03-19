package agents

import (
	"github.com/sxwebdev/sentinel/internal/dispatcher"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store"
)

type Service struct {
	store      *store.Store
	dispatcher *dispatcher.Dispatcher

	systemInfo *models.SystemInfo
}

func New(store *store.Store, dispatcher *dispatcher.Dispatcher, systemInfo *models.SystemInfo) *Service {
	return &Service{
		store:      store,
		dispatcher: dispatcher,
		systemInfo: systemInfo,
	}
}
