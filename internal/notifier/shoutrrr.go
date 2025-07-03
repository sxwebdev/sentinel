package notifier

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/containrrr/shoutrrr"
	"github.com/sxwebdev/sentinel/internal/config"
)

// Notifier определяет интерфейс для отправки уведомлений
type Notifier interface {
	SendAlert(serviceName string, incident *config.Incident) error
	SendRecovery(serviceName string, incident *config.Incident) error
}

var _ Notifier = (*NotifierImpl)(nil)

type NotifierImpl struct {
	urls []string
}

// NewNotifier создает новый экземпляр Notifier
func NewNotifier(urls []string) (Notifier, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no notification URLs provided")
	}
	return &NotifierImpl{urls: urls}, nil
}

// SendAlert sends an alert notification when a service goes down
func (s *NotifierImpl) SendAlert(serviceName string, incident *config.Incident) error {
	message := s.formatAlertMessage(serviceName, incident)
	return s.sendMessage(message)
}

// SendRecovery sends a recovery notification when a service comes back up
func (s *NotifierImpl) SendRecovery(serviceName string, incident *config.Incident) error {
	message := s.formatRecoveryMessage(serviceName, incident)
	return s.sendMessage(message)
}

// sendMessage sends a message to all configured providers
// If one provider fails, it continues with others and logs the error
func (s *NotifierImpl) sendMessage(message string) error {
	// Send to all providers concurrently
	var wg sync.WaitGroup
	errors := make(chan error, len(s.urls))

	for i, url := range s.urls {
		wg.Add(1)
		go func(index int, providerURL string) {
			defer wg.Done()

			// Create individual sender for this provider
			sender, err := shoutrrr.CreateSender(providerURL)
			if err != nil {
				log.Printf("Failed to create sender for provider %d: %v", index, err)
				errors <- fmt.Errorf("provider %d: %w", index, err)
				return
			}

			// Send message with timeout
			done := make(chan []error, 1)
			go func() {
				done <- sender.Send(message, nil)
			}()

			select {
			case errs := <-done:
				// Check if there are any actual errors (not nil)
				hasErrors := false
				for _, err := range errs {
					if err != nil {
						hasErrors = true
						break
					}
				}

				if hasErrors {
					log.Printf("Failed to send notification to provider %d: %v", index, errs)
					errors <- fmt.Errorf("provider %d: %v", index, errs)
				} else {
					log.Printf("Successfully sent notification to provider %d", index)
				}
			case <-time.After(30 * time.Second):
				log.Printf("Timeout sending notification to provider %d", index)
				errors <- fmt.Errorf("provider %d: timeout", index)
			}
		}(i, url)
	}

	wg.Wait()
	close(errors)

	// Collect all errors
	var allErrors []error
	for err := range errors {
		allErrors = append(allErrors, err)
	}

	// If all providers failed, return an error
	if len(allErrors) == len(s.urls) {
		return fmt.Errorf("all notification providers failed: %v", allErrors)
	}

	// If some providers failed, log but don't return error
	if len(allErrors) > 0 {
		log.Printf("Some notification providers failed (%d/%d), but others succeeded", len(allErrors), len(s.urls))
	}

	return nil
}

// formatAlertMessage formats an alert message
func (s *NotifierImpl) formatAlertMessage(serviceName string, incident *config.Incident) string {
	return fmt.Sprintf(
		"🔴 [ALERT] %s is DOWN\n\n"+
			"• Service: %s\n"+
			"• Error: %s\n"+
			"• Started: %s\n"+
			"• Incident ID: %s",
		serviceName,
		serviceName,
		incident.Error,
		incident.StartTime.Format("2006-01-02 15:04:05 UTC"),
		incident.ID,
	)
}

// formatRecoveryMessage formats a recovery message
func (s *NotifierImpl) formatRecoveryMessage(serviceName string, incident *config.Incident) string {
	var duration string
	if incident.Duration != nil {
		duration = formatDuration(*incident.Duration)
	} else {
		duration = formatDuration(time.Since(incident.StartTime))
	}

	var endTime string
	if incident.EndTime != nil {
		endTime = incident.EndTime.Format("2006-01-02 15:04:05 UTC")
	} else {
		endTime = time.Now().Format("2006-01-02 15:04:05 UTC")
	}

	return fmt.Sprintf(
		"🟢 [RECOVERY] %s is UP\n\n"+
			"• Service: %s\n"+
			"• Downtime: %s\n"+
			"• Recovered: %s\n"+
			"• Incident ID: %s",
		serviceName,
		serviceName,
		duration,
		endTime,
		incident.ID,
	)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
