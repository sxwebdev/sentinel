package receiver

import (
	"context"

	"github.com/sxwebdev/sentinel/internal/storage"
	"github.com/sxwebdev/sentinel/pkg/broker"
)

type TriggerServiceEventType int

const (
	TriggerServiceEventTypeCreated = iota
	TriggerServiceEventTypeUpdated
	TriggerServiceEventTypeDeleted
	TriggerServiceEventTypeCheck
)

type TriggerServiceData struct {
	EventType TriggerServiceEventType
	Svc       *storage.Service
}

func NewTriggerServiceData(
	eventType TriggerServiceEventType,
	svc *storage.Service,
) *TriggerServiceData {
	return &TriggerServiceData{
		EventType: eventType,
		Svc:       svc,
	}
}

type Receiver struct {
	serviceUpdated *broker.Broker[struct{}]
	triggerService *broker.Broker[TriggerServiceData]
}

func New() *Receiver {
	return &Receiver{
		serviceUpdated: broker.NewBroker[struct{}](),
		triggerService: broker.NewBroker[TriggerServiceData](),
	}
}

func (s Receiver) Name() string { return "receiver" }

func (s *Receiver) Start(_ context.Context) error {
	go s.serviceUpdated.Start()
	go s.triggerService.Start()

	return nil
}

func (s *Receiver) Stop(_ context.Context) error {
	s.serviceUpdated.Stop()
	s.triggerService.Stop()

	return nil
}

func (s *Receiver) ServiceUpdated() *broker.Broker[struct{}]           { return s.serviceUpdated }
func (s *Receiver) TriggerService() *broker.Broker[TriggerServiceData] { return s.triggerService }
