package incidents

import (
	"github.com/sxwebdev/sentinel/internal/store"
)

type Service struct {
	store *store.Store
}

func New(store *store.Store) *Service {
	return &Service{
		store: store,
	}
}
