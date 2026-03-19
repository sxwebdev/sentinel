package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/nicholas-fedor/shoutrrr"
	"github.com/sxwebdev/sentinel/internal/dispatcher"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_notification_history"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/sxwebdev/xutils/loopper"
	"github.com/tkcrm/mx/logger"
)

type Sender struct {
	logger logger.Logger
	store  *store.Store
	looper *loopper.Loopper

	dispatcher *dispatcher.Dispatcher
}

func newSender(l logger.Logger, store *store.Store, dispatcher *dispatcher.Dispatcher) *Sender {
	s := &Sender{
		logger:     l,
		store:      store,
		dispatcher: dispatcher,
	}

	s.looper = loopper.New(
		s.initSender,
		loopper.WithLeading(),
		loopper.WithPeriod(time.Second*5),
		loopper.WithContextTimeout(time.Second*30),
	)

	return s
}

// Name
func (s *Sender) Name() string {
	return "notifications_sender"
}

// Start
func (s *Sender) Start(ctx context.Context) error {
	s.looper.Start(ctx)
	return nil
}

// Stop
func (s *Sender) Stop(_ context.Context) error {
	s.looper.Stop()
	s.looper.Wait()
	return nil
}

// initSender
func (s *Sender) initSender(ctx context.Context) {
	if err := s.do(ctx); err != nil {
		s.logger.Errorf("failed to process notification history: %v", err)
	}
}

func (s *Sender) do(ctx context.Context) error {
	items, err := s.store.NotificationHistory().GetAllUnsent(ctx)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	s.logger.Infof("found %d unsent notifications", len(items))

	unsentAlerts := make(map[string]struct{})

	for _, item := range items {
		if item.AlertID != nil && *item.AlertID != "" {
			if _, exists := unsentAlerts[*item.AlertID]; exists {
				s.logger.Infof("skipping notification %s for alert %s as previous attempt failed", item.ID, *item.AlertID)
				continue
			}
		}

		var err error
		switch item.ProviderType {
		case models.NotificationProviderTypeShoutrrr:
			err = s.processShoutrrr(ctx, item)
		default:
			s.logger.Warnf("unsupported notification provider type: %s", item.ProviderType)
			continue
		}

		if err != nil {
			incrementErr := s.store.NotificationHistory().IncrementAttempt(ctx, repo_notification_history.IncrementAttemptParams{
				ID:           item.ID,
				ErrorMessage: utils.Pointer(err.Error()),
			})
			if incrementErr != nil {
				return incrementErr
			}

			if item.AlertID != nil && *item.AlertID != "" {
				unsentAlerts[*item.AlertID] = struct{}{}
			}

			s.logger.Errorf("failed to send notification %s: %v", item.ID, err)
		} else {
			err := s.store.NotificationHistory().MarkAsSent(ctx, nil, item.ID)
			if err != nil {
				return err
			}

			s.logger.Infof("notification %s sent successfully", item.ID)
		}

		if item.ProjectID != nil && *item.ProjectID != "" {
			s.dispatcher.NotificationHistory().Publish(dispatcher.NewBaseMessage(dispatcher.EventTypeUpdate, *item.ProjectID))
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

// processShoutrrr
func (s *Sender) processShoutrrr(ctx context.Context, item *repo_notification_history.GetAllUnsentRow) error {
	var config models.NotificationProviderShoutrrrConfig
	if err := item.Config.ConvertToAny(&config); err != nil {
		return err
	}

	if config.URL == "" {
		return fmt.Errorf("empty shoutrrr URL in config for provider ID %s", item.ProviderID)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- shoutrrr.Send(config.URL, item.Message)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
