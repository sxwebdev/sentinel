package dispatcher

import (
	"context"

	"github.com/sxwebdev/sentinel/pkg/broker"
)

type Dispatcher struct {
	agents        *broker.Broker[struct{}]
	notifications *broker.Broker[struct{}]
}

func New() *Dispatcher {
	return &Dispatcher{
		agents:        broker.NewBroker[struct{}](),
		notifications: broker.NewBroker[struct{}](),
	}
}

func (s *Dispatcher) Name() string { return "dispatcher" }

func (s *Dispatcher) Start(_ context.Context) error {
	go s.agents.Start()
	go s.notifications.Start()
	return nil
}

func (s *Dispatcher) Stop(_ context.Context) error {
	s.agents.Stop()
	s.notifications.Stop()
	return nil
}

// Agents returns the agents broker
func (s *Dispatcher) Agents() *broker.Broker[struct{}] { return s.agents }

// Notifications returns the notifications broker
func (s *Dispatcher) Notifications() *broker.Broker[struct{}] { return s.notifications }
