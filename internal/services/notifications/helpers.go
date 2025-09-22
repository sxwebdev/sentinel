package notifications

import (
	"fmt"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

// validateProviderConfig validates the configuration of a notification provider based on its type.
func validateProviderConfig(providerType models.NotificationProviderType, config storecmn.JSONField) error {
	switch providerType {
	case models.NotificationProviderTypeShoutrrr:
		var c models.NotificationProviderShoutrrrConfig
		if err := config.ConvertToAny(&c); err != nil {
			return err
		}
		return c.Validate()
	default:
		return fmt.Errorf("unsupported notification provider type: %s", providerType)
	}
}
