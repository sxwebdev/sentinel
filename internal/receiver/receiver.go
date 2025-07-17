package receiver

import (
	"context"

	"github.com/sxwebdev/sentinel/internal/storage"
	"github.com/sxwebdev/sentinel/pkg/broker"
)

type TriggerServiceEventType int

const (
	TriggerServiceEventTypeUnknown = iota
	TriggerServiceEventTypeCreated
	TriggerServiceEventTypeUpdated
	TriggerServiceEventTypeDeleted
	TriggerServiceEventTypeCheck
	TriggerServiceEventTypeUpdatedState
)

// String returns a string representation of the TriggerServiceEventType
func (e TriggerServiceEventType) String() string {
	switch e {
	case TriggerServiceEventTypeCreated:
		return "created"
	case TriggerServiceEventTypeUpdated:
		return "updated"
	case TriggerServiceEventTypeDeleted:
		return "deleted"
	case TriggerServiceEventTypeCheck:
		return "check"
	case TriggerServiceEventTypeUpdatedState:
		return "updated_state"
	default:
		return "unknown"
	}
}

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
	triggerService *broker.Broker[TriggerServiceData]
}

func New() *Receiver {
	return &Receiver{
		triggerService: broker.NewBroker[TriggerServiceData](),
	}
}

func (s Receiver) Name() string { return "receiver" }

func (s *Receiver) Start(_ context.Context) error {
	go s.triggerService.Start()
	return nil
}

func (s *Receiver) Stop(_ context.Context) error {
	s.triggerService.Stop()
	return nil
}

func (s *Receiver) TriggerService() *broker.Broker[TriggerServiceData] { return s.triggerService }
