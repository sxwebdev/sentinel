package notifications

import (
	"github.com/sxwebdev/sentinel/internal/store"
	"github.com/tkcrm/mx/logger"
)

type Service struct {
	providers *Providers
	history   *History
	sender    *Sender
}

func New(l logger.Logger, store *store.Store) *Service {
	senderSvc := newSender(l, store)
	historySvc := newHistory(l, store, senderSvc)

	return &Service{
		providers: newProviders(store, historySvc),
		history:   historySvc,
		sender:    senderSvc,
	}
}

// Providers returns notification providers service
func (s *Service) Providers() *Providers {
	return s.providers
}

// History returns notifications history service
func (s *Service) History() *History {
	return s.history
}

// Sender returns notifications sender service
func (s *Service) Sender() *Sender {
	return s.sender
}
