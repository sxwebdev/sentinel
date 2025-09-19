package service

import (
	"context"
	"fmt"

	"github.com/sxwebdev/sentinel/internal/receiver"
)

// TriggerCheck triggers a manual check for a service
func (s *Service) TriggerCheck(ctx context.Context, id string) error {
	svc, err := s.GetViewByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	s.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeCheck,
		svc,
	))

	return nil
}
