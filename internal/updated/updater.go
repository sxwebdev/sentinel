package updater

import (
	"context"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/pkg/locker"
	"github.com/tkcrm/mx/logger"
)

type Updater struct {
	logger logger.Logger
	config config.Updater

	version             string
	availableUpdateData *locker.Locker[models.AvailableUpdate]
}

func New(
	l logger.Logger,
	cfg config.Updater,
	version string,
	availableUpdateData *locker.Locker[models.AvailableUpdate],
) (*Updater, error) {
	return &Updater{
		logger:              l,
		config:              cfg,
		version:             version,
		availableUpdateData: availableUpdateData,
	}, nil
}

// Name returns the name of the updater
func (u *Updater) Name() string { return "updater" }

// Start starts the updater
func (u *Updater) Start(ctx context.Context) error {
	if !u.config.IsEnabled {
		u.logger.Info("updater is disabled")
		return nil
	}

	go u.checkNewVersionWrapper(ctx)

	return nil
}

// Stop stops the updater
func (u *Updater) Stop(_ context.Context) error {
	return nil
}
