package models

import (
	"errors"
	"fmt"

	"github.com/nicholas-fedor/shoutrrr"
)

type NotificationProviderType string

const (
	NotificationProviderTypeShoutrrr NotificationProviderType = "shoutrrr"
)

// Validate checks the validity of the NotificationProvider fields.
func (n NotificationProviderType) Validate() error {
	if n == "" {
		return fmt.Errorf("empty notification provider type")
	}

	switch n {
	case NotificationProviderTypeShoutrrr:
		return nil
	default:
		return fmt.Errorf("invalid notification provider type: %s", n)
	}
}

type NotificationProviderShoutrrrConfig struct {
	URL string `json:"url"`
}

// Validate checks the validity of the NotificationProviderShoutrrrConfig fields.
func (n NotificationProviderShoutrrrConfig) Validate() error {
	if n.URL == "" {
		return fmt.Errorf("empty shoutrrr url")
	}

	_, err := shoutrrr.CreateSender(n.URL)
	if err != nil {
		return errors.New("invalid shoutrrr url")
	}

	return nil
}

type NotificationHistoryView struct {
	NotificationHistory
	MonitorName *string `json:"monitor_name"`
}
