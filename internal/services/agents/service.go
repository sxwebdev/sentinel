package agents

import (
	"github.com/sxwebdev/sentinel/internal/dispatcher"
	"github.com/sxwebdev/sentinel/internal/store"
)

type Service struct {
	store      *store.Store
	dispatcher *dispatcher.Dispatcher
}

func New(store *store.Store, dispatcher *dispatcher.Dispatcher) *Service {
	return &Service{
		store:      store,
		dispatcher: dispatcher,
	}
}
