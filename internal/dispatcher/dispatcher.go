package dispatcher

import (
	"context"

	"github.com/sxwebdev/sentinel/pkg/broker"
)

type Dispatcher struct {
	agents                *broker.Broker[BaseMessage]
	notificationProviders *broker.Broker[BaseMessage]
	notificationHistory   *broker.Broker[BaseMessage]
}

func New() *Dispatcher {
	return &Dispatcher{
		agents:                broker.NewBroker[BaseMessage](),
		notificationProviders: broker.NewBroker[BaseMessage](),
		notificationHistory:   broker.NewBroker[BaseMessage](),
	}
}

func (s *Dispatcher) Name() string { return "dispatcher" }

func (s *Dispatcher) Start(_ context.Context) error {
	go s.agents.Start()
	go s.notificationProviders.Start()
	go s.notificationHistory.Start()
	return nil
}

func (s *Dispatcher) Stop(_ context.Context) error {
	s.agents.Stop()
	s.notificationProviders.Stop()
	s.notificationHistory.Stop()
	return nil
}

// Agents returns the agents broker
func (s *Dispatcher) Agents() *broker.Broker[BaseMessage] { return s.agents }

// NotificationProviders returns the notification providers broker
func (s *Dispatcher) NotificationProviders() *broker.Broker[BaseMessage] {
	return s.notificationProviders
}

// NotificationHistory returns the notification history broker
func (s *Dispatcher) NotificationHistory() *broker.Broker[BaseMessage] { return s.notificationHistory }
