package service

import (
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/store"
)

type Service struct {
	store    *store.Store
	receiver *receiver.Receiver
}

func New(store *store.Store, receiver *receiver.Receiver) *Service {
	return &Service{
		store:    store,
		receiver: receiver,
	}
}
