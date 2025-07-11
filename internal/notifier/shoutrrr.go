package notifier

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/containrrr/shoutrrr"
	"github.com/sxwebdev/sentinel/internal/storage"
)

// Notifier определяет интерфейс для отправки уведомлений
type Notifier interface {
	SendAlert(service *storage.Service, incident *storage.Incident) error
	SendRecovery(sservice *storage.Service, incident *storage.Incident) error
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
func (s *NotifierImpl) SendAlert(service *storage.Service, incident *storage.Incident) error {
	message := s.formatAlertMessage(service, incident)
	return s.sendMessage(message)
}

// SendRecovery sends a recovery notification when a service comes back up
func (s *NotifierImpl) SendRecovery(service *storage.Service, incident *storage.Incident) error {
	message := s.formatRecoveryMessage(service, incident)
	return s.sendMessage(message)
}

// sendMessage sends a message to all configured providers
// If one provider fails, it continues with others and returns partial errors
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
				errors <- fmt.Errorf("provider %d: failed to create sender: %w", index, err)
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
					errors <- fmt.Errorf("provider %d: failed to send: %v", index, errs)
				}
			case <-time.After(30 * time.Second):
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

	// If some providers failed, return partial error
	if len(allErrors) > 0 {
		return fmt.Errorf("some notification providers failed (%d/%d): %v", len(allErrors), len(s.urls), allErrors)
	}

	return nil
}

// formatAlertMessage formats an alert message
func (s *NotifierImpl) formatAlertMessage(service *storage.Service, incident *storage.Incident) string {
	tags := "-"
	if len(service.Tags) > 0 {
		tags = strings.Join(service.Tags, ", ")
	}

	return fmt.Sprintf(
		"🔴 [ALERT] %s is DOWN\n\n"+
			"• Service: %s\n"+
			"• Tags: %s\n"+
			"• Error: %s\n"+
			"• Started: %s\n"+
			"• Incident ID: %s",
		service.Name,
		service.Name,
		tags,
		incident.Error,
		incident.StartTime.Format("2006-01-02 15:04:05"),
		incident.ID,
	)
}

// formatRecoveryMessage formats a recovery message
func (s *NotifierImpl) formatRecoveryMessage(service *storage.Service, incident *storage.Incident) string {
	var duration string
	if incident.Duration != nil {
		duration = formatDuration(*incident.Duration)
	} else {
		duration = formatDuration(time.Since(incident.StartTime))
	}

	var endTime string
	if incident.EndTime != nil {
		endTime = incident.EndTime.Format("2006-01-02 15:04:05")
	} else {
		endTime = time.Now().Format("2006-01-02 15:04:05")
	}

	tags := "-"
	if len(service.Tags) > 0 {
		tags = strings.Join(service.Tags, ", ")
	}

	return fmt.Sprintf(
		"🟢 [RECOVERY] %s is UP\n\n"+
			"• Service: %s\n"+
			"• Tags: %s\n"+
			"• Downtime: %s\n"+
			"• Recovered: %s\n"+
			"• Incident ID: %s",
		service.Name,
		service.Name,
		tags,
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
